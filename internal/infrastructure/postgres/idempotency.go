package postgres

import (
	"context"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
)

type Idempotency struct {
	pool Querier
}

func NewIdempotency(pool Querier) ports.Idempotency {
	return &Idempotency{
		pool: pool,
	}
}

func (i *Idempotency) GetOrCreate(ctx context.Context, key *ports.IdempotencyKey) (*ports.CachedResult, error) {
	sql := `
		insert into idempotency_keys (key, scope, request_hash, result)
		values ($1, $2, $3, $4)
		on conflict (key, scope) do update set key = excluded.key
		returning request_hash, result, (xmax != 0) as found
	`

	var hash []byte

	cached := &ports.CachedResult{}

	row := i.pool.QueryRow(ctx, sql, key.Key, key.Scope, key.RequestHash[:], key.Result)

	if err := row.Scan(&hash, &cached.Result, &cached.Found); err != nil {
		return cached, err
	}

	copy(cached.RequestHash[:], hash)
	return cached, nil
}
