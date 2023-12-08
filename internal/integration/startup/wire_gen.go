// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"webooktrial/internal/web/jwt"
	"webooktrial/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLog()
	v := ioc.InitMiddlewares(cmdable, handler, loggerV1)
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := redis.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := redis.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := InitPhantomWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	articleDao := article.NewGormArticleDao(gormDB)
	articleCache := redis.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDao, loggerV1, articleCache)
	client := InitKafka()
	syncProducer := NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler(dao2 article.ArticleDao) *web.ArticleHandler {
	loggerV1 := InitLog()
	cmdable := InitRedis()
	articleCache := redis.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(dao2, loggerV1, articleCache)
	client := InitKafka()
	syncProducer := NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	return articleHandler
}

func InitUserSvc() service.UserService {
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := redis.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	loggerV1 := InitLog()
	userService := service.NewUserService(userRepository, loggerV1)
	return userService
}

func InitJwtHdl() jwt.Handler {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	return handler
}

// wire.go:

var thirdProvider = wire.NewSet(InitRedis,
	NewSyncProducer, InitTestDB, InitLog, InitKafka)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, redis.NewUserCache, repository.NewUserRepository, service.NewUserService)

var articleSvcProvider = wire.NewSet(article.NewGormArticleDao, article2.NewArticleRepository, service.NewArticleService, redis.NewRedisArticleCache)
