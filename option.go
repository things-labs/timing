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

// WithGoroutine overwrite useGoroutine or goroutine pool
// if not use goroutine,set it false and the set goroutine pool submit interface
func WithGoroutine(use bool) Option {
	return func(tim *Timing) {
		tim.useGoroutine = use
	}
}

// WithJobChanSize overwrite job chan size,default value is DefaultJobChanSize
func WithJobChanSize(size int) Option {
	return func(tim *Timing) {
		tim.jobsChanSize = size
	}
}

// WithLoggerProvider override default logger provider
func WithLoggerProvider(p LogProvider) Option {
	return func(tim *Timing) {
		tim.setLogProvider(p)
	}
}

// WithEnableLogger enable logger
func WithEnableLogger() Option {
	return func(tim *Timing) {
		tim.LogMode(true)
	}
}

func WithPanicHandler(f func(err interface{})) Option {
	return func(tim *Timing) {
		if f != nil {
			tim.pf = f
		}

	}
}
