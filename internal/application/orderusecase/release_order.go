package orderusecase

import (
	"context"
	"log/slog"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	result := order.OrderUpdateResult{}

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		var err error
		repo := rp.OrderRepository()
		result, err = repo.ReleaseOrder(ctx, orderID)
		return err
	})

	if err != nil {
		return false, err
	}

	if !result.Found {
		return false, order.ErrOrderNotFound // no-op with error
	}

	if result.InitialStatus == nil { // this is actually impossible since found = true ensures the row was found with the status itself
		slog.Error(
			"Impossible condition met in ReleaseOrder",
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
	case order.StatusPending, order.StatusReleased, order.StatusExpired:
		return result.Updated, nil // updated if the order is pending or no-op if the order is already released
	case order.StatusCanceled:
		return false, order.ErrOrderCanceled // cannot release canceled order
	case order.StatusCompleted:
		return false, order.ErrOrderCompleted // cannot release completed order, only cancel
	default:
		slog.Warn("Unknown order status",
			"orderID", orderID,
			"status", result.InitialStatus,
		)
		return false, nil // not an error since transition can happen only if the order is pending
	}
}
