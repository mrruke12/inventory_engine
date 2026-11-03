package mocks

import (
	"context"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
)

type Idempotency struct {
	GetOrCreateFunc func(context.Context, *ports.IdempotencyKey) (*ports.CachedResult, error)
}

func (i *Idempotency) GetOrCreate(ctx context.Context, key *ports.IdempotencyKey) (*ports.CachedResult, error) {
	if i.GetOrCreateFunc == nil {
		panic("GetOrCreateFunc is not set")
	}
	return i.GetOrCreateFunc(ctx, key)
}
