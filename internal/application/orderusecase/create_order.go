package orderusecase

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type CreateOrderCommand struct {
	IdempotencyKey string
	OrderItems     []order.OrderItem
}

func (s *Service) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (uuid.UUID, error) {
	result := uuid.Nil()

	o := order.NewOrder()
	items := order.ItemsFromOrderItems(o.ID, cmd.OrderItems)

	err := s.uow.Do(ctx, func(ctx context.Context, rp ports.RepositoryProvider) error {
		idempKey, err := ports.NewIdempotencyKey(cmd.IdempotencyKey, "create_order", cmd, o.ID)

		if err != nil {
			return err
		}

		cached, err := rp.Idempotency().GetOrCreate(ctx, idempKey)

		if err != nil {
			return err
		}

		if cached.Found {
			if cached.RequestHash != idempKey.RequestHash {
				return application.ErrIdempotencyKeyReused
			}

			return cached.Unmarshal(&result)
		}

		orderRepo := rp.OrderRepository()

		if err := orderRepo.CreateOrder(ctx, o); err != nil {
			return err
		}

		result = o.ID

		return orderRepo.CreateAndReserveOrderItems(ctx, o.ID, items)
	})

	return result, err
}
