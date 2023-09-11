package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"webooktrial/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	// 你的服务商
	svcs []sms.Service
	idx  int32
	// 连续超时的个数
	cnt int32

	// 阈值
	// 连续超时超过这个数字，就要切换
	threshold int32
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt >= t.threshold {
		// 这里要切换，新的下标，往后挪了一个
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 说明你切换了

			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现并发，别人换成功了

		idx = newIdx
		//idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&t.cnt, 1)
		return err
	case err == nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// 不知道什么错误

		// 你可以考虑，换下一个，语义则是：
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		return err
	}
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) sms.Service {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}
