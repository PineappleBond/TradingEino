package cronutil

import (
	"time"

	"github.com/robfig/cron/v3"
)

// GetNextExecutionTime 计算 cron 表达式的下次执行时间
// 支持 6 位 cron 表达式（秒 + 分 + 时 + 日 + 月 + 星期）
func GetNextExecutionTime(spec string) (time.Time, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(spec)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(time.Now()), nil
}
