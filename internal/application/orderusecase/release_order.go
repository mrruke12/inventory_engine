package orderusecase

import (
	"context"
	"log/slog"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	initialStatus, updated, found, err := s.orderRepo.ReleaseOrder(ctx, orderID)

	if !found {
		return false, order.ErrOrderNotFound // no-op with error
	}

	if err != nil {
		return false, err
	}

	if found && initialStatus == nil { // this is actually impossible since found = true ensures the row was found with the status itself
		slog.Error(
			"Impossible condition met in ReleaseOrder",
			"found", found,
			"initialStatus", initialStatus,
		)
		return false, application.ErrImpossibleCondition
	}

	switch *initialStatus {
	case order.StatusPending, order.StatusReleased, order.StatusExpired:
		return updated, nil // updated if the order is pending or no-op if the order is already released
	case order.StatusCanceled:
		return false, order.ErrOrderCanceled // cannot release canceled order
	case order.StatusCompleted:
		return false, order.ErrOrderCompleted // cannot release completed order, only cancel
	default:
		slog.Warn("Unknown order status",
			"orderID", orderID,
			"status", initialStatus,
		)
		return false, nil // not an error since transition can happen only if the order is pending
	}
}
