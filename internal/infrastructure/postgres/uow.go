package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
)

type UnitOfWork[T any] struct {
	pool            *pgxpool.Pool
	providerFactory func(Querier) T
}

func NewUnitOfWork[T any](pool *pgxpool.Pool, providerFactory func(Querier) T) *UnitOfWork[T] {
	return &UnitOfWork[T]{
		pool:            pool,
		providerFactory: providerFactory,
	}
}

func (u *UnitOfWork[T]) Do(ctx context.Context, fn func(context.Context, T) error) error {
	tx, err := u.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	provider := u.providerFactory(tx)

	err = fn(ctx, provider)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func NewOrderUsecaseUnitOfWork(pool *pgxpool.Pool, providerFactory func(Querier) ports.RepositoryProvider) ports.UnitOfWork {
	return &UnitOfWork[ports.RepositoryProvider]{
		pool:            pool,
		providerFactory: providerFactory,
	}
}
