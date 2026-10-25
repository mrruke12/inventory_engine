package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrruke12/inventory_engine/internal/application/ports"
)

type Outbox struct {
	pool Querier
}

func NewOutbox(pool Querier) ports.Outbox {
	return &Outbox{
		pool: pool,
	}
}

func (o *Outbox) Save(ctx context.Context, events ...ports.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	eventsLen := len(events)

	ids := make([]string, eventsLen)
	payloads := make([]json.RawMessage, eventsLen)
	routingKeys := make([]string, eventsLen)
	occurenceTimes := make([]time.Time, eventsLen)

	for i := range events {
		ids[i] = events[i].ID
		payloads[i] = events[i].Payload
		routingKeys[i] = events[i].RoutingKey
		occurenceTimes[i] = events[i].OccurredAt
	}

	sql := `
		insert into outbox_events (id, payload, routing_key, occurred_at)
		select v.id, v.payload, v.routing_key, v.occurred_at
		from unnest(
			$1::text[],
			$2::jsonb[],
			$3::text[],
			$4::timestamptz[]
		) as v(id, payload, routing_key, occurred_at)
	`

	_, err := o.pool.Exec(ctx, sql, ids, payloads, routingKeys, occurenceTimes)

	return err
}

func (o *Outbox) Pull(ctx context.Context, batchSize int) ([]ports.OutboxEvent, error) {
	sql := `
		select id, payload, routing_key, occurred_at from outbox_events
		where delivered_at is null
		limit $1 
	`

	rows, err := o.pool.Query(ctx, sql, batchSize)

	if err != nil {
		return nil, err
	}

	var events []ports.OutboxEvent

	for rows.Next() {
		e := ports.OutboxEvent{}

		if err = rows.Scan(&e.ID, &e.Payload, &e.RoutingKey, &e.OccurredAt); err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, nil
}

func (o *Outbox) MarkDelivered(ctx context.Context, ids []string) (int, error) {
	sql := `
		update outbox_events
		set delivered_at = now()
		where id = any($1::text[])
	`

	tag, err := o.pool.Exec(ctx, sql, ids)

	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}
