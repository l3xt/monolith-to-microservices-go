package handler

import (
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

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

type SystemHandler struct {
	version string
	db      HealthChecker
	auth    HealthChecker
	broker  HealthChecker
	storage HealthChecker
}

func NewSystemHandler(ver string, db, auth, broker, storage HealthChecker) *SystemHandler {
	return &SystemHandler{version: ver, db: db, auth: auth, broker: broker, storage: storage}
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
	if err := h.db.HealthCheck(ctx); err != nil {
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

func (h *SystemHandler) checkStorage(ctx context.Context) (time.Duration, error) {
	// Ограничиваем время выполнения
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	startStorage := time.Now()
	if err := h.storage.HealthCheck(ctx); err != nil {
		return time.Since(startStorage), err
	}
	return time.Since(startStorage), nil
}

func (h *SystemHandler) checkBroker(ctx context.Context) (time.Duration, error) {
	// Ограничиваем время выполнения
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	startBroker := time.Now()
	if err := h.broker.HealthCheck(ctx); err != nil {
		return time.Since(startBroker), err
	}
	return time.Since(startBroker), nil
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, dto.HealthResponse{
		Status:    dto.StatusReady,
		Service:   "books-service",
		Timestamp: time.Now(),
	})
}

func (h *SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	generalStatus := dto.StatusReady
	dbStatus := dto.StatusReady
	storageStatus := dto.StatusReady
	brokerStatus := dto.StatusReady

	var dbError, storageError, brokerError string

	dbDuration, err := h.checkDatabase(r.Context())
	if err != nil {
		log.Error("SystemHandler.Ready: database ping failed", slog.Any("error", err))
		dbStatus = dto.StatusError
		dbError = "database connection failed"
	}

	storageDuration, err := h.checkStorage(r.Context())
	if err != nil {
		log.Error("SystemHandler.Ready: storage ping failed", slog.Any("error", err))
		storageStatus = dto.StatusError
		storageError = "storage connection failed"
	}

	brokerDuration, err := h.checkBroker(r.Context())
	if err != nil {
		log.Error("SystemHandler.Ready: broker ping failed", slog.Any("error", err))
		brokerStatus = dto.StatusError
		brokerError = "broker connection failed"
	}

	if dbStatus == dto.StatusError || storageStatus == dto.StatusError || brokerStatus == dto.StatusError {
		generalStatus = dto.StatusError
	}

	resp := dto.ReadyResponse{
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
			"storage": {
				Status:   storageStatus,
				Duration: storageDuration.String(),
				Error:    storageError,
			},
			"broker": {
				Status:   brokerStatus,
				Duration: brokerDuration.String(),
				Error:    brokerError,
			},
		},
	}

	statusCode := http.StatusOK
	if generalStatus != dto.StatusReady {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}

