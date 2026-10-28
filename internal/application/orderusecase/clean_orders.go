package orderusecase

import (
	"context"
	"log/slog"
)

func (s *Service) CleanOrders(ctx context.Context, batchSize int) error {
	updated, err := s.orderRepo.BatchCancelExpiredOrders(ctx, batchSize)

	if err != nil {
		return err
	}

	slog.Info(
		"Canceled expired orders",
		"updated", updated,
	)

	return nil
}
