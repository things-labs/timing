package timing

import (
	"container/heap"
	"fmt"
	"time"
)

func Example_entryByTime() {
	entries := entryByTime([]*Entry{
		{count: 5, next: time.Date(2019, 1, 1, 1, 1, 5, 1, time.Local)},
		{count: 1, next: time.Date(2019, 1, 1, 1, 1, 1, 1, time.Local)},
		{count: 3, next: time.Date(2019, 1, 1, 1, 1, 3, 1, time.Local)},
		{count: 2, next: time.Date(2019, 1, 1, 1, 1, 2, 1, time.Local)},
		{count: 6, next: time.Time{}},
	})

	heap.Init(&entries)
	heap.Push(&entries, &Entry{count: 4, next: time.Date(2019, 1, 1, 1, 1, 4, 1, time.Local)})

	for entries.Len() > 0 {
		item := heap.Pop(&entries).(*Entry)
		fmt.Printf("%d ", item.count)
	}

	// Output:
	// 1 2 3 4 5 6
}
