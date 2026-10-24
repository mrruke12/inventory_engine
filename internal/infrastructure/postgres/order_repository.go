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

func (r *OrderRepository) BatchCancelExpiredOrders(ctx context.Context, batchSize int) ([]int, error) {
	sql := `
		with expired_orders as (
			select id from orders
			where expires_at < now()
				and status = 'pending_payment'
			order by expires_at
			limit $1
			for update skip locked
		), 
		expired_items as (
			select product_id, quantity 
			from order_items oi
			join expired_orders eo on eo.id = oi.order_id
		),
		updated_orders as (
			update orders o
			set status = 'expired',
				updated_at = now()
			from expired_orders eo
			where o.id = eo.id
			returning o.id
		),
		updated_products as (
			update products p
			set quantity = p.quantity + ei.quantity
			from expired_items ei
			where p.id = ei.product_id
			returning p.id
		)
		select 
			(select count(*) from expired_orders) as expired_orders_count,
			(select count(*) from expired_items) as expired_items_count,
			(select count(*) from updated_orders) as updated_orders_count,
			(select count(*) from updated_products) as updated_products_count;
	`

	row := r.pool.QueryRow(ctx, sql, batchSize)
	data := make([]int, 4)

	if err := row.Scan(&data[0], &data[1], &data[2], &data[3]); err != nil {
		return nil, err
	}

	return data, nil
}

/*
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
*/
