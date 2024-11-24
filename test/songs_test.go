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
			"song":  "song",
			"group": "group",
		}).
		Expect().
		Status(http.StatusCreated)
}
