//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webooktrial/interactive/events"
	"webooktrial/interactive/grpc"
	"webooktrial/interactive/ioc"
	"webooktrial/interactive/repository"
	cache "webooktrial/interactive/repository/cache/redis"
	"webooktrial/interactive/repository/dao"
	"webooktrial/interactive/service"
)

var thirdPartySet = wire.NewSet(ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	// 暂时不理会 consumer 怎么启动
	ioc.InitRedis)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitApp() *App {
	wire.Build(interactiveSvcProvider,
		thirdPartySet,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
