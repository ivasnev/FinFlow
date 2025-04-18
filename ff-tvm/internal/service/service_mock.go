// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package service is a generated GoMock package.
package service

import (
	context "context"
	ed25519 "crypto/ed25519"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockServiceRepository is a mock of ServiceRepository interface.
type MockServiceRepository struct {
	ctrl     *gomock.Controller
	recorder *MockServiceRepositoryMockRecorder
}

// MockServiceRepositoryMockRecorder is the mock recorder for MockServiceRepository.
type MockServiceRepositoryMockRecorder struct {
	mock *MockServiceRepository
}

// NewMockServiceRepository creates a new mock instance.
func NewMockServiceRepository(ctrl *gomock.Controller) *MockServiceRepository {
	mock := &MockServiceRepository{ctrl: ctrl}
	mock.recorder = &MockServiceRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceRepository) EXPECT() *MockServiceRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockServiceRepository) Create(ctx context.Context, service *Service) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, service)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockServiceRepositoryMockRecorder) Create(ctx, service interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockServiceRepository)(nil).Create), ctx, service)
}

// GetByID mocks base method.
func (m *MockServiceRepository) GetByID(ctx context.Context, id int) (*Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockServiceRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockServiceRepository)(nil).GetByID), ctx, id)
}

// GetPrivateKeyHash mocks base method.
func (m *MockServiceRepository) GetPrivateKeyHash(ctx context.Context, id int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPrivateKeyHash", ctx, id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPrivateKeyHash indicates an expected call of GetPrivateKeyHash.
func (mr *MockServiceRepositoryMockRecorder) GetPrivateKeyHash(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPrivateKeyHash", reflect.TypeOf((*MockServiceRepository)(nil).GetPrivateKeyHash), ctx, id)
}

// GetPublicKey mocks base method.
func (m *MockServiceRepository) GetPublicKey(ctx context.Context, id int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublicKey", ctx, id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublicKey indicates an expected call of GetPublicKey.
func (mr *MockServiceRepositoryMockRecorder) GetPublicKey(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublicKey", reflect.TypeOf((*MockServiceRepository)(nil).GetPublicKey), ctx, id)
}

// GrantAccess mocks base method.
func (m *MockServiceRepository) GrantAccess(ctx context.Context, from, to int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GrantAccess", ctx, from, to)
	ret0, _ := ret[0].(error)
	return ret0
}

// GrantAccess indicates an expected call of GrantAccess.
func (mr *MockServiceRepositoryMockRecorder) GrantAccess(ctx, from, to interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GrantAccess", reflect.TypeOf((*MockServiceRepository)(nil).GrantAccess), ctx, from, to)
}

// RevokeAccess mocks base method.
func (m *MockServiceRepository) RevokeAccess(ctx context.Context, from, to int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RevokeAccess", ctx, from, to)
	ret0, _ := ret[0].(error)
	return ret0
}

// RevokeAccess indicates an expected call of RevokeAccess.
func (mr *MockServiceRepositoryMockRecorder) RevokeAccess(ctx, from, to interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeAccess", reflect.TypeOf((*MockServiceRepository)(nil).RevokeAccess), ctx, from, to)
}

// MockKeyManager is a mock of KeyManager interface.
type MockKeyManager struct {
	ctrl     *gomock.Controller
	recorder *MockKeyManagerMockRecorder
}

// MockKeyManagerMockRecorder is the mock recorder for MockKeyManager.
type MockKeyManagerMockRecorder struct {
	mock *MockKeyManager
}

// NewMockKeyManager creates a new mock instance.
func NewMockKeyManager(ctrl *gomock.Controller) *MockKeyManager {
	mock := &MockKeyManager{ctrl: ctrl}
	mock.recorder = &MockKeyManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyManager) EXPECT() *MockKeyManagerMockRecorder {
	return m.recorder
}

// GenerateKeyPair mocks base method.
func (m *MockKeyManager) GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateKeyPair")
	ret0, _ := ret[0].(ed25519.PublicKey)
	ret1, _ := ret[1].(ed25519.PrivateKey)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GenerateKeyPair indicates an expected call of GenerateKeyPair.
func (mr *MockKeyManagerMockRecorder) GenerateKeyPair() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateKeyPair", reflect.TypeOf((*MockKeyManager)(nil).GenerateKeyPair))
}

// Sign mocks base method.
func (m *MockKeyManager) Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sign", data, privateKey)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Sign indicates an expected call of Sign.
func (mr *MockKeyManagerMockRecorder) Sign(data, privateKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sign", reflect.TypeOf((*MockKeyManager)(nil).Sign), data, privateKey)
}

// Verify mocks base method.
func (m *MockKeyManager) Verify(data, signature []byte, publicKey ed25519.PublicKey) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", data, signature, publicKey)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Verify indicates an expected call of Verify.
func (mr *MockKeyManagerMockRecorder) Verify(data, signature, publicKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockKeyManager)(nil).Verify), data, signature, publicKey)
}

// MockAccessManager is a mock of AccessManager interface.
type MockAccessManager struct {
	ctrl     *gomock.Controller
	recorder *MockAccessManagerMockRecorder
}

// MockAccessManagerMockRecorder is the mock recorder for MockAccessManager.
type MockAccessManagerMockRecorder struct {
	mock *MockAccessManager
}

// NewMockAccessManager creates a new mock instance.
func NewMockAccessManager(ctrl *gomock.Controller) *MockAccessManager {
	mock := &MockAccessManager{ctrl: ctrl}
	mock.recorder = &MockAccessManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessManager) EXPECT() *MockAccessManagerMockRecorder {
	return m.recorder
}

// CheckAccess mocks base method.
func (m *MockAccessManager) CheckAccess(from, to int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckAccess", from, to)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckAccess indicates an expected call of CheckAccess.
func (mr *MockAccessManagerMockRecorder) CheckAccess(from, to interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckAccess", reflect.TypeOf((*MockAccessManager)(nil).CheckAccess), from, to)
}

// GrantAccess mocks base method.
func (m *MockAccessManager) GrantAccess(from, to int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GrantAccess", from, to)
	ret0, _ := ret[0].(error)
	return ret0
}

// GrantAccess indicates an expected call of GrantAccess.
func (mr *MockAccessManagerMockRecorder) GrantAccess(from, to interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GrantAccess", reflect.TypeOf((*MockAccessManager)(nil).GrantAccess), from, to)
}

// RevokeAccess mocks base method.
func (m *MockAccessManager) RevokeAccess(from, to int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RevokeAccess", from, to)
	ret0, _ := ret[0].(error)
	return ret0
}

// RevokeAccess indicates an expected call of RevokeAccess.
func (mr *MockAccessManagerMockRecorder) RevokeAccess(from, to interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeAccess", reflect.TypeOf((*MockAccessManager)(nil).RevokeAccess), from, to)
}

// MockTicketService is a mock of TicketService interface.
type MockTicketService struct {
	ctrl     *gomock.Controller
	recorder *MockTicketServiceMockRecorder
}

// MockTicketServiceMockRecorder is the mock recorder for MockTicketService.
type MockTicketServiceMockRecorder struct {
	mock *MockTicketService
}

// NewMockTicketService creates a new mock instance.
func NewMockTicketService(ctrl *gomock.Controller) *MockTicketService {
	mock := &MockTicketService{ctrl: ctrl}
	mock.recorder = &MockTicketServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTicketService) EXPECT() *MockTicketServiceMockRecorder {
	return m.recorder
}

// GenerateTicket mocks base method.
func (m *MockTicketService) GenerateTicket(ctx context.Context, from, to int, secret string) (*Ticket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateTicket", ctx, from, to, secret)
	ret0, _ := ret[0].(*Ticket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateTicket indicates an expected call of GenerateTicket.
func (mr *MockTicketServiceMockRecorder) GenerateTicket(ctx, from, to, secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateTicket", reflect.TypeOf((*MockTicketService)(nil).GenerateTicket), ctx, from, to, secret)
}
