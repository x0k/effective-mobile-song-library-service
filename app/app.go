package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/x0k/effective-mobile-song-library-service/lib/logger/sl"
)

func Run(migrations source.Driver) {
	cfg := mustLoadConfig()
	log := mustNewLogger(&cfg.Logger)
	ctx := context.Background()
	log.Info(ctx, "starting app", slog.String("log_level", cfg.Logger.Level))
	root, err := newRoot(cfg, log, migrations)
	if err != nil {
		log.Error(ctx, "failed to initialize app", sl.Err(err))
		os.Exit(1)
	}
	if err := root.Start(ctx); err != nil {
		log.Error(ctx, "failed to start app", sl.Err(err))
		os.Exit(1)
	}
	log.Info(ctx, "app stopped")
}
