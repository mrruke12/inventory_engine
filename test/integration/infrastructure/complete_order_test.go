//go:build integration

package infrastructure

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
	"github.com/mrruke12/inventory_engine/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteOrder_NotFound(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := repo.CompleteOrder(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	require.NoError(t, err)

	assert.False(t, res.Found)
	assert.False(t, res.Updated)
	assert.False(t, res.Expired)
	assert.Nil(t, res.InitialStatus)
}

func TestCompleteOrder_PendingNotExpired_Updates(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	insertOrder(t, pool, id, order.StatusPending, time.Now().Add(5*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.True(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	var status string
	err = pool.QueryRow(ctx, `select status from orders where id = $1`, id).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusCompleted), status)
}

func TestCompleteOrder_PendingExpired_DoesNotUpdate(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	insertOrder(t, pool, id, order.StatusPending, time.Now().Add(-1*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.True(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusPending, *res.InitialStatus)

	// Статус должен остаться pending.
	var status string
	err = pool.QueryRow(ctx, `select status from orders where id = $1`, id).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, string(order.StatusPending), status)
}

func TestCompleteOrder_AlreadyCompleted_DoesNotUpdate(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	insertOrder(t, pool, id, order.StatusCompleted, time.Now().Add(5*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res.InitialStatus)
}

func TestCompleteOrder_ReleasedStatus_DoesNotUpdate(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	insertOrder(t, pool, id, order.StatusReleased, time.Now().Add(5*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusReleased, *res.InitialStatus)
}

func TestCompleteOrder_ExpiredStatus_DoesNotUpdate(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000006")
	insertOrder(t, pool, id, order.StatusExpired, time.Now().Add(-1*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.True(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusExpired, *res.InitialStatus)
}

func TestCompleteOrder_CanceledStatus_DoesNotUpdate(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000007")
	insertOrder(t, pool, id, order.StatusCanceled, time.Now().Add(5*time.Minute))

	res, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)

	assert.True(t, res.Found)
	assert.False(t, res.Updated)
	assert.False(t, res.Expired)
	require.NotNil(t, res.InitialStatus)
	assert.Equal(t, order.StatusCanceled, *res.InitialStatus)
}

// Проверяем, что повторный вызов CompleteOrder идемпотентен:
// первый — обновляет, второй — нет (статус уже completed).
func TestCompleteOrder_Idempotent(t *testing.T) {
	test.CleanupPostgres(t, pool)
	repo := postgres.NewOrderRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.MustParse("00000000-0000-0000-0000-000000000008")
	insertOrder(t, pool, id, order.StatusPending, time.Now().Add(5*time.Minute))

	res1, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)
	assert.True(t, res1.Found)
	assert.True(t, res1.Updated)
	assert.False(t, res1.Expired)

	res2, err := repo.CompleteOrder(ctx, id)
	require.NoError(t, err)
	assert.True(t, res2.Found)
	assert.False(t, res2.Updated)
	assert.False(t, res2.Expired)
	require.NotNil(t, res2.InitialStatus)
	assert.Equal(t, order.StatusCompleted, *res2.InitialStatus)
}
