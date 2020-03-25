# timing  
[![GoDoc](https://godoc.org/github.com/thinkgos/timing?status.svg)](https://godoc.org/github.com/thinkgos/timing)
[![Build Status](https://travis-ci.org/thinkgos/timing.svg?branch=master)](https://travis-ci.org/thinkgos/timing)
[![codecov](https://codecov.io/gh/thinkgos/timing/branch/master/graph/badge.svg)](https://codecov.io/gh/thinkgos/timing)
![Action Status](https://github.com/thinkgos/timing/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/timing)](https://goreportcard.com/report/github.com/thinkgos/timing)
[![Licence](https://img.shields.io/github/license/thinkgos/timing)](https://raw.githubusercontent.com/thinkgos/timing/master/LICENSE)  
 - 时间定时器,采用优先级队列
 - 时间任务调度,任务处理
 - 任务默认在一个协程池中处理,任务频繁不耗时可以使用
 - 每一个条目可以配置是否使用goroutine处理
 - 扫描超时条目时间复杂度o(1).
 - 不限最大时间

### Installation

Use go get.
```bash
    go get github.com/thinkgos/timing/v2
```

Then import the modbus package into your own code.
```bash
    import modbus "github.com/thinkgos/timing/v2"
```

### Example

---

```go
import (
	"log"
	"time"

	"github.com/thinkgos/timing/v2"
)

func main() {
	var tm *timing.Entry

	t := timing.New(timing.WithEnableLogger()).Run()

	e1 := t.AddPersistJob(timing.JobFunc(func() {
		panic("haha")
	}), time.Second*1)
	// 此条目使用goroutine
	tm = timing.NewEntry(timing.JobFunc(func() {
		log.Println("in goroutine")
		t.Start(tm)
	}), timing.OneShot, time.Second*2).WithGoroutine(true)

	t.Start(tm)

	time.Sleep(time.Second * 30)
	// 如果相应条目不使用,需要使用remove删除
	t.Remove(e1)
	t.Close()
	time.Sleep(time.Second * 5)
}
```

**Note:** 
    默认情况下在job函数处理里,任务应尽快处理,不宣有阻塞的任务,如果使用的为耗时任务,请使用goroutine
    
 