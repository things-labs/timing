package timing

// Job 定时任务接口
type Job interface {
	// Deploy 得到周期时间和执行次数
	// 执行次数
	// < 0: 表示停止
	// 0: 重复执行
	// > 0: 执行次数
	// 动态修改只在下一次任务生效
	Deploy() (*Duration, *Int32)
	// 返回false,后续不在持行这个Job,
	// 返回true,后续继续执行这个Job
	Run() bool
}
