package domain_test

import (
	"testing"

	"github.com/alhai/urlwatch/internal/domain"
)

func TestNewSummary(t *testing.T) {
	tests := []struct {
		name       string
		results    []domain.CheckResult
		durationMs int64
		wantUp     int
		wantDown   int
		wantTotal  int
	}{
		{
			name: "mixed results",
			results: []domain.CheckResult{
				{URL: "https://a.dev", OK: true},
				{URL: "https://b.dev", OK: false},
				{URL: "https://c.dev", OK: true},
			},
			durationMs: 500,
			wantTotal:  3,
			wantUp:     2,
			wantDown:   1,
		},
		{
			name:       "empty",
			results:    []domain.CheckResult{},
			durationMs: 0,
			wantTotal:  0,
			wantUp:     0,
			wantDown:   0,
		},
		{
			name: "all up",
			results: []domain.CheckResult{
				{URL: "https://a.dev", OK: true},
				{URL: "https://b.dev", OK: true},
			},
			durationMs: 100,
			wantTotal:  2,
			wantUp:     2,
			wantDown:   0,
		},
		{
			name: "all down",
			results: []domain.CheckResult{
				{URL: "https://a.dev", OK: false},
			},
			durationMs: 50,
			wantTotal:  1,
			wantUp:     0,
			wantDown:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := domain.NewSummary(tt.results, tt.durationMs)
			if s.Total != tt.wantTotal {
				t.Errorf("Total: got %d, want %d", s.Total, tt.wantTotal)
			}
			if s.Up != tt.wantUp {
				t.Errorf("Up: got %d, want %d", s.Up, tt.wantUp)
			}
			if s.Down != tt.wantDown {
				t.Errorf("Down: got %d, want %d", s.Down, tt.wantDown)
			}
			if s.DurationMs != tt.durationMs {
				t.Errorf("DurationMs: got %d, want %d", s.DurationMs, tt.durationMs)
			}
		})
	}
}
