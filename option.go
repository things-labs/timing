package timing

import (
	"time"
)

// Option 选项
type Option func(tim *Timing)

// WithLocation overrides the timezone of the instance.
func WithLocation(loc *time.Location) Option {
	return func(tim *Timing) {
		tim.location = loc
	}
}

// WithGoroutine override useGoroutine 回调使用goroutine执行
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

func WithInterval(interval time.Duration) Option {
	return func(tim *Timing) {
		tim.interval = interval
	}
}
