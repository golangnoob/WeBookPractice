package startup

import (
	"github.com/google/wire"

	"webooktrial/payment/ioc"
	"webooktrial/payment/repository"
	"webooktrial/payment/repository/dao"
	"webooktrial/payment/service/wechat"
)

var thirdPartySet = wire.NewSet(ioc.InitLogger, InitTestDB)

var wechatNativeSvcSet = wire.NewSet(
	ioc.InitWechatClient,
	dao.NewPaymentGORMDAO,
	repository.NewPaymentRepository,
	ioc.InitWechatNativeService,
	ioc.InitWechatConfig)

func InitWechatNativeService() *wechat.NativePaymentService {
	wire.Build(wechatNativeSvcSet, thirdPartySet)
	return new(wechat.NativePaymentService)
}
