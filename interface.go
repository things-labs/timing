package timing

// Job 定时任务接口
type Job interface {
	// 返回false,后续不在持行这个Job,
	// 返回true,后续继续执行这个Job
	Run() bool
}
