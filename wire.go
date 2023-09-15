//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"webooktrial/internal/repository"
	"webooktrial/internal/repository/cache/redis"
	"webooktrial/internal/repository/dao"
	"webooktrial/internal/service"
	"webooktrial/internal/web"
	ijwt "webooktrial/internal/web/jwt"
	"webooktrial/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,

		// 初始化 DAO
		dao.NewUserDAO,

		redis.NewUserCache,
		redis.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewOAuth2WechatHandler,
		web.NewUserHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
