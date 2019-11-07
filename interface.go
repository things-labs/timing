package timing

// Job 定时任务接口
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement Job interface
func (sf JobFunc) Run() {
	sf()
}
