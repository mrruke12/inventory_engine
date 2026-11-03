package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CompleteOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	result := order.PendingOrderUpdateResult{}

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		var err error
		repo := rp.OrderRepository()
		result, err = repo.CompleteOrder(ctx, orderID)

		if err != nil {
			return err
		}

		if err = result.Validate(); err != nil {
			return err
		}

		if result.Found && *result.InitialStatus == order.StatusPending && result.Updated && result.Expired {
			return application.ErrMustNotUpdate
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound
	}

	status := *result.InitialStatus

	if status == order.StatusPending {
		if !result.Expired && !result.Updated {
			return false, application.ErrMustUpdate
		}

		if result.Expired {
			return false, order.ErrOrderExpired
		}
	}

	switch *result.InitialStatus {
	case order.StatusPending, order.StatusCompleted:
		return result.Updated, nil
	default:
		return result.Updated, order.ErrOrderByStatus(status)
	}
}
