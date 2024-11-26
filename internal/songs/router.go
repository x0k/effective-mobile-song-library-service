package songs

import (
	"net/http"
)

type SongsController interface {
	GetSongs(w http.ResponseWriter, r *http.Request)
	CreateSong(w http.ResponseWriter, r *http.Request)
}

func newRouter(
	songsController SongsController,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /songs", songsController.CreateSong)
	mux.HandleFunc("GET /songs", songsController.GetSongs)
	return mux
}
