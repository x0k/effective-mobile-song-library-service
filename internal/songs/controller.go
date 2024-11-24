package songs

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/httpx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
)

type SongCreator = func(ctx context.Context, song string, group string) (Song, error)

type Controller struct {
	log         *logger.Logger
	decoder     *httpx.JsonBodyDecoder
	songCreator SongCreator
}

func newController(
	log *logger.Logger,
	songCreator SongCreator,
) *Controller {
	return &Controller{
		log: log,
		decoder: &httpx.JsonBodyDecoder{
			MaxBytes: 1 * 1024 * 1024,
		},
		songCreator: songCreator,
	}
}

type createSongDTO struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type songDTO struct {
	ID          int64          `json:"id"`
	Title       string         `json:"song"`
	Artist      string         `json:"group"`
	ReleaseDate httpx.JsonDate `json:"releaseDate"`
	Lyrics      []string       `json:"text"`
	Link        string         `json:"link"`
}

func (c *Controller) CreateSong(w http.ResponseWriter, r *http.Request) {
	createSong, httpErr := httpx.JSONBody[createSongDTO](c.log.Logger, c.decoder, w, r)
	if httpErr != nil {
		http.Error(w, httpErr.Text, httpErr.Status)
		return
	}
	if len(strings.TrimSpace(createSong.Group)) == 0 {
		http.Error(w, "group is required", http.StatusBadRequest)
		return
	}
	if len(strings.TrimSpace(createSong.Song)) == 0 {
		http.Error(w, "song is required", http.StatusBadRequest)
		return
	}
	song, err := c.songCreator(r.Context(), createSong.Song, createSong.Group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := songDTO{
		ID:          song.ID,
		Title:       song.Title,
		Artist:      song.Artist,
		ReleaseDate: httpx.NewJsonDate(song.ReleaseDate, "02.01.2006"),
		Lyrics:      song.Lyrics,
		Link:        song.Link,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}
