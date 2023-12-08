package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"webooktrial/api/proto/gen/intr/v1"
	"webooktrial/interactive/domain"
	"webooktrial/interactive/service"
)

// InteractiveServiceServer 只是把 service 包装成一个 grpc
// 和 grpc 相关的操作都限制在这里
type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	// 核心业务逻辑一定是在 service 里
	svc service.InteractiveService
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	if request.Uid < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "uid 错误")
	}
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	intr, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.GetResponse{Intr: i.toDTO(intr)}, nil
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	intrs, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetIds())
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*intrv1.Interactive, len(intrs))
	for k, v := range intrs {
		m[k] = i.toDTO(v)
	}
	return &intrv1.GetByIdsResponse{Intrs: m}, nil
}

func (i *InteractiveServiceServer) mustEmbedUnimplementedInteractiveServiceServer() {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveServiceServer) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}
