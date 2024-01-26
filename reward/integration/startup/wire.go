package startup

import (
	"github.com/google/wire"

	pmtv1 "webooktrial/api/proto/gen/payment/v1"
	"webooktrial/reward/repository"
	"webooktrial/reward/repository/cache"
	"webooktrial/reward/repository/dao"
	"webooktrial/reward/service"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) *service.WechatNativeRewardService {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository, dao.NewRewardGORMDAO)
	return new(service.WechatNativeRewardService)
}
