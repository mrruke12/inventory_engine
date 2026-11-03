package mocks

import (
	"context"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
)

type UnitOfWork struct {
	Provider ports.RepositoryProvider
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context, ports.RepositoryProvider) error) error {
	return fn(ctx, u.Provider)
}
