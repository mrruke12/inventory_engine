package orderusecase

import (
	"context"
	"log/slog"
)

const batchSize = 100

func (s *Service) CleanOrders(ctx context.Context) error {
	updated, err := s.orderRepo.BatchMarkExpiredOrders(ctx, batchSize)

	if err != nil {
		return err
	}

	slog.Info("Marked expired orders", "updated", updated)

	return nil
}
