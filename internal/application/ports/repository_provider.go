package ports

import "github.com/mrruke12/inventory_engine/internal/domains/order"

type RepositoryProvider interface {
	Idempotency() Idempotency
	OrderRepository() order.Repository
}
