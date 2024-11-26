package songs

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/filter"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
)

type Repo struct {
	log    *logger.Logger
	conn   *pgx.Conn
	filter *filter.Filter
}

func newRepo(log *logger.Logger, conn *pgx.Conn) *Repo {
	return &Repo{
		log:  log,
		conn: conn,
		filter: filter.New(
			"song",
			map[string]filter.ColumnConfig{
				"id": {
					Name: "id",
					Type: filter.NumberType,
				},
				"song": {
					Name: "title",
					Type: filter.StringType,
				},
				"group": {
					Name: "artist",
					Type: filter.StringType,
				},
				"releaseDate": {
					Name: "release_date",
					Type: filter.DateType,
				},
				"text": {
					Name: "lyrics",
					Type: filter.ArrayOf(filter.StringType),
				},
				"link": {
					Name: "link",
					Type: filter.StringType,
				},
			},
			func(s string) (any, error) {
				d, err := time.Parse(releaseDateFormat, s)
				if err != nil {
					return nil, err
				}
				return pgtype.Date{Time: d, Valid: true}, nil
			},
		),
	}
}

const saveSongQuery = `INSERT INTO song (title, artist, release_date, lyrics, link) VALUES ($1, $2, $3, $4, $5) RETURNING id;`

func (s *Repo) SaveSong(ctx context.Context, song *Song) error {
	args := []any{song.Title, song.Artist, pgtype.Date{Time: song.ReleaseDate, Valid: true}, song.Lyrics, song.Link}
	s.log.Debug(ctx, "executing query", slog.String("query", saveSongQuery), slog.Any("args", args))
	row := s.conn.QueryRow(ctx, saveSongQuery, args...)
	return row.Scan(&song.ID)
}

func (s *Repo) GetSongs(ctx context.Context, query Query) ([]Song, error) {
	q := strings.Builder{}
	q.Grow(100)
	q.WriteString(`SELECT id, title, artist, release_date, lyrics, link FROM song`)
	var args []any
	if query.LastId != 0 {
		q.WriteString(" WHERE id > $1")
		args = append(args, query.LastId)
	}
	if query.Filter != "" {
		expr, err := s.filter.Parse(query.Filter)
		if err != nil {
			return nil, err
		}
		if query.LastId == 0 {
			q.WriteString(" WHERE ")
		} else {
			q.WriteString(" AND ")
		}
		q.Grow(len(query.Filter) * 2)
		args = expr.ToSQL(&q, args)
	}
	q.WriteString(" ORDER BY id ASC")
	if query.Page > 0 {
		q.WriteString(" OFFSET $")
		args = append(args, (query.Page-1)*query.PageSize)
		q.WriteString(strconv.Itoa(len(args)))
	}
	q.WriteString(" LIMIT $")
	args = append(args, query.PageSize)
	q.WriteString(strconv.Itoa(len(args)))
	sql := q.String()
	s.log.Debug(ctx, "executing query", slog.String("query", sql), slog.Any("args", args))
	rows, err := s.conn.Query(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var songs []Song
	for rows.Next() {
		var s Song
		var d pgtype.Date
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &d, &s.Lyrics, &s.Link); err != nil {
			return nil, err
		}
		s.ReleaseDate = d.Time.In(time.Local)
		songs = append(songs, s)
	}
	s.log.Debug(ctx, "got songs", slog.Int("count", len(songs)))
	return songs, nil
}

const lyricsQuery = `SELECT lyrics[$1:$2] AS paginated FROM song WHERE id = $3`

func (s *Repo) GetLyrics(ctx context.Context, id int64, pagination Pagination) ([]string, error) {
	args := []any{pagination.Page, pagination.Page + pagination.PageSize - 1, id}
	s.log.Debug(ctx, "executing query", slog.String("query", lyricsQuery), slog.Any("args", args))
	row := s.conn.QueryRow(ctx, lyricsQuery, args...)
	var lyrics []string
	err := row.Scan(&lyrics)
	s.log.Debug(ctx, "got lyrics", slog.Int("count", len(lyrics)))
	return lyrics, err
}

const deleteSongQuery = `DELETE FROM song WHERE id = $1`

func (s *Repo) DeleteSong(ctx context.Context, id int64) error {
	s.log.Debug(ctx, "executing query", slog.String("query", deleteSongQuery), slog.Int64("id", id))
	_, err := s.conn.Exec(ctx, deleteSongQuery, id)
	return err
}

var songFieldToColumn = map[SongField]string{
	Title:       "title",
	Artist:      "artist",
	ReleaseDate: "release_date",
	Lyrics:      "lyrics",
	Link:        "link",
}

func (s *Repo) UpdateSong(ctx context.Context, id int64, upd SongUpdate) error {
	q := strings.Builder{}
	q.Grow(100)
	q.WriteString("UPDATE song SET ")
	i := 0
	var args []any
	for f, v := range upd {
		if i > 0 {
			q.WriteString(", ")
		}
		i++
		q.WriteString(songFieldToColumn[f])
		q.WriteString(" = $")
		if f == ReleaseDate {
			args = append(args, pgtype.Date{Time: v.(time.Time), Valid: true})
		} else {
			args = append(args, v)
		}
		q.WriteString(strconv.Itoa(len(args)))
	}
	q.WriteString(" WHERE id = $")
	args = append(args, id)
	q.WriteString(strconv.Itoa(len(args)))
	s.log.Debug(ctx, "executing query", slog.String("query", q.String()), slog.Any("args", args))
	_, err := s.conn.Exec(ctx, q.String(), args...)
	return err
}
