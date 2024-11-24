package http_adapters

import (
	"net/http"

	"log/slog"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/logger"
)

type statusCapturer struct {
	http.ResponseWriter
	status int
}

func (sc *statusCapturer) WriteHeader(status int) {
	sc.ResponseWriter.WriteHeader(status)
	sc.status = status
}

func Logging(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := statusCapturer{
			ResponseWriter: w,
		}
		next.ServeHTTP(&c, r)
		log.Info(
			r.Context(),
			"request",
			slog.String("method", r.Method),
			slog.String("url", r.RequestURI),
			slog.Int("status", c.status),
		)
	})
}
