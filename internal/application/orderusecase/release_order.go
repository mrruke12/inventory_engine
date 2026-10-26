package orderusecase

import (
	"context"
	"log/slog"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	initialStatus, updated, err := s.orderRepo.ReleaseOrder(ctx, orderID)

	if err != nil {
		return false, err
	}

	switch initialStatus {
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
		return updated, nil // not an error since transition can happen only from pending status
	}
}
