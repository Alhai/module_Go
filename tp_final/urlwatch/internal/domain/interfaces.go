package domain

import "context"

type Checker interface {
	Check(ctx context.Context, url string) CheckResult
}

type Store interface {
	Save(ctx context.Context, b Batch) error
	Get(ctx context.Context, id string) (Batch, error)
}
