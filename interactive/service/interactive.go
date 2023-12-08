package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"webooktrial/interactive/domain"
	"webooktrial/interactive/repository"
	"webooktrial/pkg/logger"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// Like 点赞
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	// CancelLike 取消点赞
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	// Collect 收藏, cid 是收藏夹的 ID
	// cid 不一定有，或者说 0 对应的是该用户的默认收藏夹
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Get(
	ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	// 按照 repository 的语义(完成 domain.Interactive 的完整构造)，你这里拿到的就应该是包含全部字段的
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		i.l.Error("查询用户是否点赞的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return intr, nil
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 点赞
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

// Collect 收藏
func (i *interactiveService) Collect(ctx context.Context,
	biz string, bizId, cid, uid int64) error {
	// service 还叫做收藏
	// repository
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository,
	l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}
