package pool

import (
	"context"
	"sync"
	"time"

	"github.com/alhai/urlwatch/internal/domain"
)

type Options struct {
	Concurrency int
	TimeoutMs   int64
}

// Run distributes urls across a bounded worker pool and collects results via
// fan-out/fan-in. The work channel is buffered (all URLs pre-loaded) so
// workers never block on sends. The results channel is also buffered to avoid
// blocking workers when the collector is slow. A single goroutine closes
// results after wg.Wait(), so the collector can range over it safely.
func Run(ctx context.Context, chkr domain.Checker, urls []string, opts Options) []domain.CheckResult {
	work := make(chan string, len(urls))
	for _, u := range urls {
		work <- u
	}
	close(work)

	results := make(chan domain.CheckResult, len(urls))

	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range work {
				// Fast-path: if the batch context is already cancelled, don't
				// start a new HTTP request — return an error result immediately.
				select {
				case <-ctx.Done():
					results <- domain.CheckResult{
						URL:   url,
						OK:    false,
						Error: ctx.Err().Error(),
					}
					continue
				default:
				}

				urlCtx, cancel := context.WithTimeout(ctx, time.Duration(opts.TimeoutMs)*time.Millisecond)
				r := chkr.Check(urlCtx, url)
				cancel()
				results <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]domain.CheckResult, 0, len(urls))
	for r := range results {
		out = append(out, r)
	}
	return out
}
