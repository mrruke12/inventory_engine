package orderusecase

import (
	"log/slog"
	"runtime/debug"

	"github.com/mrruke12/inventory_engine/internal/application"
	"github.com/mrruke12/inventory_engine/internal/domains/order"
)

func requireInitialStatus(initialStatus *order.Status) error {
	if initialStatus == nil {
		slog.Error(
			"Order found but initial status is nil",
			"stack", debug.Stack,
		)

		return application.ErrImpossibleCondition
	}

	return nil
}
