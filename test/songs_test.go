package test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestSongs(t *testing.T) {
	ctx := context.Background()
	appAddress := setupApp(ctx, t)
	e := httpexpect.Default(t, appAddress)
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
		WithQuery("filter", `AND(EQ(artist, "Muse"), ALIKE(lyrics, "%can you hear me%"))`).
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
}
