package pool_test

import (
	"context"
	"testing"
	"time"

	"github.com/alhai/urlwatch/internal/checker"
	"github.com/alhai/urlwatch/internal/domain"
	"github.com/alhai/urlwatch/internal/pool"
)

func makeMock(results map[string]domain.CheckResult) *checker.MockChecker {
	return &checker.MockChecker{Results: results}
}

func TestRunAllURLsProcessed(t *testing.T) {
	mock := makeMock(map[string]domain.CheckResult{
		"https://go.dev":   {URL: "https://go.dev", StatusCode: 200, OK: true, LatencyMs: 10},
		"https://fail.dev": {URL: "https://fail.dev", OK: false, Error: "connection refused"},
	})

	results := pool.Run(context.Background(), mock, []string{"https://go.dev", "https://fail.dev"}, pool.Options{
		Concurrency: 2,
		TimeoutMs:   1000,
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRunBoundedConcurrency(t *testing.T) {
	// 10 URLs with concurrency=2: all 10 must be returned.
	urls := make([]string, 10)
	results := map[string]domain.CheckResult{}
	for i := range urls {
		u := "https://url" + string(rune('0'+i)) + ".dev"
		urls[i] = u
		results[u] = domain.CheckResult{URL: u, OK: true, StatusCode: 200}
	}

	got := pool.Run(context.Background(), makeMock(results), urls, pool.Options{
		Concurrency: 2,
		TimeoutMs:   1000,
	})
	if len(got) != 10 {
		t.Errorf("expected 10 results, got %d", len(got))
	}
}

func TestRunCancellationStopsWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before Run is called

	mock := makeMock(map[string]domain.CheckResult{
		"https://go.dev": {URL: "https://go.dev", OK: true, StatusCode: 200},
	})

	results := pool.Run(ctx, mock, []string{"https://go.dev"}, pool.Options{
		Concurrency: 1,
		TimeoutMs:   1000,
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Error("expected result to be not OK when context is already cancelled")
	}
}

func TestRunTimeoutInterruptsSlowChecker(t *testing.T) {
	slow := &slowChecker{delay: 300 * time.Millisecond}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results := pool.Run(ctx, slow, []string{"https://slow.dev"}, pool.Options{
		Concurrency: 1,
		TimeoutMs:   5000,
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Error("expected result to be not OK after context timeout")
	}
}

// slowChecker blocks until the context is done or the delay elapses.
type slowChecker struct{ delay time.Duration }

func (s *slowChecker) Check(ctx context.Context, url string) domain.CheckResult {
	select {
	case <-ctx.Done():
		return domain.CheckResult{URL: url, OK: false, Error: ctx.Err().Error()}
	case <-time.After(s.delay):
		return domain.CheckResult{URL: url, OK: true, StatusCode: 200}
	}
}
