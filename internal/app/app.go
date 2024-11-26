package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	http_adapters "github.com/x0k/effective-mobile-song-library-service/internal/adapters/http"
	pgx_adapter "github.com/x0k/effective-mobile-song-library-service/internal/adapters/pgx"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger/sl"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
	"github.com/x0k/effective-mobile-song-library-service/internal/songs"
)

func Run(configPath string) {
	cfg := mustLoadConfig(configPath)
	log := mustNewLogger(&cfg.Logger)
	ctx := context.Background()

	if err := pgx_adapter.Migrate(
		ctx,
		log.Logger.With(slog.String("component", "pgx_migrate")),
		cfg.Postgres.ConnectionURI,
		cfg.Postgres.MigrationsURI,
	); err != nil {
		log.Error(ctx, "cannot migrate database", sl.Err(err))
		os.Exit(1)
	}

	pgx, err := pgx.Connect(ctx, cfg.Postgres.ConnectionURI)
	if err != nil {
		log.Error(ctx, "cannot connect to postgres", sl.Err(err))
		os.Exit(1)
	}
	defer pgx.Close(ctx)

	musicInfoClient, err := music_info.NewClientWithResponses(cfg.MusicInfoService.Address)
	if err != nil {
		log.Error(ctx, "cannot create music info client", sl.Err(err))
		os.Exit(1)
	}

	router := songs.New(
		log,
		pgx,
		musicInfoClient,
	)

	srv := http.Server{
		Addr: cfg.Server.Address,
		Handler: http_adapters.Logging(
			log.With(slog.String("component", "http_server")),
			router,
		),
	}

	go func() {
		log.Info(ctx, "starting server", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "cannot start server", sl.Err(err))
			os.Exit(1)
		}
	}()

	log.Info(ctx, "press CTRL-C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	s := <-stop
	log.Info(ctx, "signal received", slog.String("signal", s.String()))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(ctx, "force shutdown", sl.Err(err))
		os.Exit(1)
	}
	log.Info(ctx, "graceful shutdown")
}
