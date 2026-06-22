package domain

import "time"

type CheckResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	OK         bool   `json:"ok"`
	LatencyMs  int64  `json:"latency_ms"`
	Error      string `json:"error,omitempty"`
}

type Summary struct {
	Total      int   `json:"total"`
	Up         int   `json:"up"`
	Down       int   `json:"down"`
	DurationMs int64 `json:"duration_ms"`
}

type Batch struct {
	ID        string        `json:"batch_id"`
	CreatedAt time.Time     `json:"created_at"`
	Results   []CheckResult `json:"results"`
	Summary   Summary       `json:"summary"`
}

func NewSummary(results []CheckResult, durationMs int64) Summary {
	s := Summary{Total: len(results), DurationMs: durationMs}
	for _, r := range results {
		if r.OK {
			s.Up++
		} else {
			s.Down++
		}
	}
	return s
}
