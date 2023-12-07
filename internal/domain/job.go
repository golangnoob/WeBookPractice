package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Id int64

	// Name 任务名，比如ranking
	Name string
	Cron string
	// Executor 执行器名
	Executor   string
	Cfg        string
	CancelFunc func() error
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
	cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
