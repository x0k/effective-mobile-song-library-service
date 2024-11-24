package app

import (
	"net/http"

	songs_controller "github.com/x0k/effective-mobile-song-library-service/internal/controllers/http/songs"
)

func newRouter(
	songsController *songs_controller.Controller,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /songs", songsController.CreateSong)
	return mux
}
