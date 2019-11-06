package timing

import "time"

type EntryID int

type Entry struct {
	ID EntryID // ID 用于标识这个条目

	Next time.Time // Next 下一次运行时间  0: 表示未运行,或未启动

	count uint

	job CronJob
}

type entryByTime []*Entry

// Len implement sort.Interface
func (sf entryByTime) Len() int { return len(sf) }

// Swap implement sort.Interface
func (sf entryByTime) Swap(i, j int) { sf[i], sf[j] = sf[j], sf[i] }

// Less implement sort.Interface
func (sf entryByTime) Less(i, j int) bool {
	if sf[i].Next.IsZero() {
		return false
	}
	if sf[j].Next.IsZero() {
		return true
	}
	return sf[i].Next.Before(sf[j].Next)
}

// Peak 查看最近要到达超时的元素,如果没有,返回nil
func (sf entryByTime) peak() *Entry {
	if len(sf) == 0 || sf[0].Next.IsZero() {
		return nil
	}
	return sf[0]
}
