package songs_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/songs"
	"github.com/x0k/effective-mobile-song-library-service/internal/testutils"
)

func TestSongs(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	log := logger.New(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	pgx := testutils.SetupPgx(ctx, log.Logger, t)
	musicInfoClient := testutils.SetupMusicInfoClient(ctx, t)

	router := songs.New(log, pgx, musicInfoClient)

	server := httptest.NewServer(router)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)
	e.POST("/songs").
		WithJSON(map[string]string{
			"song":  "Supermassive Black Hole",
			"group": "Muse",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().IsEqual(map[string]any{
		"id":          1,
		"group":       "Muse",
		"song":        "Supermassive Black Hole",
		"releaseDate": "16.07.2006",
		"text": []string{
			"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?",
			"Ooh\nYou set my soul alight\nOoh\nYou set my soul alight",
		},
		"link": "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	})

	e.GET("/songs").
		WithQuery("filter", `AND(EQ(group, "Muse"), ALIKE(text, "%can you hear me%"), EQ(releaseDate, DATE("16.07.2006")))`).
		Expect().
		Status(http.StatusOK).
		JSON().IsEqual([]map[string]any{
		{
			"id":          1,
			"group":       "Muse",
			"song":        "Supermassive Black Hole",
			"releaseDate": "16.07.2006",
			"text": []string{
				"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?",
				"Ooh\nYou set my soul alight\nOoh\nYou set my soul alight",
			},
			"link": "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		},
	})

	e.GET("/songs/1/lyrics").
		WithQuery("page", "2").
		WithQuery("pageSize", "1").
		Expect().
		Status(http.StatusOK).
		JSON().IsEqual([]string{
		"Ooh\nYou set my soul alight\nOoh\nYou set my soul alight",
	})

	e.PATCH("/songs/1").
		WithJSON(map[string]any{
			"group":       "group",
			"song":        "song",
			"releaseDate": "08.08.2008",
			"text":        []string{"text1", "text2"},
			"link":        "link",
		}).
		Expect().
		Status(http.StatusNoContent)

	e.GET("/songs").
		WithQuery("filter", `EQ(id, 1)`).
		Expect().
		Status(http.StatusOK).
		JSON().IsEqual([]map[string]any{
		{
			"id":          1,
			"group":       "group",
			"song":        "song",
			"releaseDate": "08.08.2008",
			"text":        []string{"text1", "text2"},
			"link":        "link",
		},
	})

	e.DELETE("/songs/1").
		Expect().
		Status(http.StatusNoContent)
}
