//go:build integration

package infrastructure

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
	"github.com/mrruke12/inventory_engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrderWithItem(
	t *testing.T,
	pool *pgxpool.Pool,
	orderID uuid.UUID,
	status order.Status,
	expiresAt time.Time,
	initialStock, amount int,
) (productID uuid.UUID) {
	t.Helper()

	insertOrder(t, pool, orderID, status, expiresAt)
	productID = insertProduct(t, pool, "product-"+orderID.String(), initialStock)
	insertOrderItem(t, pool, orderID, productID, amount)
	return productID
}

func TestReleaseOrder_NotFound(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := repo.ReleaseOrder(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000101"))
	require.NoError(t, err)

	assert.False(t, res.Found)
	assert.False(t, res.Updated)
	assert.False(t, res.Expired)
	assert.Nil(t, res.InitialStatus)
}

func TestReleaseOrder_PendingNotExpired_UpdatesAndRestocks(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000102")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusPending,
		time.Now().Add(5*time.Minute),
		initialStock, amount,
	)

	res, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	assert.Equal(t, order.StatusReleased, getOrderStatus(t, pool, orderID))
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))
}

func TestReleaseOrder_MultipleItems_RestocksSum(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000103")
	insertOrder(t, pool, orderID, order.StatusPending, time.Now().Add(5*time.Minute))

	p1 := insertProduct(t, pool, "p1", 5)
	p2 := insertProduct(t, pool, "p2", 7)

	insertOrderItem(t, pool, orderID, p1, 2)
	insertOrderItem(t, pool, orderID, p2, 4)

	res, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	assert.Equal(t, order.StatusReleased, getOrderStatus(t, pool, orderID))
	assert.Equal(t, 5+2, getStock(t, pool, p1))
	assert.Equal(t, 7+4, getStock(t, pool, p2))
}

func TestReleaseOrder_PendingExpired_NoEffect(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000104")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusPending,
		time.Now().Add(-1*time.Minute),
		initialStock, amount,
	)

	res, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.True(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	assert.Equal(t, order.StatusPending, getOrderStatus(t, pool, orderID))
	assert.Equal(t, initialStock, getStock(t, pool, productID))
}

func TestReleaseOrder_NonPendingStatus_NoEffect(t *testing.T) {
	cases := []struct {
		name   string
		status order.Status
	}{
		{"completed", order.StatusCompleted},
		{"released", order.StatusReleased},
		{"canceled", order.StatusCanceled},
		{"expired", order.StatusExpired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.CleanupPostgres(t, pool)
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

			res, err := repo.ReleaseOrder(ctx, orderID)
			require.NoError(t, err)

			assert.True(t, res.Found)
			assert.False(t, res.Updated)
			assert.False(t, res.Expired)
			require.NotNil(t, res.InitialStatus)
			assert.Equal(t, tc.status, *res.InitialStatus)

			assert.Equal(t, tc.status, getOrderStatus(t, pool, orderID))
			assert.Equal(t, initialStock, getStock(t, pool, productID))
		})
	}
}

func TestReleaseOrder_NonPendingExpired_ExpiredTrue(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000105")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusCompleted,
		time.Now().Add(-1*time.Minute),
		initialStock, amount,
	)

	res, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.True(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res.InitialStatus)

	assert.Equal(t, order.StatusCompleted, getOrderStatus(t, pool, orderID))
	assert.Equal(t, initialStock, getStock(t, pool, productID))
}

func TestReleaseOrder_Idempotent(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000106")
	const (
		initialStock = 10
		amount       = 3
	)

	productID := setupOrderWithItem(
		t, pool, orderID, order.StatusPending,
		time.Now().Add(5*time.Minute),
		initialStock, amount,
	)

	res1, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, res1.Found)
	assert.True(t, res1.Updated)
	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))

	res2, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)
	assert.True(t, res2.Found)
	assert.False(t, res2.Updated)
	require.NotNil(t, res2.InitialStatus)
	assert.Equal(t, order.StatusReleased, *res2.InitialStatus)

	assert.Equal(t, initialStock+amount, getStock(t, pool, productID))
}

func TestReleaseOrder_NoItems_ReleasesWithoutRestock(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000107")
	insertOrder(t, pool, orderID, order.StatusPending, time.Now().Add(5*time.Minute))

	res, err := repo.ReleaseOrder(ctx, orderID)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	assert.Equal(t, order.StatusReleased, getOrderStatus(t, pool, orderID))
}
