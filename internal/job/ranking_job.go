package job

import (
	"context"
	"sync"
	"time"

	rlock "github.com/gotomicro/redis-lock"

	"webooktrial/internal/service"
	"webooktrial/pkg/logger"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	l         logger.LoggerV1
	key       string
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client, l logger.LoggerV1) *RankingJob {
	return &RankingJob{svc: svc,
		timeout:   timeout,
		client:    client,
		l:         l,
		key:       "rlock:cron_job:ranking",
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明没有拿到锁，试着去抢锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这边没拿到锁，极大概率是别人持有了锁
			return err
		}
		r.lock = lock
		// 一直持有锁
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			// 自动续约机制
			er := lock.AutoRefresh(r.timeout/2, time.Second)
			// 退出续约，续约失败
			if er != nil {
				// 记录日志, 下次继续抢锁
				r.l.Error("续约失败", logger.Error(err))
			}
			r.lock = nil
			// lock.Unlock(ctx)
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
