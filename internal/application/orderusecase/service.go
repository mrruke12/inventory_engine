package orderusecase

import (
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type Service struct {
	uow       ports.UnitOfWork
	idemp     ports.Idempotency
	orderRepo order.Repository
}

func NewService(uow ports.UnitOfWork, idemp ports.Idempotency, orderRepo order.Repository) *Service {
	return &Service{
		idemp:     idemp,
		uow:       uow,
		orderRepo: orderRepo,
	}
}
