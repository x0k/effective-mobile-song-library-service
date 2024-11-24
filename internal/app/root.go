package app

import (
	"log/slog"

	pgx_adapter "github.com/x0k/effective-mobile-song-library-service/internal/adapters/pgx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/module"
	"github.com/x0k/effective-mobile-song-library-service/internal/songs"
)

func newRoot(
	cfg *Config,
	log *logger.Logger,
) (module.Interface, error) {
	m := module.NewRoot(
		log.Logger.With(slog.String("module", "root")),
	)

	m.PreStart(pgx_adapter.NewMigrator(
		"pgx_migrate",
		log.Logger.With(slog.String("component", "pgx_migrate")),
		cfg.Postgres.ConnectionURI,
		cfg.Postgres.MigrationsURI,
	))

	pgx := pgx_adapter.NewPgx(m, cfg.Postgres.ConnectionURI)

	songsModule, err := songs.New(
		log,
		pgx,
		cfg.Server.Address,
		cfg.MusicInfoService.Address,
	)
	if err != nil {
		return nil, err
	}
	m.Append(songsModule)
	return m, nil
}
