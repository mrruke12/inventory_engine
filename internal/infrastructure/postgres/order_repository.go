package postgres

import (
	"context"
	"uuid"

	"github.com/mrruke12/inventory_engine/internal/domains/order"
	"github.com/mrruke12/inventory_engine/internal/domains/product"
)

type OrderRepository struct {
	pool Querier
}

func NewOrderRepository(pool Querier) order.Repository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *order.Order) error {
	sql := `
		insert into orders (id, status)
		values ($1, $2)
	`

	_, err := r.pool.Exec(ctx, sql, order.ID, order.Status)

	return err
}

func (r *OrderRepository) CreateAndReserveOrderItems(ctx context.Context, products []uuid.UUID, deltas []int) error {
	sql := `
		update products p
		set quantity = p.quantity - s.delta
		from unnest(
			$1::uuid[],
			$2::int[]
		) as s(id, delta)
		where p.id = s.id
		and p.quantity >= s.delta
	`

	tag, err := r.pool.Exec(ctx, sql, products, deltas)

	if err != err {
		return err
	}

	if int(tag.RowsAffected()) != len(products) {
		return product.ErrorNotEnoughProduct
	}

	return nil
}

func (r *OrderRepository) BatchMarkExpiredOrders(ctx context.Context, batchSize int) (int, error) {
	sql := `
		update orders o
		set status = 'expired',
			updated_at = now()
		where ctid in (
			select ctid from orders
			where expires_at < now()
				and status = 'pending_payment'
			order by expires_at
			limit $1
			for update skip locked
		)
	`

	tag, err := r.pool.Exec(ctx, sql, batchSize)

	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}
