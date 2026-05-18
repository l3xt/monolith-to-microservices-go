package handler

import (
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/transport/http/dto"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type TokenValidator interface {
	VerifyToken(ctx context.Context, req *dto.TokenRequest) (*dto.VerifyResponse, error)
}

func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := applogger.FromContext(r.Context())
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required", nil)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format", nil)
				return
			}
			// Идем в users-service по сети
			resp, err := validator.VerifyToken(r.Context(), &dto.TokenRequest{Token: parts[1]})
			if err != nil {
				log.Warn("invalid token", slog.Any("error", err))
				writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", nil)
				return
			}
			// Кладем ID в контекст
			ctx := context.WithValue(r.Context(), userIDKey, resp.UserID)

			// Кладем ID в логгер
			ctx = applogger.WithContext(ctx, log.With(
				slog.String("user_id", resp.UserID.String()),
			))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
