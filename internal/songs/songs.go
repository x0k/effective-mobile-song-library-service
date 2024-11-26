package songs

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
	"github.com/x0k/effective-mobile-song-library-service/internal/lib/music_info"
)

func New(
	log *logger.Logger,
	pgx *pgx.Conn,
	musicInfoClient music_info.ClientWithResponsesInterface,
) http.Handler {
	songsRepo := newRepo(
		log.With(slog.String("component", "songs_repo")),
		pgx,
	)

	songsService := newService(
		musicInfoClient,
		songsRepo,
	)

	songsController := newController(
		log.With(slog.String("component", "songs_controller")),
		songsService,
	)

	return newRouter(songsController)
}
