package app

import (
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/x0k/effective-mobile-song-library-service/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/lib/module"
	pgx_storage "github.com/x0k/effective-mobile-song-library-service/storage/pgx"
)

func newRoot(
	cfg *Config,
	log *logger.Logger,
	sourceName string,
	migrations source.Driver,
) (*module.Root, error) {
	m := module.NewRoot(log.Logger)

	storage := pgx_storage.New(
		cfg.Postgres.ConnectionString,
		sourceName,
		migrations,
	)
	m.PreStartFn("storage_open", storage.Open)
	m.PostStopFn("storage_close", storage.Close)

	return m, nil
}
