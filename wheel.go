package timing

import (
	"sync"
	"time"

	"github.com/thinkgos/list"
)

const (
	DefaultGranularity = time.Millisecond * 100
)

const (
	// 主级 + 4个层级共5级 占32位
	tvrBits = 8 // 主级占8位
	tvnBits = 6 // 4个层级各占6位
	tvnNum  = 4 // 层级个数
	tvrSize = 1 << tvrBits
	tvnSize = 1 << tvnBits
	tvrMask = tvrSize - 1
	tvnMask = tvnSize - 1
)

type entry struct {
	// 下一个超时
	next uint32
	// 超时时间
	interval time.Duration
	// 任务
	job Job
}

type Element list.Element

func (sf *Element) entry() *entry {
	return sf.Value.(*entry)
}

type Wheel struct {
	curTick     uint32
	granularity time.Duration
	spokes      []list.List
	doNow       list.List
	rw          sync.RWMutex
}

func NewWheel() *Wheel {
	wl := &Wheel{
		spokes:      make([]list.List, tvrSize+tvnSize*tvnNum),
		doNow:       list.List{},
		granularity: DefaultGranularity,
	}

	wl.curTick = uint32(time.Now().UnixNano() / int64(wl.granularity))
	return wl
}

func (sf *Wheel) Run() *Wheel {
	go sf.runWork()
	return sf
}

func (sf *Wheel) AddJob(job Job, timeout time.Duration) *Element {
	e := &Element{
		Value: &entry{
			next:     uint32((time.Duration(time.Now().UnixNano()) + timeout + sf.granularity - 1) / sf.granularity),
			interval: timeout,
			job:      job,
		},
	}

	sf.rw.Lock()
	defer sf.rw.Unlock()
	if e.entry().next == sf.curTick {
		return (*Element)(sf.doNow.PushElementBack((*list.Element)(e)))
	}
	return sf.addTimer(e)
}

func (sf *Wheel) addTimer(e *Element) *Element {
	var spokeIdx int

	next := e.entry().next
	if idx := next - sf.curTick; idx < tvrSize {
		spokeIdx = int(next & tvrMask)
	} else {
		// 计算在哪一个层级
		level := 0
		for idx >>= tvrBits; idx >= tvnSize && level < (tvnNum-1); level++ {
			idx >>= tvnBits
		}
		spokeIdx = tvrSize + tvnSize*level + int((next>>(tvrBits+tvnBits*level))&tvnMask)
	}
	return (*Element)(sf.spokes[spokeIdx].PushElementBack((*list.Element)(e)))
}

func (sf *Wheel) cascade() {
	for level, index := 0, 0; index == 0 && level < tvnNum; level++ {
		index = int((sf.curTick >> (tvrSize + level*tvnNum)) & tvnMask)
		spoke := sf.spokes[tvrSize+tvnSize*level+index]
		for spoke.Len() > 0 {
			sf.addTimer((*Element)(spoke.PopFront()))
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
			sf.rw.Lock()
			for ; past > 0; past-- {
				sf.curTick++
				index := sf.curTick & tvrMask
				if index == 0 {
					sf.cascade()
				}

				sf.doNow.SpliceBackList(&sf.spokes[index])
			}
			sf.rw.Unlock()
		}
		sf.rw.Lock()
		for sf.doNow.Len() > 0 {
			e := sf.doNow.PopFront()
			sf.rw.Unlock()
			(*Element)(e).entry().job.Run()
			sf.rw.Lock()
		}
		sf.rw.Unlock()
		tick.Reset(waitMs)
	}
}
