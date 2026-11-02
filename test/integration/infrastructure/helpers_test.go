//go:build integration

package infrastructure

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/stretchr/testify/require"
)

func insertOrder(t *testing.T, pool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}, id uuid.UUID, status order.Status, expiresAt time.Time) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx,
		`insert into orders (id, status, expires_at) values ($1, $2, $3)`,
		id, string(status), expiresAt,
	)
	require.NoError(t, err)
}

// helper: создаёт продукт с заданным stock и возвращает его id.
func insertProduct(t *testing.T, pool *pgxpool.Pool, name string, stock int) uuid.UUID {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.New()
	_, err := pool.Exec(ctx,
		`insert into products (id, name, stock) values ($1, $2, $3)`,
		id, name, stock,
	)
	require.NoError(t, err)
	return id
}

func insertOrderItem(t *testing.T, pool *pgxpool.Pool, orderID, productID uuid.UUID, amount int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx,
		`insert into order_items (id, order_id, product_id, amount) values ($1, $2, $3, $4)`,
		uuid.New(), orderID, productID, amount,
	)
	require.NoError(t, err)
}

func getStock(t *testing.T, pool *pgxpool.Pool, productID uuid.UUID) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stock int
	err := pool.QueryRow(ctx, `select stock from products where id = $1`, productID).Scan(&stock)
	require.NoError(t, err)
	return stock
}

func getOrderStatus(t *testing.T, pool *pgxpool.Pool, orderID uuid.UUID) order.Status {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var status string
	err := pool.QueryRow(ctx, `select status from orders where id = $1`, orderID).Scan(&status)
	require.NoError(t, err)
	return order.Status(status)
}
