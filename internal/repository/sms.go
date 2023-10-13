package repository

import (
	"context"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/dao"
)

type SMSRepository interface {
	Insert(ctx context.Context, m domain.SMSRetry) error
	GetRetryList(ctx context.Context, status int) ([]domain.SMSRetry, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}

type SMSRepo struct {
	dao dao.SMSDaoInterface
}

func NewSMSRepo(dao dao.SMSDaoInterface) SMSRepository {
	return &SMSRepo{
		dao: dao,
	}
}

func (S *SMSRepo) Insert(ctx context.Context, m domain.SMSRetry) error {
	smsDao := S.domainToEntity(m)
	return S.dao.Insert(ctx, smsDao)

}

func (S *SMSRepo) GetRetryList(ctx context.Context, status int) ([]domain.SMSRetry, error) {
	retrySlice, err := S.dao.Select(ctx, status)
	if err != nil {
		return nil, err
	}
	res := make([]domain.SMSRetry, len(retrySlice))
	for _, sR := range retrySlice {
		res = append(res, S.entityToDomain(sR))
	}
	return res, nil
}

func (S *SMSRepo) UpdateStatus(ctx context.Context, id int64, status int) error {
	return S.dao.Update(ctx, id, status)
}

func (S *SMSRepo) entityToDomain(s dao.SMSMsg) domain.SMSRetry {
	return domain.SMSRetry{
		Id:           s.Id,
		Biz:          s.Biz,
		PhoneNumbers: s.PhoneNumbers,
		Args:         s.Args,
	}
}

func (S *SMSRepo) domainToEntity(s domain.SMSRetry) dao.SMSMsg {
	return dao.SMSMsg{
		Biz:          s.Biz,
		PhoneNumbers: s.PhoneNumbers,
		Args:         s.Args,
	}
}
