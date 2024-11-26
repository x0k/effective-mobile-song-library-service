package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
	"github.com/x0k/effective-mobile-song-library-service/internal/songs"
)

func newRouter(
	log *logger.Logger,
	pgx *pgx.Conn,
	musicInfoClient music_info.ClientWithResponsesInterface,
) http.Handler {
	songsRepo := songs.NewRepo(
		log.With(slog.String("component", "songs_repo")),
		pgx,
	)

	songsService := songs.NewService(
		musicInfoClient,
		songsRepo,
	)

	songsController := songs.NewController(
		log.With(slog.String("component", "songs_controller")),
		songsService,
	)

	return songs.NewRouter(songsController)
}
