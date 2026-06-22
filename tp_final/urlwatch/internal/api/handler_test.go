package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alhai/urlwatch/internal/api"
	"github.com/alhai/urlwatch/internal/checker"
	"github.com/alhai/urlwatch/internal/domain"
	"github.com/alhai/urlwatch/internal/store"
)

func noopLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newHandler(results map[string]domain.CheckResult) *api.Handler {
	return api.NewHandler(
		&checker.MockChecker{Results: results},
		store.NewMemoryStore(),
		noopLogger(),
	)
}

func TestCreateBatchSuccess(t *testing.T) {
	h := newHandler(map[string]domain.CheckResult{
		"https://go.dev": {URL: "https://go.dev", StatusCode: 200, OK: true, LatencyMs: 10},
	})

	body := `{"urls":["https://go.dev"],"options":{"concurrency":1,"timeout_ms":1000}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/checks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateBatch(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp domain.Batch
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty batch_id")
	}
	if resp.Summary.Total != 1 {
		t.Errorf("summary.total: got %d, want 1", resp.Summary.Total)
	}
}

func TestGetBatchNotFound(t *testing.T) {
	h := newHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/checks/b_nope", nil)
	req.SetPathValue("id", "b_nope")
	w := httptest.NewRecorder()

	h.GetBatch(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	var resp api.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "batch_not_found" {
		t.Errorf("error code: got %q, want %q", resp.Error.Code, "batch_not_found")
	}
}

func TestGetBatchFound(t *testing.T) {
	// Create a batch first, then retrieve it.
	mock := &checker.MockChecker{Results: map[string]domain.CheckResult{
		"https://go.dev": {URL: "https://go.dev", StatusCode: 200, OK: true, LatencyMs: 5},
	}}
	st := store.NewMemoryStore()
	h := api.NewHandler(mock, st, noopLogger())

	// POST
	postBody := `{"urls":["https://go.dev"],"options":{"concurrency":1,"timeout_ms":500}}`
	postReq := httptest.NewRequest(http.MethodPost, "/v1/checks", bytes.NewBufferString(postBody))
	postW := httptest.NewRecorder()
	h.CreateBatch(postW, postReq)
	if postW.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d %s", postW.Code, postW.Body.String())
	}

	var created domain.Batch
	json.NewDecoder(postW.Body).Decode(&created)

	// GET
	getReq := httptest.NewRequest(http.MethodGet, "/v1/checks/"+created.ID, nil)
	getReq.SetPathValue("id", created.ID)
	getW := httptest.NewRecorder()
	h.GetBatch(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d", getW.Code)
	}
}

func TestCreateBatchValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty urls array", `{"urls":[]}`},
		{"no urls field", `{}`},
		{"invalid url scheme", `{"urls":["ftp://bad.com"]}`},
		{"concurrency out of range", `{"urls":["https://go.dev"],"options":{"concurrency":99}}`},
		{"timeout_ms out of range", `{"urls":["https://go.dev"],"options":{"timeout_ms":50}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(nil)
			req := httptest.NewRequest(http.MethodPost, "/v1/checks", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			h.CreateBatch(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	h := newHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
