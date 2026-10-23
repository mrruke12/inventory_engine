package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func (s *Service) CreateOrder(ctx context.Context, orderItems []order.OrderItem) (uuid.UUID, error) {
	o := order.NewOrder()
	items := order.ItemsFromOrderItems(o.ID, orderItems)

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		orderRepo := rp.OrderRepository()

		if err := orderRepo.CreateOrder(ctx, o); err != nil {
			return err
		}

		return orderRepo.CreateAndReserveOrderItems(ctx, items)
	})

	if err != nil {
		return uuid.Nil(), err
	}

	return o.ID, err
}
