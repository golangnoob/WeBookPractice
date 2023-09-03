package repository

import (
	"context"

	cache2 "webooktrial/internal/repository/cache"
	"webooktrial/internal/repository/cache/redis"
)

var (
	ErrCodeSendTooMany        = redis.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = redis.ErrCodeVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache2.CodeCache
}

func NewCodeRepository(c cache2.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

func (repo *CachedCodeRepository) Store(ctx context.Context, biz string,
	phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CachedCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
