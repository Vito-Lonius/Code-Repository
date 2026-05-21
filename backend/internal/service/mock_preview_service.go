package service

import (
	"context"
	"reflect"

	"code-repo/internal/model/dto"

	"github.com/golang/mock/gomock"
)

type MockPreviewService struct {
	ctrl     *gomock.Controller
	recorder *MockPreviewServiceMockRecorder
}

type MockPreviewServiceMockRecorder struct {
	mock *MockPreviewService
}

func NewMockPreviewService(ctrl *gomock.Controller) *MockPreviewService {
	mock := &MockPreviewService{ctrl: ctrl}
	mock.recorder = &MockPreviewServiceMockRecorder{mock}
	return mock
}

func (m *MockPreviewService) EXPECT() *MockPreviewServiceMockRecorder {
	return m.recorder
}

func (m *MockPreviewService) PreviewFile(ctx context.Context, fileID uint64) (*dto.PreviewResponse, error) {
	ret := m.ctrl.Call(m, "PreviewFile", ctx, fileID)
	ret0, _ := ret[0].(*dto.PreviewResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPreviewServiceMockRecorder) PreviewFile(ctx, fileID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreviewFile", reflect.TypeOf((*MockPreviewService)(nil).PreviewFile), ctx, fileID)
}

func (m *MockPreviewService) GetDirectoryTree(ctx context.Context, repoID uint64) (*dto.TreeResponse, error) {
	ret := m.ctrl.Call(m, "GetDirectoryTree", ctx, repoID)
	ret0, _ := ret[0].(*dto.TreeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPreviewServiceMockRecorder) GetDirectoryTree(ctx, repoID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDirectoryTree", reflect.TypeOf((*MockPreviewService)(nil).GetDirectoryTree), ctx, repoID)
}

func (m *MockPreviewService) GetRawFile(ctx context.Context, fileID uint64) (*dto.RawFileResponse, error) {
	ret := m.ctrl.Call(m, "GetRawFile", ctx, fileID)
	ret0, _ := ret[0].(*dto.RawFileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPreviewServiceMockRecorder) GetRawFile(ctx, fileID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRawFile", reflect.TypeOf((*MockPreviewService)(nil).GetRawFile), ctx, fileID)
}
