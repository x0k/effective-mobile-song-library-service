package songs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/httpx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger/sl"
)

var ErrLastIdCannotBeUsedWithPageParameter = errors.New("last id cannot be used with page parameter")
var ErrFilterIsTooLong = errors.New("filter is too complex")

type SongsService interface {
	CreateSong(ctx context.Context, song string, group string) (Song, error)
	GetSongs(ctx context.Context, query Query) ([]Song, error)
}

type songsController struct {
	log          *logger.Logger
	decoder      *httpx.JsonBodyDecoder
	songsService SongsService
}

func NewController(
	log *logger.Logger,
	songsRepo SongsService,
) *songsController {
	return &songsController{
		log: log,
		decoder: &httpx.JsonBodyDecoder{
			MaxBytes: 1 * 1024 * 1024,
		},
		songsService: songsRepo,
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

func toDTO(song Song) songDTO {
	return songDTO{
		ID:          song.ID,
		Title:       song.Title,
		Artist:      song.Artist,
		ReleaseDate: httpx.NewJsonDate(song.ReleaseDate, releaseDateFormat),
		Lyrics:      song.Lyrics,
		Link:        song.Link,
	}
}

func (c *songsController) CreateSong(w http.ResponseWriter, r *http.Request) {
	createSong, httpErr := httpx.JSONBody[createSongDTO](c.log.Logger, c.decoder, w, r)
	if httpErr != nil {
		http.Error(w, httpErr.Text, httpErr.Status)
		return
	}
	if len(strings.TrimSpace(createSong.Group)) == 0 {
		http.Error(w, "group is required", http.StatusBadRequest)
		c.log.Debug(r.Context(), "group is empty")
		return
	}
	if len(strings.TrimSpace(createSong.Song)) == 0 {
		http.Error(w, "song is required", http.StatusBadRequest)
		c.log.Debug(r.Context(), "song is empty")
		return
	}
	song, err := c.songsService.CreateSong(r.Context(), createSong.Song, createSong.Group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		c.log.Debug(r.Context(), "failed to create song", sl.Err(err))
		return
	}
	c.json(w, r, toDTO(song), http.StatusCreated)
}

func (c *songsController) GetSongs(w http.ResponseWriter, r *http.Request) {
	rq, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		c.log.Debug(r.Context(), "failed to parse query", sl.Err(err))
		return
	}
	sq := Query{
		PageSize: 100,
	}
	if sq.Page, err = c.parseUint(rq, "page", 64); err != nil {
		c.badRequest(w, r, err)
		return
	}
	if pageSize, err := c.parseUint(rq, "pageSize", 64); err != nil {
		c.badRequest(w, r, err)
		return
	} else if pageSize < sq.PageSize {
		sq.PageSize = pageSize
	}
	if lastId, err := c.parseUint(rq, "lastId", 63); err != nil {
		c.badRequest(w, r, err)
		return
	} else if lastId > 0 && sq.Page > 0 {
		c.badRequest(w, r, ErrLastIdCannotBeUsedWithPageParameter)
		return
	} else {
		sq.LastId = int64(lastId)
	}
	if sq.Filter = rq.Get("filter"); len(sq.Filter) > 500 {
		c.badRequest(w, r, ErrFilterIsTooLong)
		return
	}
	songs, err := c.songsService.GetSongs(r.Context(), sq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		c.log.Debug(r.Context(), "failed to get songs", sl.Err(err))
		return
	}
	dtos := make([]songDTO, len(songs))
	for i, song := range songs {
		dtos[i] = toDTO(song)
	}
	c.json(w, r, dtos, http.StatusOK)
}

func (c *songsController) json(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		c.log.Debug(r.Context(), "failed to encode JSON", sl.Err(err))
		return
	}
}

func (c *songsController) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
	c.log.Debug(r.Context(), "bad request", sl.Err(err))
}

func (c *songsController) parseUint(q url.Values, name string, bitSize int) (uint64, error) {
	v := q.Get(name)
	if len(v) == 0 {
		return 0, nil
	}
	r, err := strconv.ParseUint(v, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %q query parameter: %w", name, err)
	}
	return r, nil
}
