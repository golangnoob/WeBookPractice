package service

import (
	"context"

	"webooktrial/account/domain"
	"webooktrial/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	return a.repo.AddCredit(ctx, cr)
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}
