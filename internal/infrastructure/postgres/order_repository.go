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

func (r *OrderRepository) CreateAndReserveOrderItems(ctx context.Context, items []order.Item) error {
	itemsLen := len(items)

	orderIDs := make([]uuid.UUID, itemsLen)
	ids := make([]uuid.UUID, itemsLen)
	products := make([]uuid.UUID, itemsLen)
	deltas := make([]int, itemsLen)

	for i := range items {
		orderIDs[i] = items[i].OrderID
		ids[i] = items[i].ID
		products[i] = items[i].ProductID
		deltas[i] = items[i].Quantity
	}

	sql := `
		with input(id, order_id, product_id, delta) as (
			select * from unnest(
				$1::uuid[],
				$2::uuid[],
				$3::uuid[],
				$4::int[]
			)
		),
		reserved as (
			update products p
			set quantity = p.quantity - i.delta
			from input i
			where p.id = i.product_id
			and p.quantity >= i.delta
			returning i.id, i.order_id, i.product_id, i.delta
		),
		inserted as (
			insert into order_items (id, order_id, product_id, quantity)
			select id, order_id, product_id, delta from reserved
			returning id
		)
		select count(*) from inserted
	`

	var reserved int
	err := r.pool.QueryRow(ctx, sql, ids, orderIDs, products, deltas).Scan(&reserved)

	if err != nil {
		return err
	}

	if reserved != len(products) {
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
