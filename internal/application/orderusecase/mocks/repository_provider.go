package mocks

import (
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type RepositoryProvider struct {
	Idemp     ports.Idempotency
	OrderRepo order.Repository
}

func (p *RepositoryProvider) Idempotency() ports.Idempotency {
	if p.Idemp == nil {
		panic("Idemp is not set")
	}
	return p.Idemp
}

func (p *RepositoryProvider) OrderRepository() order.Repository {
	if p.OrderRepo == nil {
		panic("OrderRepo is not set")
	}
	return p.OrderRepo
}
