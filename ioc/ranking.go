package ioc

import (
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"

	"webooktrial/internal/job"
	"webooktrial/internal/service"
	"webooktrial/pkg/logger"
)

func InitRankingJob(svc service.RankingService,
	rlockClient *rlock.Client,
	l logger.LoggerV1) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30, rlockClient, l)
}

func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 这里每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
