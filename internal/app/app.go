package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger/sl"
)

func Run() {
	cfg := mustLoadConfig()
	log := mustNewLogger(&cfg.Logger)
	ctx := context.Background()
	log.Info(ctx, "starting app", slog.String("log_level", cfg.Logger.Level))
	root, err := newRoot(cfg, log)
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
