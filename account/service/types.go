package service

import (
	"context"

	"webooktrial/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
