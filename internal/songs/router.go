package songs

import (
	"net/http"
)

func newRouter(
	songsController *Controller,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /songs", songsController.CreateSong)
	return mux
}
