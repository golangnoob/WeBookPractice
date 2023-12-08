//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webooktrial/interactive/events"
	repository2 "webooktrial/interactive/repository"
	redis2 "webooktrial/interactive/repository/cache/redis"
	dao2 "webooktrial/interactive/repository/dao"
	service2 "webooktrial/interactive/service"
	"webooktrial/internal/events/article"
	"webooktrial/internal/repository"
	article3 "webooktrial/internal/repository/article"
	"webooktrial/internal/repository/cache/redis"
	"webooktrial/internal/repository/dao"
	article2 "webooktrial/internal/repository/dao/article"
	"webooktrial/internal/service"
	"webooktrial/internal/web"
	ijwt "webooktrial/internal/web/jwt"
	"webooktrial/ioc"
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	redis2.NewRedisInteractiveCache,
)

var rankingServiceSet = wire.NewSet(
	repository.NewCachedRankingRepository,
	redis.NewRankingRedisCache,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitRLockClient,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		interactiveSvcProvider,
		rankingServiceSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		// consumer
		events.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		// 初始化 DAO
		dao.NewUserDAO,
		article2.NewGormArticleDao,

		redis.NewUserCache,
		redis.NewCodeCache,
		redis.NewRedisArticleCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		article3.NewArticleRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewOAuth2WechatHandler,
		web.NewUserHandler,
		web.NewArticleHandler,
		//ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
