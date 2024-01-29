package ioc

import (
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	grpc2 "webooktrial/comment/grpc"
	"webooktrial/pkg/grpcx"
	"webooktrial/pkg/logger"
)

func InitGRPCxServer(comment *grpc2.CommentServiceServer,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
		EtcdTTL   int64    `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	comment.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "comment",
		L:         l,
		EtcdTTL:   cfg.EtcdTTL,
	}
}
