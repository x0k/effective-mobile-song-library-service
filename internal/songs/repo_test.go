package songs

import (
	"bytes"
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/testutils"
)

func TestRepo_SaveSong(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	log := logger.New(slog.New(slog.NewTextHandler(&buf, nil)))
	pgx := testutils.SetupPgx(ctx, log.Logger, t)
	repo := newRepo(log, pgx)

	song := Song{
		Title:       "title",
		Artist:      "artist",
		ReleaseDate: time.Now().Truncate(24 * time.Hour),
		Lyrics:      []string{"lyrics"},
		Link:        "link",
	}
	err := repo.SaveSong(ctx, &song)
	if err != nil {
		t.Fatal(err)
	}
	if song.ID == -1 {
		t.Fatal("song.ID == -1")
	}
	row := pgx.QueryRow(ctx, "select * from song where id = $1", song.ID)
	var savedSong Song
	var d pgtype.Date
	if err := row.Scan(&savedSong.ID, &savedSong.Title, &savedSong.Artist, &d, &savedSong.Lyrics, &savedSong.Link); err != nil {
		t.Fatal(err)
	}
	savedSong.ReleaseDate = d.Time.In(time.Local)
	if !reflect.DeepEqual(song, savedSong) {
		t.Log(song)
		t.Log(savedSong)
		t.Fatal("song != savedSong")
	}
}
