// Code generated by MockGen. DO NOT EDIT.
// Source: E:\webooktrial\internal\service\user.go

// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	reflect "reflect"
	domain "webooktrial/internal/domain"

	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// Edit mocks base method.
func (m *MockUserService) Edit(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Edit", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// Edit indicates an expected call of Edit.
func (mr *MockUserServiceMockRecorder) Edit(ctx, u interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Edit", reflect.TypeOf((*MockUserService)(nil).Edit), ctx, u)
}

// FindOrCreate mocks base method.
func (m *MockUserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreate", ctx, phone)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreate indicates an expected call of FindOrCreate.
func (mr *MockUserServiceMockRecorder) FindOrCreate(ctx, phone interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreate", reflect.TypeOf((*MockUserService)(nil).FindOrCreate), ctx, phone)
}

// FindOrCreateByWeChat mocks base method.
func (m *MockUserService) FindOrCreateByWeChat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreateByWeChat", ctx, wechatInfo)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreateByWeChat indicates an expected call of FindOrCreateByWeChat.
func (mr *MockUserServiceMockRecorder) FindOrCreateByWeChat(ctx, wechatInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreateByWeChat", reflect.TypeOf((*MockUserService)(nil).FindOrCreateByWeChat), ctx, wechatInfo)
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, email, password)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, email, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, email, password)
}

// Profile mocks base method.
func (m *MockUserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Profile", ctx, id)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Profile indicates an expected call of Profile.
func (mr *MockUserServiceMockRecorder) Profile(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Profile", reflect.TypeOf((*MockUserService)(nil).Profile), ctx, id)
}

// SignUp mocks base method.
func (m *MockUserService) SignUp(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignUp", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// SignUp indicates an expected call of SignUp.
func (mr *MockUserServiceMockRecorder) SignUp(ctx, u interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignUp", reflect.TypeOf((*MockUserService)(nil).SignUp), ctx, u)
}
