package orderusecase

import (
	"context"
	"encoding/json"
	"testing"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/orderusecase/mocks"
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder_Idempotency(t *testing.T) {
	ctx := context.Background()

	cachedID := uuid.MustParse("00000000-0000-0000-0000-000000000201")
	cachedResult, err := json.Marshal(cachedID)
	require.NoError(t, err)

	tests := []struct {
		name string
		err  error
		id   uuid.UUID
		mock func(ctx context.Context, key *ports.IdempotencyKey) (*ports.CachedResult, error)
	}{
		{
			"request cached",
			nil,
			cachedID,
			func(ctx context.Context, key *ports.IdempotencyKey) (*ports.CachedResult, error) {
				return &ports.CachedResult{
					RequestHash: key.RequestHash,
					Result:      cachedResult,
				}, nil
			},
		},
		{
			"request hash mismatch",
			application.ErrIdempotencyKeyReused,
			uuid.Nil(),
			func(ctx context.Context, key *ports.IdempotencyKey) (*ports.CachedResult, error) {
				return &ports.CachedResult{
					RequestHash: [32]byte{},
					Result:      cachedResult,
				}, nil
			},
		},
	}

	for _, tc := range tests {
		provider := mocks.RepositoryProvider{
			Idemp: &mocks.Idempotency{
				GetOrCreateFunc: tc.mock,
			},
		}

		s := NewService(
			&mocks.UnitOfWork{Provider: &provider},
			&mocks.OrderRepository{},
		)

		cmd := CreateOrderCommand{
			IdempotencyKey: "",
			OrderItems:     []order.OrderItem{{}},
		}

		id, err := s.CreateOrder(ctx, cmd)

		assert.ErrorIs(t, err, tc.err)
		assert.Equal(t, tc.id.String(), id.String())
	}
}

func TestCreateOrder_EmptyOrder(t *testing.T) {
	s := NewService(
		&mocks.UnitOfWork{},
		&mocks.OrderRepository{},
	)

	var cmd CreateOrderCommand
	_, err := s.CreateOrder(context.Background(), cmd)
	require.ErrorIs(t, err, order.ErrOrderEmpty)
}

func TestCreateOrder(t *testing.T) {
	provider := &mocks.RepositoryProvider{
		Idemp: &mocks.Idempotency{
			GetOrCreateFunc: func(ctx context.Context, ik *ports.IdempotencyKey) (*ports.CachedResult, error) {
				return nil, nil
			},
		},
		OrderRepo: &mocks.OrderRepository{
			CreateOrderFunc:                func(ctx context.Context, order *order.Order) error { return nil },
			CreateAndReserveOrderItemsFunc: func(ctx context.Context, orderID uuid.UUID, items []order.Item) error { return nil },
		},
	}

	s := NewService(
		&mocks.UnitOfWork{Provider: provider},
		&mocks.OrderRepository{},
	)

	id, err := s.CreateOrder(
		context.Background(),
		CreateOrderCommand{OrderItems: []order.OrderItem{{}}},
	)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil().String(), id.String())
}
