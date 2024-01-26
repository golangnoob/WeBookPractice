package account

import (
	"github.com/google/wire"

	"webooktrial/account/grpc"
	"webooktrial/account/ioc"
	"webooktrial/account/repository"
	"webooktrial/account/repository/dao"
	"webooktrial/account/service"
	"webooktrial/pkg/wego"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitGRPCxServer,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"))
	return new(wego.App)
}
