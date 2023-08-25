package repository

import (
	"context"
	"errors"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/cache"
	"webooktrial/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDao, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
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

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 里面找
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 必然是有数据
		return u, nil
	}
	// 没这个数据
	//if err == cache.ErrKeyNotExist {
	// 去数据库里面加载
	//}
	// 再从 dao 里面找
	// 找到了回写 cache

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.ID,
		Email:    ue.Email,
		Password: ue.Password,
	}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			// 我这里怎么办？
			// 打日志，做监控
			//return domain.User{}, err
		}
	}()
	return u, err

	// 这里怎么办？ err = io.EOF
	// 要不要去数据库加载？
	// 看起来我不应该加载？
	// 看起来我好像也要加载？

	// 选加载 —— 做好兜底，万一 Redis 真的崩了，你要保护住你的数据库
	// 我数据库限流呀！

	// 选不加载 —— 用户体验差一点

	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，你也不知道有没有数据
}

func (r *UserRepository) Select(ctx context.Context, id int64) (domain.User, error) {
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
