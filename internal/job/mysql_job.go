package job

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"

	"webooktrial/internal/domain"
	"webooktrial/internal/service"
	"webooktrial/pkg/logger"
)

type Executor interface {
	// Name 返回 Executor 的名称
	Name() string
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	// 具体实现来控制
	// 真正去执行一个任务
	Exec(ctx context.Context, j domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了？ %s", j.Name)
	}
	return fn(ctx, j)
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

// Scheduler 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.LoggerV1
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{svc: svc, l: l,
		limiter: semaphore.NewWeighted(200),
		execs:   make(map[string]Executor)}
}
func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 退出调度循环
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 单独执行一次数据库查询的时间
		dbCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 继续下一轮抢占
			s.l.Error("抢占任务失败", logger.Error(err))
		}
		exec, ok := s.execs[j.Executor]
		if !ok {
			// DEBUG 的时候最好中断
			// 线上就继续
			s.l.Error("未找到对应的执行器",
				logger.String("executor", j.Executor))
			continue
		}

		// 执行任务
		go func() {
			defer func() {
				s.limiter.Release(1)
				er := j.CancelFunc()
				if er != nil {
					s.l.Error("释放任务失败",
						logger.Error(er),
						logger.Int64("jid", j.Id))
				}
			}()
			er := exec.Exec(ctx, j)
			if er != nil {
				// 也可以考虑在这里重试
				s.l.Error("任务执行失败", logger.Error(er))
			}
			// 要不要考虑下一次调度？
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(er))
			}
		}()
	}
}
