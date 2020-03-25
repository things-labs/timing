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

// WithPanicHandler panic handler when it happen
func WithPanicHandler(f func(err interface{})) Option {
	return func(tim *Timing) {
		if f != nil {
			tim.pf = f
		}
	}
}
