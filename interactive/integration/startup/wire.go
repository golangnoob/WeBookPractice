//go:build wireinject

package startup

import (
	"github.com/google/wire"

	repository2 "webooktrial/interactive/repository"
	redis2 "webooktrial/interactive/repository/cache/redis"
	dao2 "webooktrial/interactive/repository/dao"
	service2 "webooktrial/interactive/service"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB, InitLog)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	redis2.NewRedisInteractiveCache,
)

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil, nil)
}
