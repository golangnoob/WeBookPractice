package service

import (
	"context"
	"time"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository"
	"webooktrial/pkg/logger"
)

type JobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	// 返回一个释放的方法，然后调用者调用
	// PreemptV1(ctx context.Context) (domain.Job, func() error,  error)
	// Release
	//Release(ctx context.Context, id int64) error
}

type cronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()

	j.CancelFunc = func() error {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return c.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		// 没有下一次调度
		return c.repo.Stop(ctx, j.Id)
	}
	return c.repo.UpdateNextTime(ctx, j.Id, next)
}

func (c *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 续约怎么个续法？
	// 更新一下更新时间就可以
	// 比如说我们的续约失败逻辑就是：处于 running 状态，但是更新时间在三分钟以前
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}
