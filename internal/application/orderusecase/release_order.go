package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	result := order.PendingOrderUpdateResult{}

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		var err error
		repo := rp.OrderRepository()
		result, err = repo.ReleaseOrder(ctx, orderID)

		if err != nil {
			return err
		}

		return result.Validate()
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound // no-op with error
	}

	status := *result.InitialStatus

	if status == order.StatusPending {
		if result.Expired {
			return false, order.ErrOrderExpired
		}

		if !result.Updated {
			return false, application.ErrMustUpdate
		}
	}

	switch status {
	case order.StatusPending, order.StatusReleased:
		return result.Updated, nil // updated if the order is pending or no-op if the order is already released
	default:
		return result.Updated, order.ErrOrderByStatus(status)
	}
}
