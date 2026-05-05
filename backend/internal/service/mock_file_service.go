package service

import (
	"context"
	"io"
	"mime/multipart"
	"reflect"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"

	"github.com/golang/mock/gomock"
)

type MockFileService struct {
	ctrl     *gomock.Controller
	recorder *MockFileServiceMockRecorder
}

type MockFileServiceMockRecorder struct {
	mock *MockFileService
}

func NewMockFileService(ctrl *gomock.Controller) *MockFileService {
	mock := &MockFileService{ctrl: ctrl}
	mock.recorder = &MockFileServiceMockRecorder{mock}
	return mock
}

func (m *MockFileService) EXPECT() *MockFileServiceMockRecorder {
	return m.recorder
}

func (m *MockFileService) UploadSimple(ctx context.Context, repoID uint64, uploaderID uint64, filePath string, file multipart.File, header *multipart.FileHeader) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "UploadSimple", ctx, repoID, uploaderID, filePath, file, header)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) UploadSimple(ctx, repoID, uploaderID, filePath, file, header interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadSimple", reflect.TypeOf((*MockFileService)(nil).UploadSimple), ctx, repoID, uploaderID, filePath, file, header)
}

func (m *MockFileService) UploadInit(ctx context.Context, uploaderID uint64, req dto.UploadInitRequest) (*dto.UploadInitResponse, error) {
	ret := m.ctrl.Call(m, "UploadInit", ctx, uploaderID, req)
	ret0, _ := ret[0].(*dto.UploadInitResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) UploadInit(ctx, uploaderID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadInit", reflect.TypeOf((*MockFileService)(nil).UploadInit), ctx, uploaderID, req)
}

func (m *MockFileService) UploadChunk(ctx context.Context, req dto.UploadChunkRequest, data []byte) (*dto.UploadChunkResponse, error) {
	ret := m.ctrl.Call(m, "UploadChunk", ctx, req, data)
	ret0, _ := ret[0].(*dto.UploadChunkResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) UploadChunk(ctx, req, data interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadChunk", reflect.TypeOf((*MockFileService)(nil).UploadChunk), ctx, req, data)
}

func (m *MockFileService) UploadComplete(ctx context.Context, req dto.UploadCompleteRequest) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "UploadComplete", ctx, req)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) UploadComplete(ctx, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadComplete", reflect.TypeOf((*MockFileService)(nil).UploadComplete), ctx, req)
}

func (m *MockFileService) GetFileDetail(ctx context.Context, fileID uint64) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "GetFileDetail", ctx, fileID)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) GetFileDetail(ctx, fileID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileDetail", reflect.TypeOf((*MockFileService)(nil).GetFileDetail), ctx, fileID)
}

func (m *MockFileService) ListFiles(ctx context.Context, repoID uint64, parentPath string) (*dto.FileListResponse, error) {
	ret := m.ctrl.Call(m, "ListFiles", ctx, repoID, parentPath)
	ret0, _ := ret[0].(*dto.FileListResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) ListFiles(ctx, repoID, parentPath interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFiles", reflect.TypeOf((*MockFileService)(nil).ListFiles), ctx, repoID, parentPath)
}

func (m *MockFileService) DownloadFile(ctx context.Context, fileID uint64) (io.Reader, *entity.File, error) {
	ret := m.ctrl.Call(m, "DownloadFile", ctx, fileID)
	ret0, _ := ret[0].(io.Reader)
	ret1, _ := ret[1].(*entity.File)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockFileServiceMockRecorder) DownloadFile(ctx, fileID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadFile", reflect.TypeOf((*MockFileService)(nil).DownloadFile), ctx, fileID)
}

func (m *MockFileService) DeleteFile(ctx context.Context, userID uint64, fileID uint64) error {
	ret := m.ctrl.Call(m, "DeleteFile", ctx, userID, fileID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockFileServiceMockRecorder) DeleteFile(ctx, userID, fileID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFile", reflect.TypeOf((*MockFileService)(nil).DeleteFile), ctx, userID, fileID)
}

func (m *MockFileService) CreateDir(ctx context.Context, userID uint64, req dto.CreateDirRequest) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "CreateDir", ctx, userID, req)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) CreateDir(ctx, userID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDir", reflect.TypeOf((*MockFileService)(nil).CreateDir), ctx, userID, req)
}

func (m *MockFileService) RenameFile(ctx context.Context, userID uint64, fileID uint64, newName string) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "RenameFile", ctx, userID, fileID, newName)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) RenameFile(ctx, userID, fileID, newName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RenameFile", reflect.TypeOf((*MockFileService)(nil).RenameFile), ctx, userID, fileID, newName)
}

func (m *MockFileService) MoveFile(ctx context.Context, userID uint64, fileID uint64, newPath string) (*dto.FileResponse, error) {
	ret := m.ctrl.Call(m, "MoveFile", ctx, userID, fileID, newPath)
	ret0, _ := ret[0].(*dto.FileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockFileServiceMockRecorder) MoveFile(ctx, userID, fileID, newPath interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveFile", reflect.TypeOf((*MockFileService)(nil).MoveFile), ctx, userID, fileID, newPath)
}
