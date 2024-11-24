package app

import (
	"log/slog"
	"net/http"

	http_adapters "github.com/x0k/effective-mobile-song-library-service/internal/adapters/http"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
	"github.com/x0k/effective-mobile-song-library-service/internal/router"
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

	srv := &http.Server{
		Addr: cfg.Server.Address,
		Handler: http_adapters.Logging(
			log.With(slog.String("component", "http_router")),
			router.New(),
		),
	}
	m.Append(http_adapters.NewService("http_server", srv, m))

	return m, nil
}
