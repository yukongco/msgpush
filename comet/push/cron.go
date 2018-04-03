package push

import (
	"fmt"
	"strconv"
	"time"

	"github.com/yukongco/cron"
)

var (
	CronRes *cron.Cron
)

func InitCron() *cron.Cron {
	CronRes = cron.New()
	return CronRes
}

// get cron express, ex: 30 1 1 15 * ?
func GetCronExpress(putUtcStr string) (string, error) {
	putUtc, err := strconv.ParseInt(putUtcStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("putUtc=%v fmt timestamp err: %v", err)
	}

	putTime := time.Unix(putUtc, 0)
	month := fmt.Sprintf("%v", int(putTime.Month()))
	day := fmt.Sprintf("%v", putTime.Day())
	hour := fmt.Sprintf("%v", putTime.Hour())
	minute := fmt.Sprintf("%v", putTime.Minute())
	second := fmt.Sprintf("%v", putTime.Second())

	express := second + " " + minute + " " + hour + " " + day + " " + month + " ?"
	return express, nil
}
