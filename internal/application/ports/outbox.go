package ports

import (
	"context"
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID          string
	Payload     json.RawMessage
	RoutingKey  string
	OccurredAt  time.Time
	DeliveredAt *time.Time
}

type Outbox interface {
	Save(ctx context.Context, events ...OutboxEvent) error
	Pull(ctx context.Context, batchSize int) (events []OutboxEvent, err error)
	MarkDelivered(ctx context.Context, ids []string) (marked int, err error)
}
