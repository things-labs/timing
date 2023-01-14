# timing  
[![GoDoc](https://godoc.org/github.com/things-labs/timing?status.svg)](https://godoc.org/github.com/things-labs/timing)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/things-labs/timing?tab=doc)
[![Build Status](https://travis-ci.org/things-labs/timing.svg)](https://travis-ci.org/things-labs/timing)
[![codecov](https://codecov.io/gh/things-labs/timing/branch/master/graph/badge.svg)](https://codecov.io/gh/things-labs/timing)
![Action Status](https://github.com/things-labs/timing/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/things-labs/timing)](https://goreportcard.com/report/github.com/things-labs/timing)
[![Licence](https://img.shields.io/github/license/things-labs/timing)](https://raw.githubusercontent.com/things-labs/timing/master/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/things-labs/timing)](https://github.com/things-labs/timing/tags)

### Feature
 - job default process in a goroutine,so the job do not take too long. if you have long time job,please use `WithGoroutine`
 - you can define every timer use goroutine.
 - scan timeout's timer time complexity O(1)
 - not limited max timeout
 
### Installation

Use go get.
```bash
    go get github.com/things-labs/timing
```

Then import the timing package into your own code.
```bash
    import "github.com/things-labs/timing"
```

### Example

---

```go
import (
	"log"
	"time"

	"github.com/things-labs/timing"
)

func main() {
	base := timing.New().Run()

	tm := timing.NewTimer()
	tm.WithJobFunc(func() {
		log.Println("hello 1")
		base.Add(tm, time.Second)
	})

	tm1 := timing.NewTimer()
	tm1.WithJobFunc(func() {
		log.Println("hello 2")
		base.Add(tm1, time.Second * 2)
	})
	base.Add(tm, time.Second * 1)
	base.Add(tm1, time.Second * 2)
	time.Sleep(time.Second * 60)
}

```

    
 
