package timing

import (
	"time"
)

type apply interface {
	setInterval(interval time.Duration)
	setGranularity(gra time.Duration)
	UseGoroutine(use bool)
}

// Option 选项
type Option func(apply)

// WithGranularity override interval 时间粒子
func WithGranularity(gra time.Duration) Option {
	return func(ap apply) {
		ap.setGranularity(gra)
	}
}

// WithInterval override interval 默认条目时间间隔
func WithInterval(interval time.Duration) Option {
	return func(ap apply) {
		ap.setInterval(interval)
	}
}

// WithGoroutine override hasGoroutine 回调使用goroutine执行
func WithGoroutine(use bool) Option {
	return func(ap apply) {
		ap.UseGoroutine(use)
	}
}
