package timing

import (
	"time"
)

// Option user's option
type Option func(tim *Timing)

// WithLocation overrides the timezone of the instance.
func WithLocation(loc *time.Location) Option {
	return func(tim *Timing) {
		tim.location = loc
	}
}

// WithGoroutine override useGoroutine or goroutine pool
func WithGoroutine(use bool) Option {
	return func(tim *Timing) {
		tim.UseGoroutine(use)
	}
}

// WithLoggerProvider override default logger provider
func WithLoggerProvider(p LogProvider) Option {
	return func(tim *Timing) {
		tim.setLogProvider(p)
	}
}

// WithLogger enable logger
func WithLogger() Option {
	return func(tim *Timing) {
		tim.LogMode(true)
	}
}

// WithGoroutinePoolCapacity overwrite goroutine pool capacity
func WithGoroutinePoolCapacity(cap int32) Option {
	return func(tim *Timing) {
		tim.capacity = cap
	}
}

// WithGoroutinePoolSurvivalTime overwrite goroutine pool survival time
func WithGoroutinePoolSurvivalTime(t time.Duration) Option {
	return func(tim *Timing) {
		tim.survivalTime = t
	}
}

// WithGoroutinePoolCleanupTime overwrite goroutine pool cleanup time
func WithGoroutinePoolCleanupTime(t time.Duration) Option {
	return func(tim *Timing) {
		tim.miniCleanupTime = t
	}
}
