package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webooktrial/interactive/grpc"
	"webooktrial/pkg/grpcx"
	"webooktrial/pkg/logger"
)

func InitGRPCxServer(l logger.LoggerV1,
	intrServer *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	intrServer.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "interactive",
		L:         l,
	}
}
