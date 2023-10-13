package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository"
	repomocks "webooktrial/internal/repository/mocks"
	"webooktrial/pkg/logger"
)

func TestUserCoreService_Login(t *testing.T) {
	// 做成一个测试用例都用到的时间
	now := time.Now()

	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository
		//ctx      context.Context
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123456qq@qq.com").
					Return(domain.User{
						Email:    "123456qq@qq.com",
						Password: "$2a$10$4kvm0LUQhCW0xQIK41RYCOr0iGfKoO9dYdrBmNW3qJuFPiFQiYImO",
						Phone:    "1234567891",
						Ctime:    now,
					}, nil)
				return repo
			},
			//ctx: context.Background(),
			email:    "123456qq@qq.com",
			password: "123456@qq",
			wantUser: domain.User{
				Email:    "123456qq@qq.com",
				Password: "$2a$10$4kvm0LUQhCW0xQIK41RYCOr0iGfKoO9dYdrBmNW3qJuFPiFQiYImO",
				Phone:    "1234567891",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123456qq@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123456qq@qq.com",
			password: "123456@qq",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123456qq@qq.com").
					Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},
			email:    "123456qq@qq.com",
			password: "123456@qq",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123456qq@qq.com").
					Return(domain.User{
						Email:    "123456qq@qq.com",
						Password: "$2a$10$4kvm0LUQhCW0xQIK41RYCOr0iGfKoO9dYdrBmNW3qJuFPiFQiYImO",
						Phone:    "1234567891",
						Ctime:    now,
					}, nil)
				return repo
			},
			email:    "123456qq@qq.com",
			password: "123456789@qq",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc := NewUserService(tc.mock(ctrl), &logger.NopLogger{})
			u, err := userSvc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("123456@qq"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
