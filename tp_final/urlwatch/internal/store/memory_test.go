package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alhai/urlwatch/internal/domain"
	"github.com/alhai/urlwatch/internal/store"
)

func TestMemoryStoreSaveAndGet(t *testing.T) {
	s := store.NewMemoryStore()
	ctx := context.Background()

	batch := domain.Batch{
		ID:        "b_test01",
		CreatedAt: time.Now(),
		Results: []domain.CheckResult{
			{URL: "https://go.dev", OK: true, StatusCode: 200, LatencyMs: 50},
		},
		Summary: domain.Summary{Total: 1, Up: 1, Down: 0, DurationMs: 50},
	}

	if err := s.Save(ctx, batch); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Get(ctx, "b_test01")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != batch.ID {
		t.Errorf("ID: got %q, want %q", got.ID, batch.ID)
	}
	if len(got.Results) != 1 {
		t.Errorf("Results len: got %d, want 1", len(got.Results))
	}
}

func TestMemoryStoreNotFound(t *testing.T) {
	s := store.NewMemoryStore()
	_, err := s.Get(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrBatchNotFound) {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestMemoryStoreOverwrite(t *testing.T) {
	s := store.NewMemoryStore()
	ctx := context.Background()

	b1 := domain.Batch{ID: "b_x", Results: []domain.CheckResult{{URL: "https://a.dev", OK: true}}}
	b2 := domain.Batch{ID: "b_x", Results: []domain.CheckResult{{URL: "https://b.dev", OK: false}}}

	_ = s.Save(ctx, b1)
	_ = s.Save(ctx, b2)

	got, _ := s.Get(ctx, "b_x")
	if got.Results[0].URL != "https://b.dev" {
		t.Errorf("expected overwritten URL, got %q", got.Results[0].URL)
	}
}
