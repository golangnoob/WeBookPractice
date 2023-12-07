package cache

import (
	"context"

	"webooktrial/internal/domain"
)

//go:generate mockgen -source=./code.go -package=cachemocks -destination=mocks/code.mock.go CodeCache
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

//type Cache interface {
//	Set(ctx context.Context, key string, val any, exp time.Duration) error
//	Get(ctx context.Context, key string) ekit.AnyValue
//}
//
//type LocalCache struct {
//}
//
//type RedisCache struct {
//}
//
//type DoubleCache struct {
//	local Cache
//	redis Cache
//}
//
//func (d *DoubleCache) Set(ctx context.Context,
//	key string, val any, exp time.Duration) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (d *DoubleCache) Get(ctx context.Context, key string) ekit.AnyValue {
//	//TODO implement me
//	panic("implement me")
//}

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}
