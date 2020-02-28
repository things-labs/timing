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
// if not use goroutine,set it false and the set goroutine pool submit interface
func WithGoroutine(use bool, submit ...Submit) Option {
	return func(tim *Timing) {
		tim.useGoroutine = use
		tim.sb = append(submit, NopSubmit{})[0]
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
