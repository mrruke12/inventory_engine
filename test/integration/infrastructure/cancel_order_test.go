//go:build integration

package infrastructure

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelOrder_NotFound(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := repo.CancelOrder(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000201"))
	require.NoError(t, err)

	assert.False(t, res.Found)
	assert.False(t, res.Updated)
	assert.Nil(t, res.InitialStatus)
}

func TestCancelOrder_Completed_UpdatesAndRestocks(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000202")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusCompleted,
		time.Now().Add(5*time.Minute),
		initialStock, amount,
	)

	res, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res.InitialStatus)

	assert.Equal(t, order.StatusCanceled, getOrderStatus(t, pool, orderID))
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))
}

func TestCancelOrder_MultipleItems_RestocksSum(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000203")
	insertOrder(t, pool, orderID, order.StatusCompleted, time.Now().Add(5*time.Minute))

	p1 := insertProduct(t, pool, "cancel-p1", 5)
	p2 := insertProduct(t, pool, "cancel-p2", 7)

	insertOrderItem(t, pool, orderID, p1, 2)
	insertOrderItem(t, pool, orderID, p2, 4)

	res, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res.InitialStatus)

	assert.Equal(t, order.StatusCanceled, getOrderStatus(t, pool, orderID))
	assert.Equal(t, 5+2, getStock(t, pool, p1))
	assert.Equal(t, 7+4, getStock(t, pool, p2))
}

func TestCancelOrder_NonCompletedStatus_NoEffect(t *testing.T) {
	cases := []struct {
		name   string
		status order.Status
	}{
		{"pending", order.StatusPending},
		{"released", order.StatusReleased},
		{"expired", order.StatusExpired},
		{"canceled", order.StatusCanceled},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			integration.CleanupPostgres(t, pool)
			repo := postgres.NewOrderRepository(pool)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			orderID := uuid.New()
			const (
				initialStock = 10
				amount       = 3
			)

			productID := setupOrderWithItem(
				t, pool, orderID, tc.status,
				time.Now().Add(5*time.Minute),
				initialStock, amount,
			)

			res, err := repo.CancelOrder(ctx, orderID)
			require.NoError(t, err)

			assert.True(t, res.Found)
			assert.False(t, res.Updated)
			require.NotNil(t, res.InitialStatus)
			assert.Equal(t, tc.status, *res.InitialStatus)

			assert.Equal(t, tc.status, getOrderStatus(t, pool, orderID))
			assert.Equal(t, initialStock, getStock(t, pool, productID))
		})
	}
}

func TestCancelOrder_NonCompletedExpired_NoEffect(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000205")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusPending,
		time.Now().Add(-1*time.Minute),
		initialStock, amount,
	)

	res, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	assert.Equal(t, order.StatusPending, getOrderStatus(t, pool, orderID))
	assert.Equal(t, initialStock, getStock(t, pool, productID))
}

func TestCancelOrder_Idempotent(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000206")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusCompleted,
		time.Now().Add(5*time.Minute),
		initialStock, amount,
	)

	res1, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, res1.Found)
	assert.True(t, res1.Updated)
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))

	res2, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, res2.Found)
	assert.False(t, res2.Updated)
	require.NotNil(t, res2.InitialStatus)
	assert.Equal(t, order.StatusCanceled, *res2.InitialStatus)

	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))
}

func TestCancelOrder_NoItems_CancelsWithoutRestock(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000207")
	insertOrder(t, pool, orderID, order.StatusCompleted, time.Now().Add(5*time.Minute))

	res, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res.InitialStatus)

	assert.Equal(t, order.StatusCanceled, getOrderStatus(t, pool, orderID))
}

func TestCompleteThenCancel_RestocksOnce(t *testing.T) {
	integration.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000208")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusPending,
		time.Now().Add(5*time.Minute),
		initialStock, amount,
	)

	completeRes, err := repo.CompleteOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, completeRes.Updated)
	assert.Equal(t, initialStock, getStock(t, pool, productID))

	cancelRes, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, cancelRes.Updated)
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))

	cancelRes2, err := repo.CancelOrder(ctx, orderID)
	require.NoError(t, err)
	assert.False(t, cancelRes2.Updated)
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))
}
