//go:build integration

package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests ReleaseOrder response for different order statuses
func TestReleaseOrder_Existing(t *testing.T) {
	tests := []struct {
		name            string
		status          order.Status
		expectedUpdated bool
		expectedError   error
	}{
		{"release pending", order.StatusPending, true, nil},
		{"release expired", order.StatusExpired, false, nil},
		{"release released", order.StatusReleased, false, nil},
		{"release canceled", order.StatusCanceled, false, order.ErrOrderCanceled},
		{"release completed", order.StatusCompleted, false, order.ErrOrderCompleted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			integration.CleanupPostgres(t, pool)

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
			defer cancel()

			productID := requireTestProduct(ctx, t, 1)
			orderID, err := createOrder(ctx, productID, uuid.New().String(), 1)

			require.NoError(t, err)

			sql := `
				update orders set status = $1
				where id = $2
			`

			_, err = pool.Exec(ctx, sql, tc.status, orderID)
			require.NoError(t, err)

			updated, err := orderService.ReleaseOrder(ctx, orderID)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedUpdated, updated)
		})
	}
}

// Tests that ReleaseOrder returns order.ErrOrderNotFound
// for not exitsting order
func TestReleaseOrder_NotExisting(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	updated, err := orderService.ReleaseOrder(ctx, uuid.NewV7())

	assert.ErrorIs(t, err, order.ErrOrderNotFound)
	assert.False(t, updated)
}

// Tests if concurrent ReleaseOrder calls are idempotent
// i.e. have no double-release effect.
func TestReleaseOrder_Concurrent(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	orders := 100
	routines := 10
	productID := requireTestProduct(ctx, t, 1)

	var releasedCount, noopCount atomic.Int32

	for range orders {
		var wg sync.WaitGroup

		orderID, err := createOrder(ctx, productID, uuid.New().String(), 1)
		require.NoError(t, err)

		for range routines {
			wg.Add(1)

			go func() {
				defer wg.Done()

				updated, err := orderService.ReleaseOrder(ctx, orderID)

				if err != nil {
					t.Errorf("ReleaseOrder: %v", err)
					return
				}

				if updated {
					releasedCount.Add(1)
				} else {
					noopCount.Add(1)
				}
			}()
		}

		wg.Wait()

		require.NoError(t, ctx.Err(), "test context expired before completion")
	}

	stock, _ := requireProductStockAndOrders(ctx, t, productID)
	assert.Equal(t, 1, stock, "stock mismatch: double-release happened")

	assert.Equal(t, int32(orders), releasedCount.Load(), "released count")
	assert.Equal(t, int32(orders*(routines-1)), noopCount.Load(), "no-op count")
}
