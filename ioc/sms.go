package ioc

import (
	"github.com/redis/go-redis/v9"

	"webooktrial/internal/service/sms"
	"webooktrial/internal/service/sms/memory"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 换内存，还是换别的
	//svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
	//	limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	//return retryable.NewService(svc, 3)
	//return metrics.NewPrometheusDecorator(memory.NewService())
	return memory.NewService()
}
