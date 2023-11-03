package cache

import "context"

//go:generate mockgen -source=./code.go -package=cachemocks -destination=mocks/code.mock.go CodeCache
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}
