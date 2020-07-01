package timing

type heapData struct {
	queue []*Timer
	items map[*Timer]struct{}
}

func (h *heapData) Len() int { return len(h.queue) }
func (h heapData) Swap(i, j int) {
	h.queue[i].index, h.queue[j].index = h.queue[j].index, h.queue[i].index
	h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
}
func (h *heapData) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "smaller" than any other time.
	// (To sort it at the front of the list.)
	if h.queue[i].next.IsZero() {
		return true
	}
	if h.queue[j].next.IsZero() {
		return false
	}
	return h.queue[i].next.Before(h.queue[j].next)
}

func (h *heapData) Push(x interface{}) {
	n := len(h.queue)
	item := x.(*Timer)
	item.index = n
	h.items[item] = struct{}{}
	h.queue = append(h.queue, item)
}

func (h *heapData) Pop() interface{} {
	n := len(h.queue)
	item := h.queue[n-1]
	item.index = -1 // for safety
	h.queue = h.queue[:n-1]
	delete(h.items, item)
	return item
}
func (h *heapData) peek() *Timer {
	if len(h.queue) > 0 {
		return h.queue[0]
	}
	return nil
}

func (h *heapData) contains(tm *Timer) bool {
	_, ok := h.items[tm]
	return ok
}
