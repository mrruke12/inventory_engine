package rabbit

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/mrruke12/inventory_engine/pkg/backoff"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer func(context.Context, *amqp.Channel) error

type ConsumerManager struct {
	connString             string
	consumers              []Consumer
	declareTopology        DeclareTopologyFunc
	connectionBackoff      *backoff.Backoff
	consumerBackoffFactory func() *backoff.Backoff
	shutdownTimeout        time.Duration
}

func NewConsumerManager(
	connString string,
	consumers []Consumer,
	declareTopology DeclareTopologyFunc,
	connectionBackoff *backoff.Backoff,
	consumerBackoffFactory func() *backoff.Backoff,
	shutdownTimeout time.Duration,
) *ConsumerManager {
	return &ConsumerManager{
		connString:             connString,
		consumers:              slices.Clone(consumers),
		connectionBackoff:      connectionBackoff,
		consumerBackoffFactory: consumerBackoffFactory,
		shutdownTimeout:        shutdownTimeout,
	}
}

// Blocking Consumers WatchDog.
//
// Runs consumers, handles lifecycle of connection, channels.
// Automatically reconnects and restarts consumers.
//
// Note: consumers MUST handle context cancellation.
func (m *ConsumerManager) WatchConnection(ctx context.Context) {
	for {
		if !m.connectionBackoff.Wait(ctx) {
			return
		}

		m.connectionBackoff.AutoDecreaseDelay() // Automatically determine if backoff delay should be decreased (e.g. if delay hasn't increased for 2x current delay time)

		m.runSession(ctx)
	}
}

// Establishes a connection, runs consumers, waits for connection
// to close or context to be canceled, returns when consumers
// are exited.
func (m *ConsumerManager) runSession(ctx context.Context) {
	sessionCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	conn, err := Connect(m.connString)

	if err != nil {
		slog.Warn("RabbitMQ connection is not available", "error", err)
		m.connectionBackoff.IncreaseDelay()
		return // RabbitMQ is not available, exit and try to reconnect
	}

	defer conn.Close()

	if err = m.declareTopology(conn); err != nil {
		slog.Warn("Could not declare topology for RabbitMQ connection", "error", err)
		m.connectionBackoff.IncreaseDelay()
		return // Exit and try to reconnect
	}

	notify := conn.NotifyClose(make(chan *amqp.Error, 1))
	var wg sync.WaitGroup

	for i := range m.consumers {
		wg.Add(1)
		m.watchChannel(sessionCtx, cancel, conn, &wg, m.consumers[i]) // Run consumers
	}

	select {
	case <-notify: // conn is closed
		slog.Warn("RabbitMQ connection is closed")
		cancel(context.Canceled) // signal to the consumers to stop
		m.connectionBackoff.IncreaseDelay()
	case <-ctx.Done():
		_ = conn.Close()
	case <-sessionCtx.Done(): // ctx or sessionCtx is canceled
		_ = conn.Close()
		if ctx.Err() == nil { // sessionCtx is canceled by a consumer, not by the parent context
			slog.Warn("Consumer channel faced an error", "error", context.Cause(sessionCtx))
			m.connectionBackoff.IncreaseDelay()
		}
	}

	m.waitConsumers(&wg) // wait for consumers to exit
}

// Creates a channel and restart consumer on errors.
// Stop watching on context cancellation, or conn error.
func (m *ConsumerManager) watchChannel(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	conn *amqp.Connection,
	wg *sync.WaitGroup,
	consumer Consumer,
) {
	go func() {
		defer wg.Done()

		consumerBackoff := m.consumerBackoffFactory()

		for { // Loop restarts consumer on error or panic if the connection is still alive
			if !consumerBackoff.Wait(ctx) { // Wait backoff delay or exit if context is canceled
				return
			}

			consumerBackoff.AutoDecreaseDelay()

			ch, err := conn.Channel()

			if err != nil {
				cancel(err) // Stop the consumer and other consumers in the session, since the conn must be updated
				return
			}

			func() { // Run consumer in anonymous function to handle panic
				defer ch.Close()

				defer func() {
					if r := recover(); r != nil {
						slog.Warn("Consumer paniced", "error", r)
						consumerBackoff.IncreaseDelay()
					}
				}()

				// In order to ensure the correct behaviour in a case
				// of the context being canceled, consumer MUST handle
				// its cancellation; otherwise consumer might end up
				// in a panic.
				err := consumer(ctx, ch)

				if err != nil {
					slog.Warn("Consumer exited with error", "error", err)
					consumerBackoff.IncreaseDelay()
				}
			}()
		}
	}()
}

// Wait for consumers to exit. Exit on timeout to prevent
// infinite waiting.
func (m *ConsumerManager) waitConsumers(wg *sync.WaitGroup) {
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(m.shutdownTimeout):
		slog.Warn("Consumers didn't exit on time", "timeout", m.shutdownTimeout)
	case <-done:
	}
}
