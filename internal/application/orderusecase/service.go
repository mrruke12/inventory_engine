package orderusecase

import (
	"github.com/mrruke12/inventory_engine/internal/application/ports"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

type Service struct {
	uow       ports.UnitOfWork
	orderRepo order.Repository
}

func NewService(uow ports.UnitOfWork, orderRepo order.Repository) *Service {
	return &Service{
		uow:       uow,
		orderRepo: orderRepo,
	}
}
