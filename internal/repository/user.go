package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/cache/redis"
	"webooktrial/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
	FindProfile(ctx context.Context, id int64) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache redis.UserCache
}

func NewUserRepository(dao dao.UserDAO, c redis.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email` = ?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) Update(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		ID:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Describe: u.AboutMe,
	})
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

	u = r.entityToDomain(ue)

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

func (r *CachedUserRepository) FindProfile(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.Select(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("查询失败")
	}
	return domain.User{
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		AboutMe:  u.Describe,
	}, nil
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		ID: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 是否有邮箱
			Valid: u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			// 是否有手机号
			Valid: u.Phone != "",
		},
		Password: u.Password,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.ID,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
