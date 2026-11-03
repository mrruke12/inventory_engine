package order

import (
	"log/slog"
	"runtime/debug"

	"github.com/mrruke12/inventory_engine/internal/application"
)

type PendingOrderUpdateResult struct {
	InitialStatus *Status
	Expired       bool
	Updated       bool
	Found         bool
}

func (r *PendingOrderUpdateResult) Validate() error {
	return validate(r.InitialStatus, r.Updated, r.Found)
}

type CancelOrderResult struct {
	InitialStatus *Status
	Updated       bool
	Found         bool
}

func (r *CancelOrderResult) Validate() error {
	return validate(r.InitialStatus, r.Updated, r.Found)
}

func validate(initialStatus *Status, updated, found bool) error {
	if found && initialStatus == nil ||
		!found && updated {
		slog.Warn("Impossible condition",
			"initial status", initialStatus,
			"updated", updated,
			"found", found,
			"stack", debug.Stack,
		)
		return application.ErrImpossibleCondition
	}

	return nil
}
