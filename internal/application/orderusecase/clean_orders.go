package orderusecase

import (
	"context"
	"log/slog"
)

const batchSize = 100

func (s *Service) CleanOrders(ctx context.Context) error {
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
