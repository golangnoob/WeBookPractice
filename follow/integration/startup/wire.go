package startup

import (
	"github.com/google/wire"

	"webooktrial/follow/grpc"
	"webooktrial/follow/repository"
	"webooktrial/follow/repository/cache"
	"webooktrial/follow/repository/dao"
	"webooktrial/follow/service"
)

func InitServer() *grpc.FollowServiceServer {
	wire.Build(
		InitRedis,
		InitLog,
		InitTestDB,
		dao.NewGORMFollowRelationDAO,
		cache.NewRedisFollowCache,
		repository.NewCachedRelationRepository,
		service.NewFollowRelationService,
		grpc.NewFollowRelationServiceServer,
	)
	return new(grpc.FollowServiceServer)
}
