package checker

import (
	"context"
	"net/http"
	"time"

	"github.com/alhai/urlwatch/internal/domain"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker() *HTTPChecker {
	return &HTTPChecker{client: &http.Client{}}
}

func (c *HTTPChecker) Check(ctx context.Context, url string) domain.CheckResult {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.CheckResult{
			URL:       url,
			OK:        false,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     err.Error(),
		}
	}

	resp, err := c.client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return domain.CheckResult{
			URL:       url,
			OK:        false,
			LatencyMs: latency,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	return domain.CheckResult{
		URL:        url,
		StatusCode: resp.StatusCode,
		OK:         ok,
		LatencyMs:  latency,
	}
}
