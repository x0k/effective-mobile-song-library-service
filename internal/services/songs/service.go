package songs_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/x0k/effective-mobile-song-library-service/internal/entities"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
	db_storage "github.com/x0k/effective-mobile-song-library-service/internal/storage/db"
)

var ErrFailedToGetInfo = errors.New("failed to get info")
var ErrFailedToParseReleaseDate = errors.New("failed to parse release date")
var ErrFailedToSaveSong = errors.New("failed to save song")

type SongsService struct {
	musicInfo music_info.ClientWithResponsesInterface
	storage   *db_storage.Storage
}

func New(
	storage *db_storage.Storage,
	musicInfo music_info.ClientWithResponsesInterface,
) *SongsService {
	return &SongsService{
		musicInfo: musicInfo,
		storage:   storage,
	}
}

func (s *SongsService) CreateSong(ctx context.Context, title string, artist string) (entities.Song, error) {
	r, err := s.musicInfo.GetInfoWithResponse(ctx, &music_info.GetInfoParams{
		Group: artist,
		Song:  title,
	})
	if err != nil {
		return entities.Song{}, fmt.Errorf("%w: %v", ErrFailedToGetInfo, err)
	}
	lyrics := strings.Split(r.JSON200.Text, "\n\n")
	releaseDate, err := time.Parse("02.01.2006", r.JSON200.ReleaseDate)
	if err != nil {
		return entities.Song{}, fmt.Errorf("%w: %v", ErrFailedToParseReleaseDate, err)
	}
	song := entities.NewSong(
		title,
		artist,
		releaseDate,
		lyrics,
		r.JSON200.Link,
	)
	if err := s.storage.SaveSong(ctx, &song); err != nil {
		return entities.Song{}, fmt.Errorf("%w: %v", ErrFailedToSaveSong, err)
	}
	return song, nil
}
