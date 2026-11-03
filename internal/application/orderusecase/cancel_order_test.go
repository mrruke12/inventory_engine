package orderusecase

import (
	"context"
	"testing"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/application/orderusecase/mocks"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/stretchr/testify/assert"
)

func TestCancelOrder(t *testing.T) {
	ctx := context.Background()

	statusToPointer := func(status order.Status) *order.Status {
		return &status
	}

	tests := []struct {
		name    string
		result  order.CancelOrderResult
		updated bool
		err     error
	}{
		{
			name: "found pending",
			result: order.CancelOrderResult{
				InitialStatus: statusToPointer(order.StatusPending),
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderPending,
		},
		{
			name: "found expired",
			result: order.CancelOrderResult{
				InitialStatus: statusToPointer(order.StatusExpired),
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderExpired,
		},
		{
			name: "found completed",
			result: order.CancelOrderResult{
				InitialStatus: statusToPointer(order.StatusCompleted),
				Updated:       true,
				Found:         true,
			},
			updated: true,
			err:     nil,
		},
		{
			name: "found canceled",
			result: order.CancelOrderResult{
				InitialStatus: statusToPointer(order.StatusCanceled),
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     nil,
		},
		{
			name: "not found",
			result: order.CancelOrderResult{
				InitialStatus: nil,
				Updated:       false,
				Found:         false,
			},
			updated: false,
			err:     order.ErrOrderNotFound,
		},
		{
			name: "found completed but didn't update",
			result: order.CancelOrderResult{
				InitialStatus: statusToPointer(order.StatusCompleted),
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     application.ErrMustUpdate,
		},
		{
			name: "found but initial status is nil",
			result: order.CancelOrderResult{
				InitialStatus: nil,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     application.ErrImpossibleCondition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			provider := &mocks.RepositoryProvider{
				OrderRepo: &mocks.OrderRepository{
					CancelOrderFunc: func(ctx context.Context, orderID uuid.UUID) (result order.CancelOrderResult, err error) {
						return tc.result, nil
					},
				},
			}

			s := NewService(
				&mocks.UnitOfWork{Provider: provider},
				&mocks.OrderRepository{},
			)

			updated, err := s.CancelOrder(ctx, uuid.Nil())
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.updated, updated)
		})

	}
}
