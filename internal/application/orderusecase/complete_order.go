package orderusecase

import (
	"context"
	"log/slog"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CompleteOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	result := order.OrderUpdateResult{}

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		var err error
		repo := rp.OrderRepository()
		result, err = repo.CompleteOrder(ctx, orderID)
		return err
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound
	}

	if result.InitialStatus == nil { // this is actually impossible since found = true ensures the row was found with the status itself
		slog.Error(
			"Impossible condition met in CompleteOrder",
			"found", result.Found,
			"initialStatus", result.InitialStatus,
		)
		return false, application.ErrImpossibleCondition
	}

	if result.Expired {
		return false, order.ErrOrderExpired
	}

	if *result.InitialStatus == order.StatusPending && !result.Updated {
		return false, application.ErrMustUpdate
	}

	switch *result.InitialStatus {
	case order.StatusPending, order.StatusCompleted:
		return result.Updated, nil
	case order.StatusExpired:
		return false, order.ErrOrderExpired
	case order.StatusReleased:
		return false, order.ErrOrderReleased
	case order.StatusCanceled:
		return false, order.ErrOrderCanceled
	default:
		slog.Warn("Unknown order status",
			"orderID", orderID,
			"status", result.InitialStatus,
		)
		return result.Updated, nil
	}
}
