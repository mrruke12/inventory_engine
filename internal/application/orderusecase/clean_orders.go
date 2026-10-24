package orderusecase

import (
	"context"
	"log/slog"
)

const batchSize = 1

func (s *Service) CleanOrders(ctx context.Context) error {
	updated, err := s.orderRepo.BatchCancelExpiredOrders(ctx, batchSize)

	if err != nil {
		return err
	}

	slog.Info(
		"Canceled expired orders",
		"expired_orders_count", updated[0],
		"expired_items_count", updated[1],
		"updated_orders_count", updated[2],
		"updated_products_count", updated[3],
	)

	return nil
}
