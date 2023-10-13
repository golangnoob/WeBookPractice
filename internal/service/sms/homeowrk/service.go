package homeowrk

import (
	"context"
	"errors"
	"sort"
	"time"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository"
	"webooktrial/internal/service/sms"
	"webooktrial/pkg/logger"
)

// 最近三分之一的请求平均响应阈值超过500MS切换服务商发送
// 如果没有符合要求的则按照历史响应排序顺序发送
// 按平均响应时间将服务商排列
// 一条短信发送时间超过1或者发送失败，异步转储到数据库，每一分钟启动一次
// 重试三次以后仍未成功则标记为失败，需要人工介入
// status 1表示未重试 2表示重成功 3表示重试失败

type Service struct {
	svc           sms.Service
	respHistory   []int
	avgRespTime   int
	latestAvgTime int
}

type FailoverService struct {
	svcs []Service
	// 最大重试次数，默认为3次
	retryMax int
	// 平均响应阈值，默认超过500ms则切换服务商
	threshold int
	change    chan struct{}
	ticker    *time.Ticker
	avgTk     *time.Ticker
	smsRepo   repository.SMSRepository
	l         logger.LoggerV1
}

func NewFailoverService(svcs []Service, retryMax, threshold int,
	smsRepo repository.SMSRepository, l logger.LoggerV1) sms.Service {
	if retryMax <= 0 {
		retryMax = 3
	}
	if threshold <= 0 {
		threshold = 500
	}
	failoverService := &FailoverService{
		svcs:      svcs,
		retryMax:  retryMax,
		threshold: threshold,
		change:    make(chan struct{}),
		ticker:    time.NewTicker(time.Minute),
		avgTk:     time.NewTicker(3 * time.Minute),
		smsRepo:   smsRepo,
		l:         l,
	}
	go failoverService.calAvg()
	go failoverService.Async()
	return failoverService
}

func (f *FailoverService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()
	Svc := f.svcs[0]
	ch := make(chan struct{})
	start := time.Now().UnixMilli()
	go func() {
		defer func() {
			ch <- struct{}{}
		}()
		err := Svc.svc.Send(ctx, biz, args, numbers...)
		if err != nil {
			for i := 1; i < len(f.svcs); i++ {
				if f.svcs[i].latestAvgTime > f.threshold {
					continue
				}
				err := f.svcs[i].svc.Send(ctx, biz, args, numbers...)
				if err != nil {
					continue
				}
				f.svcs[0], f.svcs[i] = f.svcs[i], f.svcs[0]
			}
			for i := 1; i < len(f.svcs); i++ {
				err := f.svcs[i].svc.Send(ctx, biz, args, numbers...)
				if err != nil {
					continue
				}
				f.svcs[0], f.svcs[i] = f.svcs[i], f.svcs[0]
			}
		}

	}()
	select {
	// 超时
	case <-ctx.Done():
		end := time.Now().UnixMilli()
		Svc.respHistory = append(Svc.respHistory, int(end-start))
		go func() {
			m := domain.SMSRetry{
				Biz:          biz,
				Args:         args,
				PhoneNumbers: numbers,
			}
			err := f.smsRepo.Insert(context.Background(), m)
			if err != nil {
				f.l.Error("存储重试短信失败", logger.Error(err))
				return
			}
		}()
		return errors.New("短信发送失败")
	case <-ch:
		end := time.Now().UnixMilli()
		Svc.respHistory = append(Svc.respHistory, int(end-start))
		return nil
	}
}

func (f *FailoverService) calAvg() {
	go func() {
		for {
			<-f.change
			sort.Slice(f.svcs, func(i, j int) bool {
				return f.svcs[i].avgRespTime > f.svcs[j].avgRespTime
			})
		}
	}()
	for {
		select {
		case <-f.avgTk.C:
			var changed bool

			for _, Svc := range f.svcs {
				if len(Svc.respHistory) == 0 {
					continue
				}
				changed = true
				var (
					sumTime  int
					sumLTime int
				)
				hLen := len(Svc.respHistory)
				for i := 0; i < hLen; i++ {
					sumTime += Svc.respHistory[i]
				}
				LhTime := Svc.respHistory[hLen/3*2:]
				for _, t := range LhTime {
					sumLTime += t
				}
				avgH := sumTime / hLen
				avgL := sumLTime / hLen / 3
				Svc.avgRespTime = avgH
				Svc.latestAvgTime = avgL
			}
			if changed {
				f.change <- struct{}{}
			}

		}
	}
}

func (f *FailoverService) Async() {
	for {
		select {
		case <-f.ticker.C:
			retryList, err := f.smsRepo.GetRetryList(context.Background(), 1)
			if err != nil {
				f.l.Error("获取重试短信列表失败", logger.Error(err))
			}
			for _, retryMsg := range retryList {
				retryMsg := retryMsg
				go func(retry domain.SMSRetry) {
					serviceLen := len(f.svcs)
					var success bool
					for i := 0; i < f.retryMax; i++ {
						// 防止越界
						Svc := f.svcs[i%serviceLen]
						err := Svc.svc.Send(context.Background(), retryMsg.Biz, retryMsg.Args, retryMsg.PhoneNumbers...)
						if err == nil {
							success = true
						}
					}
					if success {
						err := f.smsRepo.UpdateStatus(context.Background(), retryMsg.Id, 2)
						if err != nil {
							return
						}
						return
					}
					f.l.Error("重试失败", logger.Int64("id", retryMsg.Id))
				}(retryMsg)
			}
		}
	}
}
