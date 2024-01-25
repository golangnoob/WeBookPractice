package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webooktrial/payment/grpc"
	"webooktrial/pkg/grpcx"
	"webooktrial/pkg/grpcx/interceptors/logging"
	"webooktrial/pkg/logger"
)

func InitGRPCServer(wesvc *grpc2.WechatServiceServer,
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
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.NewInterceptorBuilder(l).BuildUnaryServerInterceptor()))
	wesvc.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "payment",
		L:         l,
		EtcdTTL:   cfg.EtcdTTL,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
