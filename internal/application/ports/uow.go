package ports

import "context"

type UnitOfWork interface {
	Do(context.Context, func(context.Context, RepositoryProvider) error) error
}
