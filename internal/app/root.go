package app

import (
	"log/slog"
	"net/http"

	http_adapters "github.com/x0k/effective-mobile-song-library-service/internal/adapters/http"
	songs_controller "github.com/x0k/effective-mobile-song-library-service/internal/controllers/http/songs"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
	songs_service "github.com/x0k/effective-mobile-song-library-service/internal/services/songs"
	db_storage "github.com/x0k/effective-mobile-song-library-service/internal/storage/db"
)

func newRoot(
	cfg *Config,
	log *logger.Logger,
) (*module.Root, error) {
	m := module.NewRoot(log.Logger)

	storage := db_storage.New(
		log.With(slog.String("component", "storage")),
		cfg.Postgres.ConnectionURI,
	)
	m.PreStartFn("storage_open", storage.Open)
	m.PostStopFn("storage_close", storage.Close)

	musicInfoClient, err := music_info.NewClientWithResponses(cfg.MusicInfoService.Address)
	if err != nil {
		return nil, err
	}
	songsService := songs_service.New(
		storage,
		musicInfoClient,
	)
	songsController := songs_controller.New(
		log.With(slog.String("component", "songs_controller")),
		songsService.CreateSong,
	)
	srv := &http.Server{
		Addr: cfg.Server.Address,
		Handler: http_adapters.Logging(
			log.With(slog.String("component", "http_router")),
			newRouter(songsController),
		),
	}
	m.Append(http_adapters.NewService("http_server", srv, m))

	return m, nil
}
