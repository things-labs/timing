package timing

// job 定时任务接口
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (sf JobFunc) Run() {
	sf()
}
