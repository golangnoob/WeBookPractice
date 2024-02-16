// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"webooktrial/follow/grpc"
	"webooktrial/follow/repository"
	"webooktrial/follow/repository/cache"
	"webooktrial/follow/repository/dao"
	"webooktrial/follow/service"
)

// Injectors from wire.go:

func InitServer() *grpc.FollowServiceServer {
	gormDB := InitTestDB()
	followRelationDao := dao.NewGORMFollowRelationDAO(gormDB)
	cmdable := InitRedis()
	followCache := cache.NewRedisFollowCache(cmdable)
	loggerV1 := InitLog()
	followRepository := repository.NewCachedRelationRepository(followRelationDao, followCache, loggerV1)
	followRelationService := service.NewFollowRelationService(followRepository)
	followServiceServer := grpc.NewFollowRelationServiceServer(followRelationService)
	return followServiceServer
}