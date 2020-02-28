# timing  
[![GoDoc](https://godoc.org/github.com/thinkgos/timing?status.svg)](https://godoc.org/github.com/thinkgos/timing)
[![Build Status](https://travis-ci.org/thinkgos/timing.svg?branch=master)](https://travis-ci.org/thinkgos/timing)
[![codecov](https://codecov.io/gh/thinkgos/timing/branch/master/graph/badge.svg)](https://codecov.io/gh/thinkgos/timing)
![Action Status](https://github.com/thinkgos/timing/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/timing)](https://goreportcard.com/report/github.com/thinkgos/timing)
[![Licence](https://img.shields.io/github/license/thinkgos/timing)](https://raw.githubusercontent.com/thinkgos/timing/master/LICENSE)  
 - 实现hash时间定时器,时间轮定时器
 - 实现简单时间调度,任务处理
 - 任务默认在回调中处理,任务频繁却又不耗时. 可以配置使用goroutine处理
 - 默认时基精度100ms,默认条目时间间隔1ms
## hash map timer
 - 插入,删除,修改时间复杂度o(1),扫描超时条目时间复杂度o(n)
 - 不限最大时间
