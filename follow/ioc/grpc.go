package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webooktrial/follow/grpc"
	"webooktrial/pkg/grpcx"
)

func InitGRPCxServer(followRelation *grpc2.FollowServiceServer) *grpcx.Server {
	type Config struct {
		Addr []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	followRelation.Register(server)
	return &grpcx.Server{
		Server:    server,
		EtcdAddrs: cfg.Addr,
	}
}
