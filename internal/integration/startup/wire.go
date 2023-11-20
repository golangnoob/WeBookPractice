//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	article3 "webooktrial/internal/events/article"
	"webooktrial/internal/repository"
	article2 "webooktrial/internal/repository/article"
	"webooktrial/internal/repository/cache/redis"
	"webooktrial/internal/repository/dao"
	"webooktrial/internal/repository/dao/article"
	"webooktrial/internal/service"
	"webooktrial/internal/web"
	ijwt "webooktrial/internal/web/jwt"
	"webooktrial/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog, InitKafka)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	redis.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	article.NewGormArticleDao,
	article2.NewArticleRepository,
	service.NewArticleService,
	redis.NewRedisArticleCache,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	redis.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,
		redis.NewCodeCache,
		//article.NewGormArticleDao,
		repository.NewCodeRepository,
		//article2.NewArticleRepository,
		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitSMSService,
		ioc.NewSyncProducer,
		article3.NewKafkaProducer,
		// 指定啥也不干的 wechat service
		InitPhantomWechatService,
		service.NewCodeService,
		//service.NewArticleService,
		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		//InitWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler(dao article.ArticleDao) *web.ArticleHandler {
	wire.Build(thirdProvider,
		//userSvcProvider,
		redis.NewRedisArticleCache,
		//wire.InterfaceValue(new(article.ArticleDAO), dao),
		//article.NewGormArticleDao,
		article2.NewArticleRepository,
		service.NewArticleService,
		ioc.NewSyncProducer,
		article3.NewKafkaProducer,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}
