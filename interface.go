package timing

import (
	"time"
)

type CronJob interface {
	Next() (time.Duration, uint)
	Run()
}
