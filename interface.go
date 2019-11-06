package timing

// CronJob 定时任务接口
type CronJob interface {
	// Deploy 得到周期时间和执行次数
	// 执行次数
	// < 0: 表示停止
	// 0: 重复执行
	// > 0: 执行次数
	// 动态修改只在下一次任务生效
	Deploy() (*Duration, *Int32)
	// 返回false 将停止后续定时器,
	// 返回tru 继续执行
	Run() bool
}
