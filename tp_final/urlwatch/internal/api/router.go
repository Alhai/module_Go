package api

import (
	"log/slog"
	"net/http"
)

func NewRouter(h *Handler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("POST /v1/checks", h.CreateBatch)
	mux.HandleFunc("GET /v1/checks/{id}", h.GetBatch)

	// Logging wraps mux; Recovery wraps everything so panics are always caught.
	var handler http.Handler = mux
	handler = LoggingMiddleware(logger, handler)
	handler = RecoveryMiddleware(logger, handler)
	return handler
}
