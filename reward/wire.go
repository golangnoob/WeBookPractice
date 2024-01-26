package reward

import (
	"github.com/google/wire"

	"webooktrial/pkg/wego"
	"webooktrial/reward/grpc"
	"webooktrial/reward/ioc"
	"webooktrial/reward/repository"
	"webooktrial/reward/repository/cache"
	"webooktrial/reward/repository/dao"
	"webooktrial/reward/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis)

func Init() *wego.App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGRPCxServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
