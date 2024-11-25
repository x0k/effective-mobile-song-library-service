package songs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
)

var ErrFailedToGetInfo = errors.New("failed to get info")
var ErrFailedToParseReleaseDate = errors.New("failed to parse release date")
var ErrFailedToSaveSong = errors.New("failed to save song")

type SongSaver = func(ctx context.Context, song *Song) error

type SongsService struct {
	musicInfo music_info.ClientWithResponsesInterface
	saveSong  SongSaver
}

func newService(
	musicInfo music_info.ClientWithResponsesInterface,
	songSaver SongSaver,
) *SongsService {
	return &SongsService{
		musicInfo: musicInfo,
		saveSong:  songSaver,
	}
}

func (s *SongsService) CreateSong(ctx context.Context, title string, artist string) (Song, error) {
	r, err := s.musicInfo.GetInfoWithResponse(ctx, &music_info.GetInfoParams{
		Group: artist,
		Song:  title,
	})
	if err != nil {
		return Song{}, fmt.Errorf("%w: %v", ErrFailedToGetInfo, err)
	}
	if r.JSON200 == nil {
		return Song{}, fmt.Errorf("%w: no response", ErrFailedToGetInfo)
	}
	lyrics := strings.Split(r.JSON200.Text, "\n\n")
	releaseDate, err := time.Parse("02.01.2006", r.JSON200.ReleaseDate)
	if err != nil {
		return Song{}, fmt.Errorf("%w: %v", ErrFailedToParseReleaseDate, err)
	}
	song := NewSong(
		title,
		artist,
		releaseDate,
		lyrics,
		r.JSON200.Link,
	)
	if err := s.saveSong(ctx, &song); err != nil {
		return Song{}, fmt.Errorf("%w: %v", ErrFailedToSaveSong, err)
	}
	return song, nil
}
