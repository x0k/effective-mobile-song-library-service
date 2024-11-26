package songs

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/db"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/filter"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
)

type Repo struct {
	log     *logger.Logger
	conn    *pgx.Conn
	queries *db.Queries
	filter  *filter.Filter
}

func newRepo(log *logger.Logger, conn *pgx.Conn) *Repo {
	return &Repo{
		log:     log,
		conn:    conn,
		queries: db.New(conn),
		filter: filter.New(
			"song",
			map[string]filter.ValueType{
				"id":          filter.NumberType,
				"title":       filter.StringType,
				"artist":      filter.StringType,
				"releaseDate": filter.DateType,
				"lyrics":      filter.ArrayOf(filter.StringType),
				"link":        filter.StringType,
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

func (s *Repo) SaveSong(ctx context.Context, song *Song) error {
	id, err := s.queries.InsertSongAndReturnId(ctx, db.InsertSongAndReturnIdParams{
		Title:       song.Title,
		Artist:      song.Artist,
		ReleaseDate: pgtype.Date{Time: song.ReleaseDate, Valid: true},
		Lyrics:      song.Lyrics,
		Link:        song.Link,
	})
	if err != nil {
		return err
	}
	song.ID = id
	return nil
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
