package timing

import "time"

// EntryID 条日ID
type EntryID int

// Entry 条目
type Entry struct {
	id EntryID // id 用于标识这个条目

	next time.Time // next 下一次运行时间  0: 表示未运行,或未启动

	count int32 // 任务执行的次数

	job CronJob
}

type entryByTime []*Entry

// Len implement sort.Interface
func (sf entryByTime) Len() int { return len(sf) }

// Swap implement sort.Interface
func (sf entryByTime) Swap(i, j int) { sf[i], sf[j] = sf[j], sf[i] }

// Less implement sort.Interface
func (sf entryByTime) Less(i, j int) bool {
	if sf[i].next.IsZero() {
		return false
	}
	if sf[j].next.IsZero() {
		return true
	}
	return sf[i].next.Before(sf[j].next)
}
