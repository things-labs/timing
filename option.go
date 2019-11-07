package timing

import (
	"time"
)

// Option 选项
type Option func(*Timing)

// WithTick override interval
func WithTick(tick time.Duration) Option {
	return func(timing *Timing) {
		timing.tick = tick
	}
}

// WithInterval override interval
func WithInterval(interval time.Duration) Option {
	return func(timing *Timing) {
		timing.interval = interval
	}
}
