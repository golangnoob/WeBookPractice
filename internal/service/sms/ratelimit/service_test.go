package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"webooktrial/internal/service/sms"
	smsmocks "webooktrial/internal/service/sms/mocks"
	"webooktrial/pkg/ratelimit"
	limitmocks "webooktrial/pkg/ratelimit/mocks"
)

func TestRatelimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (ratelimit.Limiter, sms.Service)
		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (ratelimit.Limiter, sms.Service) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), "testtpl", []string{}, "12345678912").
					Return(nil)
				return limiter, svc
			},
			wantErr: nil,
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (ratelimit.Limiter, sms.Service) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return limiter, svc
			},
			wantErr: errors.New("触发了限流"),
		},
		{
			name: "限流器异常",
			mock: func(ctrl *gomock.Controller) (ratelimit.Limiter, sms.Service) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("限流器异常"))
				return limiter, svc
			},
			wantErr: fmt.Errorf("短信服务判断是否限流出现问题，%w", errors.New("限流器异常")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limiter, svc := tc.mock(ctrl)
			limitSvc := NewRatelimitSMSService(svc, limiter)
			err := limitSvc.Send(context.Background(), "testtpl", []string{}, "12345678912")
			assert.Equal(t, tc.wantErr, err)
		})
	}

}
