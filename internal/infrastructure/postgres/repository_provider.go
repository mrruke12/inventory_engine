package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type RepositoryProvider struct {
	tx Querier
}

func NewRepositoryProvider(tx Querier) ports.RepositoryProvider {
	return &RepositoryProvider{
		tx: tx,
	}
}

func (p *RepositoryProvider) OrderRepository() order.Repository {
	return NewOrderRepository(p.tx)
}
