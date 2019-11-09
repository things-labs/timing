package timing

import (
	"time"
)

// Option 选项
type Option func(*Timing)

// WithTick override interval 时间粒子
func WithTick(tick time.Duration) Option {
	return func(timing *Timing) {
		timing.tick = tick
	}
}

// WithInterval override interval 默认条目时间间隔
func WithInterval(interval time.Duration) Option {
	return func(timing *Timing) {
		timing.interval = interval
	}
}

// WithGoroutine override useGoroutine 回调使用goroutine执行
func WithGoroutine() Option {
	return func(timing *Timing) {
		timing.useGoroutine = true
	}
}
