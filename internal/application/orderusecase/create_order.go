package orderusecase

import (
	"context"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CreateOrder(ctx context.Context, orderID string, orderItems []order.OrderItem) error {
	o := order.NewOrder(orderID)
	items := order.ItemsFromOrderItems(orderID, orderItems)

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		orderRepo := rp.OrderRepository()

		if err := orderRepo.CreateOrder(ctx, o); err != nil {
			return err
		}

		return orderRepo.CreateAndReserveOrderItems(ctx, items)
	})

	if err != nil {
		return err
	}

	return nil
}
