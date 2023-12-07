package repository

import (
	"context"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/cache/local"
	"webooktrial/internal/repository/cache/redis"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	// 使用具体实现，可读性更好，对测试不友好，因为咩有面向接口编程
	redis *redis.RankingRedisCache
	local *local.RankingLocalCache
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 先放入本地缓存，因为优先查询本地缓存并且本地缓存几乎不可能失败
	_ = c.local.Set(ctx, arts)
	return c.redis.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	// 获取热榜优先从本地缓存获取
	data, err := c.local.Get(ctx)
	if err == nil {
		return data, nil
	}
	data, err = c.redis.Get(ctx)
	if err == nil {
		// 在这里将热榜数据塞到本地缓存
		c.local.Set(ctx, data)
	} else {
		return c.local.ForceGet(ctx)
	}
	return data, err
}

func NewCachedRankingRepository(
	redis *redis.RankingRedisCache,
	local *local.RankingLocalCache,
) RankingRepository {
	return &CachedRankingRepository{local: local, redis: redis}
}
