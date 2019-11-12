package timing

import (
	"sync"
	"time"
)

var defaultTiming Base
var once sync.Once

func lazyInit() {
	once.Do(func() {
		defaultTiming = NewWheel().Run()
	})
}

// HasRunning 运行状态
func HasRunning() bool {
	lazyInit()
	return defaultTiming.HasRunning()
}

// UseGoroutine use goroutine or callback
func UseGoroutine(use bool) {
	lazyInit()
	defaultTiming.UseGoroutine(use)
}

// Len 条目个数
func Len() int {
	lazyInit()
	return defaultTiming.Len()
}

// NewJob 新建一个条目,条目未启动定时
func NewJob(job Job, num uint32, interval ...time.Duration) Timer {
	lazyInit()
	return defaultTiming.NewJob(job, num, interval...)
}

// NewJobFunc 新建一个条目,条目未启动定时
func NewJobFunc(f JobFunc, num uint32, interval ...time.Duration) Timer {
	return NewJob(f, num, interval...)
}

// AddJob 添加任务
func AddJob(job Job, num uint32, interval ...time.Duration) Timer {
	lazyInit()
	return defaultTiming.AddJob(job, num, interval...)
}

// AddOneShotJob 添加一次性任务
func AddOneShotJob(job Job, interval ...time.Duration) Timer {
	return AddJob(job, OneShot, interval...)
}

// AddPersistJob 添加周期性任务
func AddPersistJob(job Job, interval ...time.Duration) Timer {
	return AddJob(job, Persist, interval...)
}

// AddJobFunc 添加任务函数
func AddJobFunc(f JobFunc, num uint32, interval ...time.Duration) Timer {
	return AddJob(f, num, interval...)
}

// AddOneShotJobFunc 添加一次性任务函数
func AddOneShotJobFunc(f JobFunc, interval ...time.Duration) Timer {
	return AddJob(f, OneShot, interval...)
}

// AddPersistJobFunc 添加周期性函数
func AddPersistJobFunc(f JobFunc, interval ...time.Duration) Timer {
	return AddJob(f, Persist, interval...)
}

// Start 启动或重始启动e的计时
func Start(e Timer, newTimeout ...time.Duration) Base {
	lazyInit()
	return defaultTiming.Start(e, newTimeout...)
}

// Delete 删除条目
func Delete(e Timer) Base {
	lazyInit()
	return defaultTiming.Delete(e)
}

// Modify 修改条目的周期时间,重置计数且重新启动定时器
func Modify(e Timer, interval time.Duration) Base {
	lazyInit()
	return defaultTiming.Modify(e, interval)
}
