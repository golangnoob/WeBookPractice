package repository

import (
	"context"

	"webooktrial/follow/domain"
	"webooktrial/follow/repository/cache"
	"webooktrial/follow/repository/dao"
	"webooktrial/pkg/logger"
)

type FollowRepository interface {
	// GetFollowee 获取某人的关注列表
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	// FollowInfo 查看关注人的详情
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type CachedRelationRepository struct {
	dao   dao.FollowRelationDao
	cache cache.FollowCache
	l     logger.LoggerV1
}

func NewCachedRelationRepository(dao dao.FollowRelationDao, cache cache.FollowCache, l logger.LoggerV1) FollowRepository {
	return &CachedRelationRepository{dao: dao, cache: cache, l: l}
}

func (c *CachedRelationRepository) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	followerList, err := c.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return c.genFollowRelationList(followerList), nil
}

func (c *CachedRelationRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	f, err := c.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return c.toDomain(f), nil
}

func (c *CachedRelationRepository) AddFollowRelation(ctx context.Context, f domain.FollowRelation) error {
	err := c.dao.CreateFollowRelation(ctx, c.toEntity(f))
	if err != nil {
		return err
	}
	return c.cache.Follow(ctx, f.Follower, f.Followee)
}

func (c *CachedRelationRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := c.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return c.cache.CancelFollow(ctx, follower, followee)
}

func (c *CachedRelationRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	// 快路径
	res, err := c.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, err
	}
	// 慢路径
	res.Followers, err = c.dao.CntFollower(ctx, uid)
	if err != nil {
		return res, err
	}
	res.Followees, err = c.dao.CntFollowee(ctx, uid)
	if err != nil {
		return res, err
	}
	err = c.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil {
		// 这里记录日志
		c.l.Error("缓存关注统计信息失败",
			logger.Error(err),
			logger.Int64("uid", uid))
	}
	return res, nil
}

func (c *CachedRelationRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(followerList))
	for _, f := range followerList {
		res = append(res, c.toDomain(f))
	}
	return res
}

func (c *CachedRelationRepository) toDomain(fr dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: fr.Followee,
		Follower: fr.Follower,
	}
}

func (c *CachedRelationRepository) toEntity(f domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: f.Followee,
		Follower: f.Follower,
	}
}
