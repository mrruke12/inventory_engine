package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	var result order.CancelOrderResult

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		repo := rp.OrderRepository()

		var err error
		result, err = repo.CancelOrder(ctx, orderID)

		if err != nil {
			return err
		}

		if err := result.Validate(); err != nil {
			return err
		}

		if result.Found && result.InitialStatus != nil &&
			*result.InitialStatus == order.StatusCompleted && !result.Updated {
			return application.ErrMustUpdate
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound
	}

	switch *result.InitialStatus {
	case order.StatusCompleted, order.StatusCanceled:
		return result.Updated, nil
	default:
		return result.Updated, order.ErrOrderByStatus(*result.InitialStatus)
	}
}
