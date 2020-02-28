package timing

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (sf JobFunc) Run() {
	sf()
}

// Submit goroutine pool interface
type Submit interface {
	Submit(job Job)
}

// NopSubmit empty struct implement Submit interface
type NopSubmit struct{}

// Submit implement Submit interface
func (NopSubmit) Submit(Job) {}
