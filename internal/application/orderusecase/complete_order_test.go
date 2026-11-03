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

func TestCompleteOrder(t *testing.T) {
	ctx := context.Background()

	statusToPointer := func(status order.Status) *order.Status {
		return &status
	}

	tests := []struct {
		name    string
		result  order.PendingOrderUpdateResult
		updated bool
		err     error
	}{
		{
			name: "found pending",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusPending),
				Expired:       false,
				Updated:       true,
				Found:         true,
			},
			updated: true,
			err:     nil,
		},
		{
			name: "found pending but expired",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusPending),
				Expired:       true,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderExpired,
		},
		{
			name: "found expired",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusExpired),
				Expired:       true,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderExpired,
		},
		{
			name: "found completed",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusCompleted),
				Expired:       false,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     nil,
		},
		{
			name: "found completed but expired",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusCompleted),
				Expired:       true,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     nil,
		},
		{
			name: "found canceled",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusCanceled),
				Expired:       false,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderCanceled,
		},
		{
			name: "found canceled but expired",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusCanceled),
				Expired:       true,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     order.ErrOrderCanceled,
		},
		{
			name: "not found",
			result: order.PendingOrderUpdateResult{
				InitialStatus: nil,
				Expired:       false,
				Updated:       false,
				Found:         false,
			},
			updated: false,
			err:     order.ErrOrderNotFound,
		},
		{
			name: "found pending but didn't update",
			result: order.PendingOrderUpdateResult{
				InitialStatus: statusToPointer(order.StatusPending),
				Expired:       false,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     application.ErrMustUpdate,
		},
		{
			name: "found but initial status is nil",
			result: order.PendingOrderUpdateResult{
				InitialStatus: nil,
				Expired:       false,
				Updated:       false,
				Found:         true,
			},
			updated: false,
			err:     application.ErrImpossibleCondition,
		},
		{
			name: "not found but updated",
			result: order.PendingOrderUpdateResult{
				InitialStatus: nil,
				Expired:       false,
				Updated:       true,
				Found:         false,
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
					CompleteOrderFunc: func(ctx context.Context, orderID uuid.UUID) (result order.PendingOrderUpdateResult, err error) {
						return tc.result, nil
					},
				},
			}

			s := NewService(
				&mocks.UnitOfWork{Provider: provider},
				&mocks.OrderRepository{},
			)

			updated, err := s.CompleteOrder(ctx, uuid.Nil())
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.updated, updated)
		})

	}
}
