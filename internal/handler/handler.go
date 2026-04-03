package handler

import (
	"bookshelf/internal/domain"
	applogger "bookshelf/internal/logger"
	"bookshelf/internal/transport/http/dto"
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
	db domain.Pinger
}

func NewSystemHandler(db domain.Pinger) *SystemHandler {
	return &SystemHandler{db: db}
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

// health функции
func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := dto.HealthResponse{
		Status:    dto.StatusReady,
		Version:   "1.0.0",
		Timestamp: time.Now(),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	dbStatus := dto.StatusReady

	if err := h.db.Ping(r.Context()); err != nil {
		dbStatus = dto.StatusError
	}

	resp := dto.ReadyResponse{
		HealthResponse: dto.HealthResponse{
			Status:    dto.StatusReady,
			Version:   "1.0.0",
			Timestamp: time.Now(),
		},
		Checks: dto.CheckList{
			Database: dbStatus,
		},
	}

	statusCode := http.StatusOK
	if dbStatus != dto.StatusReady {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}
