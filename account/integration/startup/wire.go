package startup

import (
	"github.com/google/wire"

	"webooktrial/account/grpc"
	"webooktrial/account/repository"
	"webooktrial/account/repository/dao"
	"webooktrial/account/service"
)

func InitAccountService() *grpc.AccountServiceServer {
	wire.Build(InitTestDB,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer)
	return new(grpc.AccountServiceServer)
}
