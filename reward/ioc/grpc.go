package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"webooktrial/pkg/grpcx"
	"webooktrial/pkg/logger"
	grpc2 "webooktrial/reward/grpc"
)

func InitGRPCxServer(reward *grpc2.RewardServiceServer,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdTTL   int64    `yaml:"etcdTTL"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	reward.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "reward",
		L:         l,
		EtcdTTL:   cfg.EtcdTTL,
	}
}
