package timing

import (
	"sync"
)

type pool struct {
	*sync.Pool
}

func newPool() pool {
	return pool{
		&sync.Pool{
			New: func() interface{} { return &Entry{} },
		},
	}
}

func (sf pool) get() *Entry {
	e := sf.Get().(*Entry)
	e.useGoroutine = 0
	return e
}

func (sf pool) put(e *Entry) {
	sf.Put(e)
}
