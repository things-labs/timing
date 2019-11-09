package timing

import (
	"container/list"
	"time"
)

const (
	DefaultGranularity = time.Millisecond * 100
)

const (
	// 主级 + 4个层级共32位
	tvrBits = 8 // 主级占8位
	tvnBits = 6 // 4个层级各占6位
	tvnNum  = 4 // 级轮个数
	tvrSize = 1 << tvrBits
	tvnSize = 1 << tvnBits
	tvrMask = tvrSize - 1
	tvnMask = tvnSize - 1
)

type entry struct {
	next uint32
	// 任务已经执行的次数
	count uint32
	// 任务需要执行的次数
	number uint32
	// 超时时间
	interval time.Duration
	// 任务
	job Job
}

type Element *list.Element

type Wheel struct {
	curTick     uint32
	granularity time.Duration
	spokes      []*list.List
	doNow       *list.List
}

func NewWheel() {
	wl := &Wheel{
		spokes:      make([]*list.List, tvrSize+tvnSize*tvnNum),
		doNow:       list.New(),
		granularity: DefaultGranularity,
	}

	// init all spoke
	for i := range wl.spokes {
		wl.spokes[i] = list.New()
	}

	wl.curTick = uint32(time.Now().UnixNano() / int64(wl.granularity))
}

func (sf *Wheel) AddJob(job Job, num uint32, interval time.Duration) Element {
	et := &entry{
		next:     uint32((time.Duration(time.Now().UnixNano()) + interval + sf.granularity - 1) / sf.granularity),
		count:    0,
		number:   num,
		interval: interval,
		job:      job,
	}

	if et.next == sf.curTick {
		return sf.doNow.PushBack(et)
	}
	return sf.addTimer(et)
}

func (sf *Wheel) addTimer(et *entry) *list.Element {
	next := et.next
	idx := next - sf.curTick
	if idx < tvrSize {
		spokeIdx := next & tvrMask
		return sf.spokes[spokeIdx].PushBack(et)
	}

	// 计算在哪一个层级
	level := 0
	for idx >>= tvrBits; idx >= tvnSize && level < (tvnNum-1); level++ {
		idx >>= tvnBits
	}
	spokeIdx := tvrSize + tvnSize*level +
		((next >> (tvrBits + tvnBits*level)) & tvnMask)
	return sf.spokes[spokeIdx].PushBack(et)
}

func (sf *Wheel) cascade() {
	for level, index := 0, 0; index == 0 && level < tvnNum; level++ {
		index = (sf.curTick >> (tvrMask + level*tvnNum)) & tvnMask
		spokeIdx := tvrSize + tvnSize*level + index
		lists := sf.spokes[spokeIdx]
		for e, tmp := lists.Front(), lists.Front(); e != nil; e = tmp {
			tmp = e.Next()
			lists.Remove(e)
			sf.addTimer(e.Value.(*entry))
		}
	}
}

func (sf *Wheel) runWork() {
	var waitMs time.Duration
	tick := time.NewTimer(sf.granularity)

	for {
		select {
		case now := <-tick.C:
			nano := now.UnixNano()
			waitMs = time.Duration(nano) % sf.granularity
			past := uint32(nano/int64(sf.granularity)) - sf.curTick
			for ; past > 0; past-- {
				sf.curTick++
				index := sf.curTick & tvrMask
				if index == 0 {
					sf.cascade()
				}

				sf.doNow.PushBackList(sf.spokes[index])
				sf.spokes[index] = list.New()
			}
		}
		// TODO: do all read work

		tick.Reset(waitMs)
	}

}
