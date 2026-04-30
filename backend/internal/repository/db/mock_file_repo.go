package db

import (
	"code-repo/internal/model/entity"
	"reflect"

	"github.com/golang/mock/gomock"
)

type MockFileRepository struct {
	ctrl     *gomock.Controller
	recorder *MockFileRepositoryMockRecorder
}

type MockFileRepositoryMockRecorder struct {
	mock *MockFileRepository
}

func NewMockFileRepository(ctrl *gomock.Controller) *MockFileRepository {
	mock := &MockFileRepository{ctrl: ctrl}
	mock.recorder = &MockFileRepositoryMockRecorder{mock}
	return mock
}

func (m *MockFileRepository) EXPECT() *MockFileRepositoryMockRecorder {
	return m.recorder
}

func (m *MockFileRepository) Create(file *entity.File) error {
	ret := m.ctrl.Call(m, "Create", file)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockFileRepositoryMockRecorder) Create(file interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockFileRepository)(nil).Create), file)
}

func (m *MockFileRepository) GetByID(id uint64) (*entity.File, error) {
	ret := m.ctrl.Call(m, "GetByID", id)
	ret0, _ := ret[0].(*entity.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileRepositoryMockRecorder) GetByID(id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockFileRepository)(nil).GetByID), id)
}

func (m *MockFileRepository) GetByPath(repoID uint64, path string) (*entity.File, error) {
	ret := m.ctrl.Call(m, "GetByPath", repoID, path)
	ret0, _ := ret[0].(*entity.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileRepositoryMockRecorder) GetByPath(repoID, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPath", reflect.TypeOf((*MockFileRepository)(nil).GetByPath), repoID, path)
}

func (m *MockFileRepository) ListByRepo(repoID uint64, parentPath string) ([]entity.File, error) {
	ret := m.ctrl.Call(m, "ListByRepo", repoID, parentPath)
	ret0, _ := ret[0].([]entity.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileRepositoryMockRecorder) ListByRepo(repoID, parentPath interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByRepo", reflect.TypeOf((*MockFileRepository)(nil).ListByRepo), repoID, parentPath)
}

func (m *MockFileRepository) Update(file *entity.File) error {
	ret := m.ctrl.Call(m, "Update", file)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockFileRepositoryMockRecorder) Update(file interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockFileRepository)(nil).Update), file)
}

func (m *MockFileRepository) Delete(id uint64) error {
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockFileRepositoryMockRecorder) Delete(id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockFileRepository)(nil).Delete), id)
}

func (m *MockFileRepository) DeleteByRepo(repoID uint64) error {
	ret := m.ctrl.Call(m, "DeleteByRepo", repoID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockFileRepositoryMockRecorder) DeleteByRepo(repoID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByRepo", reflect.TypeOf((*MockFileRepository)(nil).DeleteByRepo), repoID)
}

type MockUploadTaskRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUploadTaskRepositoryMockRecorder
}

type MockUploadTaskRepositoryMockRecorder struct {
	mock *MockUploadTaskRepository
}

func NewMockUploadTaskRepository(ctrl *gomock.Controller) *MockUploadTaskRepository {
	mock := &MockUploadTaskRepository{ctrl: ctrl}
	mock.recorder = &MockUploadTaskRepositoryMockRecorder{mock}
	return mock
}

func (m *MockUploadTaskRepository) EXPECT() *MockUploadTaskRepositoryMockRecorder {
	return m.recorder
}

func (m *MockUploadTaskRepository) Create(task *entity.UploadTask) error {
	ret := m.ctrl.Call(m, "Create", task)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUploadTaskRepositoryMockRecorder) Create(task interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUploadTaskRepository)(nil).Create), task)
}

func (m *MockUploadTaskRepository) GetByUploadID(uploadID string) (*entity.UploadTask, error) {
	ret := m.ctrl.Call(m, "GetByUploadID", uploadID)
	ret0, _ := ret[0].(*entity.UploadTask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUploadTaskRepositoryMockRecorder) GetByUploadID(uploadID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUploadID", reflect.TypeOf((*MockUploadTaskRepository)(nil).GetByUploadID), uploadID)
}

func (m *MockUploadTaskRepository) UpdateUploadedChunks(uploadID string, uploadedChunks int, indices string) error {
	ret := m.ctrl.Call(m, "UpdateUploadedChunks", uploadID, uploadedChunks, indices)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUploadTaskRepositoryMockRecorder) UpdateUploadedChunks(uploadID, uploadedChunks, indices interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUploadedChunks", reflect.TypeOf((*MockUploadTaskRepository)(nil).UpdateUploadedChunks), uploadID, uploadedChunks, indices)
}

func (m *MockUploadTaskRepository) UpdateStatus(uploadID string, status string) error {
	ret := m.ctrl.Call(m, "UpdateStatus", uploadID, status)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUploadTaskRepositoryMockRecorder) UpdateStatus(uploadID, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockUploadTaskRepository)(nil).UpdateStatus), uploadID, status)
}

func (m *MockUploadTaskRepository) Delete(uploadID string) error {
	ret := m.ctrl.Call(m, "Delete", uploadID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUploadTaskRepositoryMockRecorder) Delete(uploadID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUploadTaskRepository)(nil).Delete), uploadID)
}
