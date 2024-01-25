package grpcx

import (
	"context"
	"net"
	"strconv"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"

	"webooktrial/pkg/logger"
	"webooktrial/pkg/netx"
)

type Server struct {
	*grpc.Server
	Port      int
	EtcdAddrs []string
	Name      string
	L         logger.LoggerV1
	kaCancel  func()
	em        endpoints.Manager
	// ETCD 服务注册租约 TTL
	EtcdTTL    int64
	EtcdClient *etcdv3.Client
	EtcdKey    string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: s.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	s.EtcdClient = client
	// endpoint 以服务为维度。一个服务一个 Manager
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	if err != nil {
		return err
	}
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + addr
	s.EtcdKey = key
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leaseResp, err := client.Grant(ctx, s.EtcdTTL)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))

	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return nil
}

// Close 你可以叫做 Shutdown
func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.em.DeleteEndpoint(ctx, s.EtcdKey)
		if err != nil {
			return err
		}
	}
	if s.EtcdClient != nil {
		err := s.EtcdClient.Close()
		if err != nil {
			return err
		}
	}
	s.Server.GracefulStop()
	return nil
}
