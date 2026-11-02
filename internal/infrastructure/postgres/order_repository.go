package postgres

import (
	"context"
	"errors"
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

func (r *OrderRepository) CreateOrder(ctx context.Context, o *order.Order) error {
	sql := `
		insert into orders (id, status)
		values ($1, $2)
	`

	_, err := r.pool.Exec(ctx, sql, o.ID, o.Status)

	classifiedErr := classifyError(err)

	if errors.Is(classifiedErr, ErrConflict) {
		return order.ErrOrderExists
	}

	return err
}

func (r *OrderRepository) CreateAndReserveOrderItems(ctx context.Context, orderID uuid.UUID, items []order.Item) error {
	itemsLen := len(items)

	ids := make([]uuid.UUID, itemsLen)
	products := make([]uuid.UUID, itemsLen)
	amounts := make([]int, itemsLen)

	for i := range items {
		ids[i] = items[i].ID
		products[i] = items[i].ProductID
		amounts[i] = items[i].Amount
	}

	sql := `
		with locked_order as (
			select * from orders
			where id = $4
			for update
		),
		inserted_items as (
			insert into order_items (id, order_id, product_id, amount)
			select 
				t.id, 
				(select id from locked_order) as order_id, 
				t.product_id, 
				t.amount
			from unnest(
				$1::uuid[],
				$2::uuid[],
				$3::int[]
			) as t(id, product_id, amount)
			returning product_id, amount
		),
		locked_products as (
			select i.product_id as id, i.amount from products p
			join inserted_items i on i.product_id = p.id
			for update of p
		),
		reserved_products as (
			update products p
			set stock = p.stock - lp.amount
			from locked_products lp
			where p.id = lp.id
			and p.stock >= lp.amount
			returning p.id
		),
		info as ( 
			select 
				(select id from locked_order) is not null as found_order,
				(select count(*) from inserted_items) as items,
				(select count(*) from reserved_products) as products
		)
		select 
			found_order,
			(items = products) as valid
			from info
	`

	var found, valid bool
	err := r.pool.QueryRow(ctx, sql, ids, products, amounts, orderID).Scan(&found, &valid)

	if err != nil {
		return err
	}

	if !found {
		return order.ErrOrderNotFound
	}

	if !valid {
		return product.ErrNotEnoughProduct
	}

	return nil
}

func (r *OrderRepository) BatchMarkExpiredOrders(ctx context.Context, batchSize int) (int, error) {
	sql := `
		with expired_orders as (
			select id from orders
			where expires_at < now()
				and status = 'pending'
			order by expires_at, id
			limit $1
			for update skip locked
		), 
		expired_items as (
			select oi.product_id, sum(oi.amount) as amount
			from order_items oi
			join expired_orders eo on eo.id = oi.order_id
			group by oi.product_id
		),
		locked_products as (
			select p.id, ei.amount
			from products p
			join expired_items ei on p.id = ei.product_id
			order by p.id
			for update of p
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
			set stock = p.stock + ei.amount
			from expired_items ei
			where p.id = ei.product_id
			returning p.id
		)
		select count(*) as expired_orders_count from expired_orders
	`

	row := r.pool.QueryRow(ctx, sql, batchSize)
	var updated int

	if err := row.Scan(&updated); err != nil {
		return updated, err
	}

	return updated, nil
}

func (r *OrderRepository) ReleaseOrder(ctx context.Context, orderID uuid.UUID) (order.PendingOrderUpdateResult, error) {
	sql := `
		with locked_order as (
			select id, status, expires_at from orders
			where id = $1
			for update
		),
		items as (
			select * from order_items i
			join locked_order o 
			on o.id = i.order_id
			and o.status = 'pending' -- do not select items if the order status is invalid, so we don't lock any products doing no-op
			and o.expires_at > now() -- do not select items if the order is expired
		),
		locked_products as (
			select p.id, i.amount from products p
			join items i on i.product_id = p.id
			order by p.id
			for update of p
		),
		updated_order as (
			update orders o
			set status = 'released'
			from locked_order l
			where o.id = l.id
				and o.status = 'pending'
				and o.expires_at > now()
			returning o.status
		),
		updated_products as (
			update products p
			set stock = p.stock + lp.amount
			from locked_products lp
			where p.id = lp.id
		)
		select
			(select status from locked_order) is not null as is_found,
			(
				select coalesce(
					(select expires_at from locked_order) <= now(),
					false
				)
			) as is_expired,
			(select status from locked_order) as initial_status,
			(select status from updated_order) is not null as is_updated
	`

	result := order.PendingOrderUpdateResult{}

	err := r.pool.QueryRow(ctx, sql, orderID).Scan(&result.Found, &result.Expired, &result.InitialStatus, &result.Updated)

	return result, err
}

func (r *OrderRepository) CompleteOrder(ctx context.Context, orderID uuid.UUID) (order.PendingOrderUpdateResult, error) {
	sql := `
		with locked_order as (
			select id, status, expires_at
			from orders where id = $1
		),
		updated_order as (
			update orders o
			set status = 'completed'
			from locked_order lo
			where o.id = lo.id
				and o.status = 'pending'
				and o.expires_at > now()
			returning o.status
		)
		select
			(select status from locked_order) is not null as is_found,
			(
				select coalesce(
					(select expires_at from locked_order) <= now(),
					false
				)
			) as is_expired,
			(select status from locked_order) as initial_status,
			(select status from updated_order) is not null as is_updated
	`

	result := order.PendingOrderUpdateResult{}

	err := r.pool.QueryRow(ctx, sql, orderID).Scan(&result.Found, &result.Expired, &result.InitialStatus, &result.Updated)

	return result, err
}

func (r *OrderRepository) CancelOrder(ctx context.Context, orderID uuid.UUID) (order.CancelOrderResult, error) {
	sql := `
		with locked_order as (
			select id, status from orders
			where id = $1
			for update
		),
		items as (
			select * from order_items i
			join locked_order o 
			on o.id = i.order_id
			and o.status = 'completed' -- do not select items if the order status is invalid, so we don't lock any products doing no-op
		),
		locked_products as (
			select p.id, i.amount from products p
			join items i on i.product_id = p.id
			order by p.id
			for update of p
		),
		updated_order as (
			update orders o
			set status = 'canceled'
			from locked_order l
			where o.id = l.id
				and o.status = 'completed'
			returning o.status
		),
		updated_products as (
			update products p
			set stock = p.stock + lp.amount
			from locked_products lp
			where p.id = lp.id
		)
		select
			(select status from locked_order) is not null as is_found,
			(select status from locked_order) as initial_status,
			(select status from updated_order) is not null as is_updated
	`

	result := order.CancelOrderResult{}

	err := r.pool.QueryRow(ctx, sql, orderID).Scan(&result.Found, &result.InitialStatus, &result.Updated)

	return result, err
}
