package orderusecase

import "github.com/mrruke12/inventory_engine/internal/application/ports"

type Service struct {
	uow ports.UnitOfWork
}
