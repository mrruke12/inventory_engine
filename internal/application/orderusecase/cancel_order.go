package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	var result order.CancelOrderResult

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		repo := rp.OrderRepository()
		var err error
		result, err = repo.CancelOrder(ctx, orderID)
		return err
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound
	}

	if err := requireInitialStatus(result.InitialStatus); err != nil {
		return false, err
	}

	status := *result.InitialStatus

	switch status {
	case order.StatusCompleted, order.StatusCanceled:
		return result.Updated, nil
	default:
		return result.Updated, order.ErrOrderByStatus(status)
	}
}
