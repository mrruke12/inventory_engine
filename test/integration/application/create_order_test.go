//go:build integration

package application

import (
	"context"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/domains/product"
	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests common CreateOrder cases
func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		initStock      int
		amount         int
		expectedStock  int
		expectedOrders int
		expectedError  error
	}{
		{"enough product", 1, 1, 0, 1, nil},
		{"enough product", 100, 50, 50, 1, nil},
		{"not enough product", 1, 2, 1, 0, product.ErrNotEnoughProduct},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			integration.CleanupPostgres(t, pool)

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
			defer cancel()

			productID := requireTestProduct(ctx, t, tc.initStock)

			_, err := createOrder(ctx, productID, uuid.New().String(), tc.amount)

			require.ErrorIs(t, err, tc.expectedError)

			stock, orders := requireProductStockAndOrders(ctx, t, productID)

			assert.Equalf(t, tc.expectedStock, stock, "expected stock of %d, got %d", tc.expectedStock, stock)
			assert.Equalf(t, tc.expectedOrders, orders, "expected %d orders, got %d", tc.expectedOrders, orders)
		})
	}
}

// Tests that concurrent CreateOrder calls don't cause race condition
// for product stock
//
// Note: acquires min(10, pool.Config().MaxConns) pool connections
func TestCreateOrder_Concurrent(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	initStock := 50
	productID := requireTestProduct(ctx, t, initStock)

	workers := int(min(10, pool.Config().MaxConns))
	require.GreaterOrEqual(t, workers, 2, "need ≥2 pool connections for concurrency test")

	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)

		go func() {

			defer wg.Done()

			for range initStock {
				if _, err := createOrder(ctx, productID, uuid.New().String(), 1); err != nil {
					assert.ErrorIs(t, err, product.ErrNotEnoughProduct)
				}
			}
		}()
	}

	wg.Wait()

	stock, orders := requireProductStockAndOrders(ctx, t, productID)

	assert.Equalf(t, 0, stock, "expected stock of 0, got %d", stock)
	assert.Equalf(t, initStock, orders, "expected %d orders, got %d", initStock, orders)
}

// Tests if CreateOrder calls with the same idempotency key do not
// create 2 orders, but 1
func TestCreateOrder_Idempotency(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	productID := requireTestProduct(ctx, t, 1)
	idempotencyKey := uuid.New().String()

	id, err := createOrder(ctx, productID, idempotencyKey, 1)
	require.NoError(t, err)

	id2, err := createOrder(ctx, productID, idempotencyKey, 1)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	_, err = createOrder(ctx, productID, idempotencyKey, 2)
	assert.ErrorIs(t, err, application.ErrIdempotencyKeyReused)

	_, err = createOrder(ctx, uuid.NewV7(), idempotencyKey, 1)
	assert.ErrorIs(t, err, application.ErrIdempotencyKeyReused)
}

// Tests that concurrent CreateOrder calls with the same idempotency key
// create exactly one order and return the same ID to every caller.
func TestCreateOrder_IdempotencyConcurrent(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	productID := requireTestProduct(ctx, t, 10)
	idempotencyKey := uuid.New().String()

	workers := int(min(10, pool.Config().MaxConns))
	require.GreaterOrEqual(t, workers, 2, "need ≥2 pool connections for concurrency test")

	var (
		wg    sync.WaitGroup
		start = make(chan struct{})

		mu  sync.Mutex
		ids = make([]uuid.UUID, 0, workers)
	)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			id, err := createOrder(ctx, productID, idempotencyKey, 1)
			if err != nil {
				t.Errorf("createOrder: %v", err)
				return
			}

			mu.Lock()
			ids = append(ids, id)
			mu.Unlock()
		}()
	}

	close(start)
	wg.Wait()

	require.NoError(t, ctx.Err(), "test context expired")

	require.Len(t, ids, workers)
	for i := 1; i < len(ids); i++ {
		assert.Equalf(t, ids[0], ids[i], "idempotency violated: got different IDs %s and %s", ids[0], ids[i])
	}

	stock, orders := requireProductStockAndOrders(ctx, t, productID)
	assert.Equalf(t, 9, stock, "expected stock of 9, got %d", stock)
	assert.Equalf(t, 1, orders, "expected 1 order, got %d", orders)
}
