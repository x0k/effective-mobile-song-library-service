package songs

import (
	"net/http"
)

type SongsController interface {
	GetSongs(w http.ResponseWriter, r *http.Request)
	CreateSong(w http.ResponseWriter, r *http.Request)
	GetLyrics(w http.ResponseWriter, r *http.Request)
	DeleteSong(w http.ResponseWriter, r *http.Request)
	UpdateSong(w http.ResponseWriter, r *http.Request)
}

func newRouter(
	songsController SongsController,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /songs", songsController.CreateSong)
	mux.HandleFunc("GET /songs", songsController.GetSongs)
	mux.HandleFunc("GET /songs/{songId}/lyrics", songsController.GetLyrics)
	mux.HandleFunc("DELETE /songs/{songId}", songsController.DeleteSong)
	mux.HandleFunc("PATCH /songs/{songId}", songsController.UpdateSong)
	return mux
}
