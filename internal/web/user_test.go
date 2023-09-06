package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"webooktrial/internal/domain"
	"webooktrial/internal/service"
	svcmocks "webooktrial/internal/service/mocks"
)

func TestEncrypt(t *testing.T) {
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestNil(t *testing.T) {
	testTypeAssert(nil)
}

func testTypeAssert(c any) {
	_, ok := c.(*UserClaims)
	println(ok)
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123456qq@qq.com",
					Password: "123456@qq",
				}).Return(nil)
				// 注册成功返回nil
				return userSvc
			},
			reqBody: `
{
	"email": "123456qq@qq.com",
	"password": "123456@qq",
	"confirm_password": "123456@qq"
}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数不对，bind 失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email": "xqc123@qq.com",
	"password": "xqcQq123456",
	"confirm_password": "xqcQq123456"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email": "xqc123@q",
	"password": "xqcQq123456",
	"confirm_password": "xqcQq123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱格式不对",
		},
		{
			name: "两次输入的密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email": "123456qq@qq.com",
	"password": "123456@qq",
	"confirm_password": "123456qq"
}
`,
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不一致",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello123",
	"confirm_password": "hello123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "密码必须大于8位，包含数字、特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				// 注册成功是 return nil
				return userSvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirm_password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("随便一个 error"))
				// 注册成功是 return nil
				return userSvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirm_password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			// 目前用不上 codeSvc
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是JSON格式
			req.Header.Set("Content-Type", "application/json")
			// 这里可以继续使用req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBody  string
		wantCode int
		wantBody Result
		uid      int64
	}{
		{
			name: "手机号码不合法",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBody: `
{
	"phone": "1234567891",
	"code": "123456" 
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "请输入合法的手机号",
			},
		},
		{
			name: "验证码不合法",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBody: `
{
	"phone": "12345678912",
	"code": "12345" 
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "请输入合法的验证码",
			},
		},
		{
			name: "验证码系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "12345678912", "123456").
					Return(false, errors.New("随便来个错误"))
				return userSvc, codeSvc
			},
			reqBody: `
{
	"phone": "12345678912",
	"code": "123456" 
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "验证码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSVC := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "12345678912", "123456").
					Return(false, nil)
				return userSVC, codeSvc
			},
			reqBody: `
{
	"phone": "12345678912",
	"code": "123456" 
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "验证码错误",
			},
		},
		{
			name: "验证码校验通过，用户登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "12345678912").
					Return(domain.User{
						Phone: "12345678912",
						Id:    6,
					}, nil)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "12345678912", "123456").
					Return(true, nil)
				return userSvc, codeSvc
			},
			reqBody: `
{
	"phone": "12345678912",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Msg: "验证码校验通过",
			},
		},
		{
			name: "验证码校验通过，数据库异常",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "12345678912").
					Return(domain.User{
						Phone: "12345678912",
					}, errors.New("系统内部错误"))
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "12345678912", "123456").
					Return(true, nil)
				return userSvc, codeSvc
			},
			reqBody: `
{
	"phone": "12345678912",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)
			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tokenStr := resp.Header().Get("X-Jwt-Token")
			if tokenStr != "" {
				claims := &UserClaims{}
				_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("bCTF)phY%[u5yA=Wl60mt]Q,SbVRwP!H"), nil
				})
				if err != nil {
					t.Fatal("解析token失败:", err)
				}
				assert.Equal(t, tc.uid, claims.Uid)
			} else {
				t.Log("用户没有登录")
			}
		})
	}
}
