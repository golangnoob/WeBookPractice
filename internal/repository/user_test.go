package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/cache/redis"
	cachemocks "webooktrial/internal/repository/cache/redis/mocks"
	"webooktrial/internal/repository/dao"
	daomocks "webooktrial/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	now = time.UnixMilli(now.UnixMilli())

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, redis.UserCache)
		ctx      context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中， 查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, redis.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{}, redis.ErrKeyNotExist)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{
						ID: 123,
						Email: sql.NullString{
							String: "123456qq@qq.com",
							Valid:  true,
						},
						Password: "random123",
						Phone: sql.NullString{
							String: "1234567891",
							Valid:  true,
						},
						Ctime: now.UnixMilli(),
						Utime: now.UnixMilli(),
					}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123456qq@qq.com",
					Password: "random123",
					Phone:    "1234567891",
					Ctime:    now,
				}).Return(nil)
				return d, c
			},
			id: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123456qq@qq.com",
				Password: "random123",
				Phone:    "1234567891",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, redis.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{
						Id:       123,
						Email:    "123456qq@qq.com",
						Password: "random123",
						Phone:    "1234567891",
						Ctime:    now,
					}, nil)
				d := daomocks.NewMockUserDAO(ctrl)
				return d, c
			},
			id: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123456qq@qq.com",
				Password: "random123",
				Phone:    "1234567891",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, redis.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{}, redis.ErrKeyNotExist)

				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{}, errors.New("mock db 错误"))
				return d, c
			},
			ctx:      context.Background(),
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
			time.Sleep(time.Second)
		})
	}
}
