package handler

import (
	"bookshelf/books-service/internal/domain"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/transport/http/dto"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	auth    domain.HealthChecker
}

func NewSystemHandler(ver string, db domain.Pinger, auth domain.HealthChecker) *SystemHandler {
	return &SystemHandler{version: ver, db: db, auth: auth}
}

// хелпер функции
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details []dto.ErrorDetail) {
	log := applogger.FromContext(r.Context())
	requestID := r.Header.Get("X-Request-ID")

	// Логируем 5xx ошибки
	if status >= http.StatusInternalServerError {
		log.Error("internal server error",
			slog.String("code", code),
			slog.String("message", message),
		)
	}

	resp := dto.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Details:   details,
	}

	writeJSON(w, status, resp)
}

func writeSystemError(w http.ResponseWriter, r *http.Request, msg string) {
	writeError(w, r, http.StatusInternalServerError, "SYSTEM_ERROR", msg, nil)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, details []dto.ErrorDetail) {
	writeError(w, r,
		http.StatusUnprocessableEntity,
		"VALIDATION_ERROR",
		"request validation failed",
		details,
	)
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

func getUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, ErrEmptyContextUserID
	}

	return id, nil
}

func getIntParam(r *http.Request, paramName string) (int, error) {
	paramStr := r.URL.Query().Get(paramName)
	if paramStr == "" {
		return 0, ErrEmptyParam
	}

	return strconv.Atoi(paramStr)
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

func (h *SystemHandler) checkAuthService(ctx context.Context) (time.Duration, error) {
	// Ограничиваем время выполнения
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	startAuth := time.Now()
	if err := h.auth.HealthCheck(ctx); err != nil {
		return time.Since(startAuth), err
	}
	return time.Since(startAuth), nil
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
		Service:   "books-service",
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
	authStatus := dto.StatusReady
	var dbError, authError string

	dbDuration, err := h.checkDatabase(r.Context())
	if err != nil {
		log.Error("readiness check: database ping failed", slog.Any("error", err))
		dbStatus = dto.StatusError
		dbError = "database connection failed"
		isReady = false
	}

	authDuration, err := h.checkAuthService(r.Context())
	if err != nil {
		log.Error("readiness check: auth service failed", slog.Any("error", err))
		authStatus = dto.StatusError
		authError = "authentication service failed"
		isReady = false
	}

	resp := dto.ReadyResponse{
		Ready:     isReady,
		Service:   "books-service",
		Timestamp: time.Now(),
		Checks: map[string]dto.Check{
			"database": {
				Status:   dbStatus,
				Duration: dbDuration.String(),
				Error:    dbError,
			},
			"authentication": {
				Status:   authStatus,
				Duration: authDuration.String(),
				Error:    authError,
			},
		},
	}

	statusCode := http.StatusOK
	if !isReady {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}
