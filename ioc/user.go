package ioc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	cache "webooktrial/internal/repository/cache/redis"
	"webooktrial/pkg/redisx"
)

//func InitUserHandler(repo repository.UserRepository) service.UserService {
//	l, err := zap.NewDevelopment()
//	if err != nil {
//		panic(err)
//	}
//	return service.NewUserService(repo, )
//}

func InitUserCache(client *redis.Client) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(
		prometheus.SummaryOpts{
			Namespace: "go_study",
			Subsystem: "webook",
			Name:      "gin_http",
			Help:      "统计 GIN 的 HTTP 接口",
			ConstLabels: map[string]string{
				"biz": "user",
			},
		}))
	panic("你别调用")
}
