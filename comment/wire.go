package comment

import (
	"github.com/google/wire"

	"webooktrial/comment/grpc"
	"webooktrial/comment/ioc"
	"webooktrial/comment/repository"
	"webooktrial/comment/repository/dao"
	"webooktrial/comment/service"
	"webooktrial/pkg/wego"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentService,
	grpc.NewCommentServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
