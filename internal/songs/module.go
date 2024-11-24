package songs

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	http_adapters "github.com/x0k/effective-mobile-song-library-service/internal/adapters/http"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
)

func New(
	log *logger.Logger,
	pgx *pgx.Conn,
	songsServerAddress string,
	musicInfoServiceAddress string,
) (module.Interface, error) {
	m := module.New(
		log.Logger.With(slog.String("module", "songs")),
		"songs",
	)

	musicInfoClient, err := music_info.NewClientWithResponses(musicInfoServiceAddress)
	if err != nil {
		return nil, err
	}
	songsRepo := newRepo(pgx)
	songsService := newService(
		musicInfoClient,
		songsRepo.SaveSong,
	)
	songsController := newController(
		log.With(slog.String("component", "songs_controller")),
		songsService.CreateSong,
	)
	srv := &http.Server{
		Addr: songsServerAddress,
		Handler: http_adapters.Logging(
			log.With(slog.String("component", "http_router")),
			newRouter(songsController),
		),
	}
	m.Append(http_adapters.NewService("http_server", srv, m))

	return m, nil
}
