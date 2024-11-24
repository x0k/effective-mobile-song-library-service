package songs

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/db"
)

type Repo struct {
	conn    *pgx.Conn
	queries *db.Queries
}

func newRepo(conn *pgx.Conn) *Repo {
	return &Repo{
		conn:    conn,
		queries: db.New(conn),
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
