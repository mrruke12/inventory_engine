package orderusecase

import (
	"context"
	"fmt"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/pkg/aggr"
)

func (s *Service) CreateOrder(ctx context.Context, orderItems []order.OrderItem) (uuid.UUID, error) {
	o := order.NewOrder()
	items := order.ItemsFromOrderItems(o.ID, orderItems)
	products := aggr.Map(items, func(v order.Item, i int) uuid.UUID { return v.ProductID })
	deltas := aggr.Map(items, func(v order.Item, i int) int { return v.Quantity })
	fmt.Println(products, deltas)

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		orderRepo := rp.OrderRepository()

		if err := orderRepo.CreateOrder(ctx, o); err != nil {
			return err
		}

		return orderRepo.CreateAndReserveOrderItems(ctx, products, deltas)
	})

	if err != nil {
		return uuid.Nil(), err
	}

	return o.ID, err
}
