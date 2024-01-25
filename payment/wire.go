package payment

import (
	"github.com/google/wire"

	"webooktrial/payment/grpc"
	"webooktrial/payment/ioc"
	"webooktrial/payment/repository"
	"webooktrial/payment/repository/dao"
	"webooktrial/payment/web"
	"webooktrial/pkg/wego"
)

func InitApp() *wego.App {
	wire.Build(
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitWechatClient,
		dao.NewPaymentGORMDAO,
		ioc.InitDB,
		repository.NewPaymentRepository,
		grpc.NewWechatServiceServer,
		ioc.InitLogger,
		ioc.InitGRPCServer,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		web.NewWechatHandler,
		ioc.InitGinServer,
		wire.Struct(new(wego.App), "WebServer", "GRPCServer"))
	return new(wego.App)
}
