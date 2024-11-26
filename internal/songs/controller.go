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
	"time"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/httpx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger/sl"
)

var ErrLastIdCannotBeUsedWithPageParameter = errors.New("last id cannot be used with page parameter")
var ErrFilterIsTooLong = errors.New("filter is too complex")
var ErrInvalidDate = errors.New("invalid date")
var ErrNothingToUpdate = errors.New("nothing to update")
var ErrInvalidField = errors.New("invalid song field")

type SongsService interface {
	CreateSong(ctx context.Context, song string, group string) (Song, error)
	GetSongs(ctx context.Context, query Query) ([]Song, error)
	GetLyrics(ctx context.Context, id int64, pagination Pagination) ([]string, error)
	DeleteSong(ctx context.Context, id int64) error
	UpdateSong(ctx context.Context, id int64, songUpdate SongUpdate) error
}

type songsController struct {
	log          *logger.Logger
	decoder      *httpx.JsonBodyDecoder
	songsService SongsService
	maxPageSize  uint64
}

func newController(
	log *logger.Logger,
	songsRepo SongsService,
) *songsController {
	return &songsController{
		log: log,
		decoder: &httpx.JsonBodyDecoder{
			MaxBytes: 1 * 1024 * 1024,
		},
		songsService: songsRepo,
		maxPageSize:  100,
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

type updateSongDTO struct {
	Title       *string   `json:"song"`
	Artist      *string   `json:"group"`
	ReleaseDate *string   `json:"releaseDate"`
	Lyrics      *[]string `json:"text"`
	Link        *string   `json:"link"`
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
		c.badRequest(w, r, fmt.Errorf("%w: %v", ErrInvalidField, "group"))
		return
	}
	if len(strings.TrimSpace(createSong.Song)) == 0 {
		c.badRequest(w, r, fmt.Errorf("%w: %v", ErrInvalidField, "song"))
		return
	}
	song, err := c.songsService.CreateSong(r.Context(), createSong.Song, createSong.Group)
	if err != nil {
		c.serverError(w, r, err, "failed to create song")
		return
	}
	c.json(w, r, toDTO(song), http.StatusCreated)
}

func (c *songsController) GetSongs(w http.ResponseWriter, r *http.Request) {
	rq, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		c.badRequest(w, r, err)
		return
	}
	sq := Query{
		Pagination: Pagination{
			PageSize: c.maxPageSize,
		},
	}
	if err = c.parsePagination(&sq.Pagination, rq); err != nil {
		c.badRequest(w, r, err)
		return
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
		c.serverError(w, r, err, "failed to get songs")
		return
	}
	dtos := make([]songDTO, len(songs))
	for i, song := range songs {
		dtos[i] = toDTO(song)
	}
	c.json(w, r, dtos, http.StatusOK)
}

func (c *songsController) GetLyrics(w http.ResponseWriter, r *http.Request) {
	songId, err := c.parseSongId(r)
	if err != nil {
		c.badRequest(w, r, err)
		return
	}
	Pagination := Pagination{
		PageSize: c.maxPageSize,
	}
	if err := c.parsePagination(&Pagination, r.URL.Query()); err != nil {
		c.badRequest(w, r, err)
		return
	}
	lyrics, err := c.songsService.GetLyrics(r.Context(), songId, Pagination)
	if err != nil {
		c.serverError(w, r, err, "failed to get lyrics")
		return
	}
	c.json(w, r, lyrics, http.StatusOK)
}

func (c *songsController) DeleteSong(w http.ResponseWriter, r *http.Request) {
	songId, err := c.parseSongId(r)
	if err != nil {
		c.badRequest(w, r, err)
		return
	}
	if err := c.songsService.DeleteSong(r.Context(), songId); err != nil {
		c.serverError(w, r, err, "failed to delete song")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *songsController) UpdateSong(w http.ResponseWriter, r *http.Request) {
	songId, err := c.parseSongId(r)
	if err != nil {
		c.badRequest(w, r, err)
		return
	}
	u, httpErr := httpx.JSONBody[updateSongDTO](c.log.Logger, c.decoder, w, r)
	if httpErr != nil {
		http.Error(w, httpErr.Text, httpErr.Status)
		return
	}
	upd := make(SongUpdate, 5)
	if u.Title != nil {
		upd[Title] = *u.Title
	}
	if u.Artist != nil {
		upd[Artist] = *u.Artist
	}
	if u.ReleaseDate != nil {
		time, err := time.Parse(releaseDateFormat, *u.ReleaseDate)
		if err != nil {
			c.badRequest(w, r, fmt.Errorf("%w: %v", ErrInvalidDate, err))
			return
		}
		upd[ReleaseDate] = time
	}
	if u.Lyrics != nil {
		upd[Lyrics] = *u.Lyrics
	}
	if u.Link != nil {
		upd[Link] = *u.Link
	}
	if len(upd) == 0 {
		c.badRequest(w, r, ErrNothingToUpdate)
		return
	}
	if err := c.songsService.UpdateSong(r.Context(), songId, upd); err != nil {
		c.serverError(w, r, err, "failed to update song")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *songsController) parseSongId(r *http.Request) (int64, error) {
	songIdStr := r.PathValue("songId")
	if songIdStr == "" {
		return 0, fmt.Errorf("%w: %v", ErrInvalidField, "songId is empty")
	}
	songId, err := strconv.ParseInt(songIdStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: songId %v", ErrInvalidField, err)
	}
	return songId, nil
}

func (c *songsController) parsePagination(p *Pagination, rq url.Values) error {
	var err error
	if p.Page, err = c.parseUint(rq, "page", 64); err != nil {
		return err
	}
	if pageSize, err := c.parseUint(rq, "pageSize", 64); err != nil {
		return err
	} else if pageSize > 0 && pageSize < p.PageSize {
		p.PageSize = pageSize
	}
	return nil
}

func (c *songsController) json(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		c.serverError(w, r, err, "failed to encode JSON")
	}
}

func (c *songsController) serverError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	c.log.Debug(r.Context(), msg, sl.Err(err))
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
