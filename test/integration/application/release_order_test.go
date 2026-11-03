//go:build integration

package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
