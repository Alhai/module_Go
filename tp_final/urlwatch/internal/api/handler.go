package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alhai/urlwatch/internal/domain"
	"github.com/alhai/urlwatch/internal/pool"
)

type Handler struct {
	checker domain.Checker
	store   domain.Store
	logger  *slog.Logger
}

func NewHandler(chkr domain.Checker, st domain.Store, logger *slog.Logger) *Handler {
	return &Handler{checker: chkr, store: st, logger: logger}
}

func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if err := validateRequest(&req); err != nil {
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, "invalid_request", ve.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if req.Options.Concurrency == 0 {
		req.Options.Concurrency = 8
	}
	if req.Options.TimeoutMs == 0 {
		req.Options.TimeoutMs = 5000
	}

	start := time.Now()
	results := pool.Run(r.Context(), h.checker, req.URLs, pool.Options{
		Concurrency: req.Options.Concurrency,
		TimeoutMs:   req.Options.TimeoutMs,
	})
	durationMs := time.Since(start).Milliseconds()

	batch := domain.Batch{
		ID:        generateID(),
		CreatedAt: time.Now().UTC(),
		Results:   results,
		Summary:   domain.NewSummary(results, durationMs),
	}

	if err := h.store.Save(r.Context(), batch); err != nil {
		h.logger.Error("failed to save batch", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "failed to save batch")
		return
	}

	h.logger.Info("batch created", "batch_id", batch.ID, "total", batch.Summary.Total)
	writeJSON(w, http.StatusCreated, batch)
}

func (h *Handler) GetBatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batch, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBatchNotFound) {
			writeError(w, http.StatusNotFound, "batch_not_found", fmt.Sprintf("aucun lot avec l'id %s", id))
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "failed to retrieve batch")
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: ErrorDetail{Code: code, Message: message}})
}

func generateID() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("b_%x", b)
}
