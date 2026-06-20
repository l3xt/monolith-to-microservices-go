package handler

import (
	"bookshelf/worker-service/internal/domain"
	applogger "bookshelf/worker-service/internal/logger"
	"bookshelf/worker-service/internal/transport/http/dto"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Ключи для хранения userID в контексте
type contextKey string

const userIDKey contextKey = "userID"

const defaultMaxBytes = 1024 // 1 MB

var (
	ErrEmptyContextUserID = errors.New("userID not found in context")
	ErrEmptyParam         = errors.New("param is empty")
)

type SystemHandler struct {
	version string
	db      domain.Pinger
}

func NewSystemHandler(ver string, db domain.Pinger) *SystemHandler {
	return &SystemHandler{version: ver, db: db}
}

// хелпер функции
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T

	// Ограничиваем тело запроса
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)

	// Запрещаем неизвестные поля
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func (h *SystemHandler) checkDatabase(ctx context.Context) (time.Duration, error) {
	// Ограничиваем время выполнения
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	startDB := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		return time.Since(startDB), err
	}
	return time.Since(startDB), nil
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	generalStatus := dto.StatusReady
	dbStatus := dto.StatusReady
	var dbError string

	dbDuration, err := h.checkDatabase(r.Context())
	if err != nil {
		log.Error("health check: database ping failed", slog.Any("error", err))
		dbStatus = dto.StatusError
		dbError = "database connection failed"
	}

	if dbStatus == dto.StatusError {
		generalStatus = dto.StatusError
	}

	resp := dto.HealthResponse{
		Status:    generalStatus,
		Service:   "worker-service",
		Version:   h.version,
		Timestamp: time.Now(),
		Checks: map[string]dto.Check{
			"database": {
				Status:   dbStatus,
				Duration: dbDuration.String(),
				Error:    dbError,
			},
		},
	}

	statusCode := http.StatusOK
	if dbStatus != dto.StatusReady {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}

func (h *SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	isReady := true
	dbStatus := dto.StatusReady
	var dbError string

	dbDuration, err := h.checkDatabase(r.Context())
	if err != nil {
		log.Error("readiness check: database ping failed", slog.Any("error", err))
		dbStatus = dto.StatusError
		dbError = "database connection failed"
		isReady = false
	}

	resp := dto.ReadyResponse{
		Ready:     isReady,
		Service:   "worker-service",
		Timestamp: time.Now(),
		Checks: map[string]dto.Check{
			"database": {
				Status:   dbStatus,
				Duration: dbDuration.String(),
				Error:    dbError,
			},
		},
	}

	statusCode := http.StatusOK
	if !isReady {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}
