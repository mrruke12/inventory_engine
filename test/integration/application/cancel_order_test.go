//go:build integration

package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelOrder_Concurrent(t *testing.T) {
	test.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	orders := 100
	routines := 10
	productID := requireTestProduct(ctx, t, orders)

	var canceledCount, noopCount atomic.Int32

	for range orders {
		var wg sync.WaitGroup

		orderID, err := createOrder(ctx, productID, uuid.New().String(), 1)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `update orders set status = 'completed' where id = $1`, orderID)
		require.NoError(t, err)

		for range routines {
			wg.Add(1)

			go func() {
				defer wg.Done()

				updated, err := orderService.CancelOrder(ctx, orderID)

				if err != nil {
					t.Errorf("CancelOrder: %v", err)
					return
				}

				if updated {
					canceledCount.Add(1)
				} else {
					noopCount.Add(1)
				}
			}()
		}

		wg.Wait()

		require.NoError(t, ctx.Err(), "test context expired before completion")
	}

	stock, _ := requireProductStockAndOrders(ctx, t, productID)
	assert.Equal(t, orders, stock, "stock mismatch: double-cancel happened")

	assert.Equal(t, int32(orders), canceledCount.Load(), "canceled count")
	assert.Equal(t, int32(orders*(routines-1)), noopCount.Load(), "no-op count")
}
