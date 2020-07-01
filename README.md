# timing  
[![GoDoc](https://godoc.org/github.com/thinkgos/timing?status.svg)](https://godoc.org/github.com/thinkgos/timing)
[![Build Status](https://travis-ci.org/thinkgos/timing.svg?branch=master)](https://travis-ci.org/thinkgos/timing)
[![codecov](https://codecov.io/gh/thinkgos/timing/branch/master/graph/badge.svg)](https://codecov.io/gh/thinkgos/timing)
![Action Status](https://github.com/thinkgos/timing/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/timing)](https://goreportcard.com/report/github.com/thinkgos/timing)
[![Licence](https://img.shields.io/github/license/thinkgos/timing)](https://raw.githubusercontent.com/thinkgos/timing/master/LICENSE)  

### Feature
 - job default process in a goroutine,so the job do not take too long. if you have long time job,please use `WithGoroutine`
 - you can define every timer use goroutine.
 - scan timeout's timer time complexity O(1)
 - not limited max timeout
 
### Installation

Use go get.
```bash
    go get github.com/thinkgos/timing/v4
```

Then import the modbus package into your own code.
```bash
    import modbus "github.com/thinkgos/timing/v4"
```

### Example

---

```go
import (
	"log"
	"time"

	"github.com/thinkgos/timing/v4"
)

func main() {
	base := timing.New().Run()

	tm := timing.NewTimer(time.Second)
	tm.WithJobFunc(func() {
		log.Println("hello 1")
		base.Add(tm)
	})

	tm1 := timing.NewTimer(time.Second * 2)
	tm1.WithJobFunc(func() {
		log.Println("hello 2")
		base.Add(tm1)
	})
	base.Add(tm)
	base.Add(tm1)
	time.Sleep(time.Second * 60)
}

```

    
 