package handler

import (
	applogger "bookshelf/books-service/internal/logger"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			reqLogger := logger.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_ip", r.RemoteAddr),
			)

			ctx := applogger.WithContext(r.Context(), reqLogger)

			start := time.Now()
			h.ServeHTTP(w, r.WithContext(ctx))

			reqLogger.Info("request completed",
				slog.Duration("latency", time.Since(start)),
			)
		})
	}
}
