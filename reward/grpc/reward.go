package grpc

import (
	"context"

	"google.golang.org/grpc"

	rewardv1 "webooktrial/api/proto/gen/reward/v1"
	"webooktrial/reward/domain"
	"webooktrial/reward/service"
)

type RewardServiceServer struct {
	rewardv1.UnimplementedRewardServiceServer
	svc service.RewardService
}

func NewRewardServiceServer(svc service.RewardService) *RewardServiceServer {
	return &RewardServiceServer{svc: svc}
}

func (r *RewardServiceServer) Register(server *grpc.Server) {
	rewardv1.RegisterRewardServiceServer(server, r)
}

func (r *RewardServiceServer) PreReward(ctx context.Context, req *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	codeURL, err := r.svc.PreReward(ctx, domain.Reward{
		Uid: req.Uid,
		Target: domain.Target{
			Biz:     req.Biz,
			BizId:   req.BizId,
			BizName: req.BizName,
			Uid:     req.TargetUid,
		},
		Amt: req.Amt,
	})
	return &rewardv1.PreRewardResponse{
		CodeUrl: codeURL.URL,
		Rid:     codeURL.Rid,
	}, err
}

func (r *RewardServiceServer) GetReward(ctx context.Context, req *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	rw, err := r.svc.GetReward(ctx, req.Rid, req.Uid)
	if err != nil {
		return nil, err
	}
	return &rewardv1.GetRewardResponse{
		Status: rewardv1.RewardStatus(rw.Status),
	}, err
}
