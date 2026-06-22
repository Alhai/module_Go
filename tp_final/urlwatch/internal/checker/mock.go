package checker

import (
	"context"

	"github.com/alhai/urlwatch/internal/domain"
)

// MockChecker is a deterministic Checker for tests — no network calls.
type MockChecker struct {
	Results map[string]domain.CheckResult
}

func (m *MockChecker) Check(_ context.Context, url string) domain.CheckResult {
	if r, ok := m.Results[url]; ok {
		return r
	}
	return domain.CheckResult{URL: url, OK: false, Error: "mock: url not configured"}
}
