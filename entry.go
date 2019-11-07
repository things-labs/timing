package timing

import (
	"time"
)

// Entry 条目
type Entry struct {
	next time.Time // next 下一次运行时间  0: 表示未运行,或未启动

	count int32 // 任务执行的次数

	number int32 //任务需要执行的次数

	interval time.Duration

	job Job
}

type entriesByTime []*Entry

// Len implement sort.Interface
func (sf entriesByTime) Len() int { return len(sf) }

// Swap implement sort.Interface
func (sf entriesByTime) Swap(i, j int) { sf[i], sf[j] = sf[j], sf[i] }

// Less implement sort.Interface
func (sf entriesByTime) Less(i, j int) bool {
	if sf[i].next.IsZero() {
		return false
	}
	if sf[j].next.IsZero() {
		return true
	}
	return sf[i].next.Before(sf[j].next)
}

// Push implement heap.Interface
func (sf *entriesByTime) Push(x interface{}) {
	*sf = append(*sf, x.(*Entry))
}

// Pop implement heap.Interface
func (sf *entriesByTime) Pop() interface{} {
	old := *sf
	n := len(old)
	x := old[n-1]
	*sf = old[0 : n-1]
	return x
}

// 主要用于直接删除,未排序
func (sf *entriesByTime) remove(entry *Entry) {
	entries := []*Entry(*sf)
	for i, e := range entries {
		if e == entry {
			entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	*sf = entries
}
