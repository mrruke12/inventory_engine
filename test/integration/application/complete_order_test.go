//go:build integration

package application

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests CompleteOrder response for different order statuses
func TestCompleteOrder_Existing(t *testing.T) {
	tests := []struct {
		name            string
		status          order.Status
		expectedUpdated bool
		expectedError   error
	}{
		{"release pending", order.StatusPending, true, nil},
		{"release expired", order.StatusExpired, false, order.ErrOrderExpired},
		{"release released", order.StatusReleased, false, order.ErrOrderReleased},
		{"release canceled", order.StatusCanceled, false, order.ErrOrderCanceled},
		{"release completed", order.StatusCompleted, false, nil},
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

			updated, err := orderService.CompleteOrder(ctx, orderID)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedUpdated, updated)
		})
	}
}

func TestCompleteOrder_ExpiredButNotMarked(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	productID := requireTestProduct(ctx, t, 1)
	orderID, err := createOrder(ctx, productID, uuid.New().String(), 1)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `update orders set expires_at = now() - interval '1 hours' where id = $1`, orderID)
	require.NoError(t, err)

	updated, err := orderService.CompleteOrder(ctx, orderID)
	assert.ErrorIs(t, err, order.ErrOrderExpired)
	assert.False(t, updated)
}

// Tests that CompleteOrder returns order.ErrOrderNotFound
// for not exitsting order
func TestCompleteOrder_NotExisting(t *testing.T) {
	integration.CleanupPostgres(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	updated, err := orderService.CompleteOrder(ctx, uuid.NewV7())

	assert.ErrorIs(t, err, order.ErrOrderNotFound)
	assert.False(t, updated)
}
