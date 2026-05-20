package handler

import (
	applogger "bookshelf/users-service/internal/logger"
	"bookshelf/users-service/internal/service"
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func AuthMiddleware(tokenManager service.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := applogger.FromContext(r.Context())

			// Извлечение токена
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("authorization header is missing")
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required", nil)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("invalid authorization header format")
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format", nil)
				return
			}

			claims, err := tokenManager.Validate(parts[1])
			if err != nil {
				log.Warn("invalid or expired token")
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", nil)
				return
			}

			// Пробрасываем через контекст
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

			// Кладем user_id в логгер
			ctx = applogger.WithContext(ctx, log.With(
				slog.String("user_id", claims.UserID.String()),
			))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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

func ServiceKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := applogger.FromContext(r.Context())

			key := r.Header.Get("X-Service-Key")
			if key == "" {
				log.Warn("service key is missing")
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Service key required", nil)
				return
			}
			if key != expectedKey {
				log.Warn("invalid service key", slog.String("expected_key", expectedKey), slog.String("provided_key", key))
				writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Invalid service key", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
