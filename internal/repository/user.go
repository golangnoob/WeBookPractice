package repository

import (
	"context"
	"errors"
	"webooktrial/internal/domain"
	"webooktrial/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email` = ?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.ID,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) Update(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		ID:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.BirthDay,
		Describe: u.Describe,
	})
}

func (r *UserRepository) FindById(int64) {
	// 先从 cache 里面找
	// 再从 dao 里面找
	// 找到了回写 cache
}

func (r UserRepository) Select(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.Select(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("查询失败")
	}
	return domain.User{
		Nickname: u.Nickname,
		BirthDay: u.Birthday,
		Describe: u.Describe,
	}, nil
}
