package backoff

import (
	"context"
	"log/slog"
	"time"
)

type Backoff struct {
	delay         time.Duration
	minDelay      time.Duration
	maxDelay      time.Duration
	delayStrategy DelayStrategy
}

func NewBackoff(minDelay, maxDelay int, delayStrategy DelayStrategy) *Backoff {
	return &Backoff{
		minDelay:      time.Duration(minDelay) * time.Millisecond,
		maxDelay:      time.Duration(maxDelay) * time.Millisecond,
		delay:         delayStrategy.Init(),
		delayStrategy: delayStrategy,
	}
}

// Return current delay.
func (m *Backoff) Delay() time.Duration {
	return m.delay
}

// Increases the delay accordingly with the delayStrategy.
func (m *Backoff) IncreaseDelay() {
	if m.delayStrategy == nil {
		slog.Warn("Strategy is not set for Backoff")
		return
	}

	m.setDelay(m.delayStrategy.Increase(m.delay))
}

// Decreases the delay accordingly with the delayStrategy.
func (m *Backoff) DecreaseDelay() {
	if m.delayStrategy == nil {
		slog.Warn("Strategy is not set for Backoff")
		return
	}

	m.setDelay(m.delayStrategy.Decrease(m.delay))
}

// Automatically decide whether decrease the delay
// or leave it unchanged, accordingly with the delayStrategy.
func (m *Backoff) AutoDecreaseDelay() {
	if m.delayStrategy == nil {
		slog.Warn("Strategy is not set for Backoff")
		return
	}

	m.setDelay(m.delayStrategy.AutoDecrease(m.delay))
}

// Resets the delay accrodingly with the delayStrategy.
func (m *Backoff) ResetDelay() {
	if m.delayStrategy == nil {
		slog.Warn("Strategy is not set for Backoff")
		return
	}

	m.setDelay(m.delayStrategy.Reset(m.delay))
}

// Wait for delay or context cancellation.
//
// Returns true if exits on delay, and false if the context is canceled.
func (m *Backoff) Wait(ctx context.Context) bool {
	select {
	case <-time.After(m.delay):
		return true
	case <-ctx.Done():
		return false
	}
}

func (m *Backoff) setDelay(delay time.Duration) {
	if delay < m.minDelay {
		m.delay = m.minDelay
	} else if delay > m.maxDelay {
		m.delay = m.maxDelay
	} else {
		m.delay = delay
	}
}
