//go:build wireinject

package main

import (
	"github.com/google/wire"

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

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		// consumer
		article.NewInteractiveReadEventBatchConsumer,
		article.NewKafkaProducer,

		// 初始化 DAO
		dao.NewUserDAO,
		article2.NewGormArticleDao,
		dao.NewGORMInteractiveDAO,

		redis.NewUserCache,
		redis.NewCodeCache,
		redis.NewRedisInteractiveCache,
		redis.NewRedisArticleCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedInteractiveRepository,
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
