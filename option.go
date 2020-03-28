package timing

import (
	"sync/atomic"
	"time"
)

// EntryOption entry option
type EntryOption func(e *Entry)

// WithGoroutine the entry will use goroutine to do the job
// if not use goroutine which set it false,it will done on one goroutine
// default not use goroutine
func WithGoroutine() EntryOption {
	return func(e *Entry) {
		atomic.StoreUint32(&e.useGoroutine, 1)
	}
}

// Option user's option
type Option func(tim *Timing)

// WithLocation overrides the timezone of the instance.
func WithLocation(loc *time.Location) Option {
	return func(tim *Timing) {
		tim.location = loc
	}
}

// WithLimitSize overwrite job chan size,default value is DefaultLimitSize
func WithLimitSize(size int) Option {
	return func(tim *Timing) {
		tim.limitSize = size
	}
}

// WithTimeoutLimit overwrite timeout limit,default value is DefaultTimeoutLimit
func WithTimeoutLimit(tm time.Duration) Option {
	return func(tim *Timing) {
		if tm > 0 {
			tim.timeoutLimit = tm
		}
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
			tim.panicHandle = f
		}
	}
}
