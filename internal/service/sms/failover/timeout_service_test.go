package failover

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"webooktrial/internal/service/sms"
	smsmocks "webooktrial/internal/service/sms/mocks"
)

func TestTimeoutFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []sms.Service
		threshold int32
		// 通过控制私有字段的取值，来模拟各种场景
		idx int32
		cnt int32

		wantErr error
		wantIdx int32
		wantCnt int32
	}{
		{
			name: "超时，但是没连续超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)
				return []sms.Service{svc0}
			},
			threshold: 3,
			wantErr:   context.DeadlineExceeded,
			wantCnt:   1,
			wantIdx:   0,
		},
		{
			name: "触发了切换，切换之后成功了",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			// 重置了
			wantCnt: 0,
			// 切换到了 1
			wantIdx: 1,
		},
		{
			name: "触发了切换，切换之后失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("发送失败"))

				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			// 重置了，因为不是超时错误，所以没有增加
			wantCnt: 0,
			// 切换到了 1
			wantIdx: 1,
			wantErr: errors.New("发送失败"),
		},
		{
			name: "触发了切换，切换之后依旧超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)

				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			// 重置之后超时
			wantCnt: 1,
			// 切换到了 1
			wantIdx: 1,
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, ok := NewTimeoutFailoverSMSService(tc.mock(ctrl), tc.threshold).(*TimeoutFailoverSMSService)
			require.Equal(t, true, ok)
			svc.idx = tc.idx
			svc.cnt = tc.cnt

			err := svc.Send(context.Background(), "testtpl",
				[]string{}, "12345678912")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantIdx, svc.idx)
			assert.Equal(t, tc.wantCnt, svc.cnt)

		})
	}
}
