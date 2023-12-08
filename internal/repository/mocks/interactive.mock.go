// Code generated by MockGen. DO NOT EDIT.
// Source: ./interactive.go

// Package repomocks is a generated GoMock package.
package repomocks

import (
	"context"
	"reflect"

	"go.uber.org/mock/gomock"

	"webooktrial/interactive/domain"
)

// MockInteractiveRepository is a mock of InteractiveRepository interface.
type MockInteractiveRepository struct {
	ctrl     *gomock.Controller
	recorder *MockInteractiveRepositoryMockRecorder
}

// MockInteractiveRepositoryMockRecorder is the mock recorder for MockInteractiveRepository.
type MockInteractiveRepositoryMockRecorder struct {
	mock *MockInteractiveRepository
}

// NewMockInteractiveRepository creates a new mock instance.
func NewMockInteractiveRepository(ctrl *gomock.Controller) *MockInteractiveRepository {
	mock := &MockInteractiveRepository{ctrl: ctrl}
	mock.recorder = &MockInteractiveRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInteractiveRepository) EXPECT() *MockInteractiveRepositoryMockRecorder {
	return m.recorder
}

// AddCollectionItem mocks base method.
func (m *MockInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId, cid, uid int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddCollectionItem", ctx, biz, bizId, cid, uid)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCollectionItem indicates an expected call of AddCollectionItem.
func (mr *MockInteractiveRepositoryMockRecorder) AddCollectionItem(ctx, biz, bizId, cid, uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCollectionItem", reflect.TypeOf((*MockInteractiveRepository)(nil).AddCollectionItem), ctx, biz, bizId, cid, uid)
}

// Collected mocks base method.
func (m *MockInteractiveRepository) Collected(ctx context.Context, biz string, id, uid int64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Collected", ctx, biz, id, uid)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Collected indicates an expected call of Collected.
func (mr *MockInteractiveRepositoryMockRecorder) Collected(ctx, biz, id, uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Collected", reflect.TypeOf((*MockInteractiveRepository)(nil).Collected), ctx, biz, id, uid)
}

// DecrLike mocks base method.
func (m *MockInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId, uid int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DecrLike", ctx, biz, bizId, uid)
	ret0, _ := ret[0].(error)
	return ret0
}

// DecrLike indicates an expected call of DecrLike.
func (mr *MockInteractiveRepositoryMockRecorder) DecrLike(ctx, biz, bizId, uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecrLike", reflect.TypeOf((*MockInteractiveRepository)(nil).DecrLike), ctx, biz, bizId, uid)
}

// Get mocks base method.
func (m *MockInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, biz, bizId)
	ret0, _ := ret[0].(domain.Interactive)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockInteractiveRepositoryMockRecorder) Get(ctx, biz, bizId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockInteractiveRepository)(nil).Get), ctx, biz, bizId)
}

// IncrLike mocks base method.
func (m *MockInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId, uid int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncrLike", ctx, biz, bizId, uid)
	ret0, _ := ret[0].(error)
	return ret0
}

// IncrLike indicates an expected call of IncrLike.
func (mr *MockInteractiveRepositoryMockRecorder) IncrLike(ctx, biz, bizId, uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrLike", reflect.TypeOf((*MockInteractiveRepository)(nil).IncrLike), ctx, biz, bizId, uid)
}

// IncrReadCnt mocks base method.
func (m *MockInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncrReadCnt", ctx, biz, bizId)
	ret0, _ := ret[0].(error)
	return ret0
}

// IncrReadCnt indicates an expected call of IncrReadCnt.
func (mr *MockInteractiveRepositoryMockRecorder) IncrReadCnt(ctx, biz, bizId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrReadCnt", reflect.TypeOf((*MockInteractiveRepository)(nil).IncrReadCnt), ctx, biz, bizId)
}

// Liked mocks base method.
func (m *MockInteractiveRepository) Liked(ctx context.Context, biz string, id, uid int64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Liked", ctx, biz, id, uid)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Liked indicates an expected call of Liked.
func (mr *MockInteractiveRepositoryMockRecorder) Liked(ctx, biz, id, uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Liked", reflect.TypeOf((*MockInteractiveRepository)(nil).Liked), ctx, biz, id, uid)
}
