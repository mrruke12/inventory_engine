package backoff

import (
	"time"
)

type DelayStrategy interface {
	Init() time.Duration
	Increase(time.Duration) time.Duration
	Decrease(time.Duration) time.Duration
	AutoDecrease(time.Duration) time.Duration
	Reset(time.Duration) time.Duration
}

type ExponentialDelayStrategy struct {
	base       time.Duration
	maxExp     int
	exp        int
	lastUpdate time.Time
}

func NewExponentialDelayStrategy(base, maxExp int) DelayStrategy {
	return &ExponentialDelayStrategy{
		base:       time.Duration(base) * time.Millisecond,
		maxExp:     maxExp,
		exp:        1,
		lastUpdate: time.Now(),
	}
}

func (s *ExponentialDelayStrategy) Init() time.Duration {
	return s.base
}

func (s *ExponentialDelayStrategy) Increase(delay time.Duration) time.Duration {
	if s.exp < s.maxExp {
		s.exp++
		return delay * time.Duration(s.base)
	}

	return delay
}

func (s *ExponentialDelayStrategy) Decrease(delay time.Duration) time.Duration {
	if s.exp > 0 {
		s.exp--
		return delay / time.Duration(s.base)
	}

	return delay
}

func (s *ExponentialDelayStrategy) AutoDecrease(delay time.Duration) time.Duration {
	if time.Since(s.lastUpdate) > delay*2 {
		return s.Decrease(delay)
	}

	return delay
}

func (s *ExponentialDelayStrategy) Reset(_ time.Duration) time.Duration {
	return 0
}
