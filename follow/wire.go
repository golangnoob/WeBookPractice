package service

import (
	"github.com/google/wire"

	grpc2 "webooktrial/follow/grpc"
	"webooktrial/follow/ioc"
	"webooktrial/follow/repository"
	"webooktrial/follow/repository/cache"
	"webooktrial/follow/repository/dao"
	"webooktrial/follow/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMFollowRelationDAO,
	repository.NewCachedRelationRepository,
	service.NewFollowRelationService,
	cache.NewRedisFollowCache,
	grpc2.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
