// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	repository2 "webooktrial/interactive/repository"
	redis2 "webooktrial/interactive/repository/cache/redis"
	dao2 "webooktrial/interactive/repository/dao"
	service2 "webooktrial/interactive/service"
	article3 "webooktrial/internal/events/article"
	"webooktrial/internal/repository"
	article2 "webooktrial/internal/repository/article"
	"webooktrial/internal/repository/cache/local"
	"webooktrial/internal/repository/cache/redis"
	"webooktrial/internal/repository/dao"
	"webooktrial/internal/repository/dao/article"
	"webooktrial/internal/service"
	"webooktrial/internal/web"
	"webooktrial/internal/web/jwt"
	"webooktrial/ioc"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, handler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewUserDAO(db)
	userCache := redis.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := redis.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	articleDao := article.NewGormArticleDao(db)
	articleCache := redis.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDao, loggerV1, articleCache, userRepository)
	client := ioc.InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	clientv3Client := ioc.InitEtcd()
	interactiveServiceClient := ioc.InitIntrGRPCClientV1(clientv3Client)
	articleHandler := web.NewArticleHandler(articleService, loggerV1, interactiveServiceClient)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	v2 := ioc.NewConsumers()
	rankingRedisCache := redis.NewRankingRedisCache(cmdable)
	rankingLocalCache := local.NewRankingLocalCache()
	rankingRepository := repository.NewCachedRankingRepository(rankingRedisCache, rankingLocalCache)
	rankingService := service.NewBatchRankingService(articleService, interactiveServiceClient, rankingRepository)
	rlockClient := ioc.InitRLockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, rlockClient, loggerV1)
	cron := ioc.InitJobs(loggerV1, rankingJob)
	app := &App{
		web:       engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

var interactiveSvcProvider = wire.NewSet(service2.NewInteractiveService, repository2.NewCachedInteractiveRepository, dao2.NewGORMInteractiveDAO, redis2.NewRedisInteractiveCache)

var rankingServiceSet = wire.NewSet(repository.NewCachedRankingRepository, redis.NewRankingRedisCache, local.NewRankingLocalCache, service.NewBatchRankingService)
