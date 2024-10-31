// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/G-Research/yunikorn-history-server/internal/database/repository (interfaces: Repository)
//
// Generated by this command:
//
//	mockgen -destination=mock_repository.go -package=repository github.com/G-Research/yunikorn-history-server/internal/database/repository Repository
//

// Package repository is a generated GoMock package.
package repository

import (
	context "context"
	reflect "reflect"

	dao "github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	model "github.com/G-Research/yunikorn-history-server/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// DeleteApplicationsNotInIDs mocks base method.
func (m *MockRepository) DeleteApplicationsNotInIDs(arg0 context.Context, arg1 []string, arg2 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApplicationsNotInIDs", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteApplicationsNotInIDs indicates an expected call of DeleteApplicationsNotInIDs.
func (mr *MockRepositoryMockRecorder) DeleteApplicationsNotInIDs(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplicationsNotInIDs", reflect.TypeOf((*MockRepository)(nil).DeleteApplicationsNotInIDs), arg0, arg1, arg2)
}

// DeleteNodesNotInIDs mocks base method.
func (m *MockRepository) DeleteNodesNotInIDs(arg0 context.Context, arg1 []string, arg2 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNodesNotInIDs", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteNodesNotInIDs indicates an expected call of DeleteNodesNotInIDs.
func (mr *MockRepositoryMockRecorder) DeleteNodesNotInIDs(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNodesNotInIDs", reflect.TypeOf((*MockRepository)(nil).DeleteNodesNotInIDs), arg0, arg1, arg2)
}

// DeletePartitionsNotInIDs mocks base method.
func (m *MockRepository) DeletePartitionsNotInIDs(arg0 context.Context, arg1 []string, arg2 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePartitionsNotInIDs", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePartitionsNotInIDs indicates an expected call of DeletePartitionsNotInIDs.
func (mr *MockRepositoryMockRecorder) DeletePartitionsNotInIDs(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePartitionsNotInIDs", reflect.TypeOf((*MockRepository)(nil).DeletePartitionsNotInIDs), arg0, arg1, arg2)
}

// DeleteQueuesNotInIDs mocks base method.
func (m *MockRepository) DeleteQueuesNotInIDs(arg0 context.Context, arg1 string, arg2 []string, arg3 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteQueuesNotInIDs", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteQueuesNotInIDs indicates an expected call of DeleteQueuesNotInIDs.
func (mr *MockRepositoryMockRecorder) DeleteQueuesNotInIDs(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteQueuesNotInIDs", reflect.TypeOf((*MockRepository)(nil).DeleteQueuesNotInIDs), arg0, arg1, arg2, arg3)
}

// GetAllApplications mocks base method.
func (m *MockRepository) GetAllApplications(arg0 context.Context, arg1 ApplicationFilters) ([]*model.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllApplications", arg0, arg1)
	ret0, _ := ret[0].([]*model.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllApplications indicates an expected call of GetAllApplications.
func (mr *MockRepositoryMockRecorder) GetAllApplications(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllApplications", reflect.TypeOf((*MockRepository)(nil).GetAllApplications), arg0, arg1)
}

// GetAllPartitions mocks base method.
func (m *MockRepository) GetAllPartitions(arg0 context.Context, arg1 PartitionFilters) ([]*model.Partition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllPartitions", arg0, arg1)
	ret0, _ := ret[0].([]*model.Partition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllPartitions indicates an expected call of GetAllPartitions.
func (mr *MockRepositoryMockRecorder) GetAllPartitions(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllPartitions", reflect.TypeOf((*MockRepository)(nil).GetAllPartitions), arg0, arg1)
}

// GetAllQueues mocks base method.
func (m *MockRepository) GetAllQueues(arg0 context.Context) ([]*model.Queue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllQueues", arg0)
	ret0, _ := ret[0].([]*model.Queue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllQueues indicates an expected call of GetAllQueues.
func (mr *MockRepositoryMockRecorder) GetAllQueues(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllQueues", reflect.TypeOf((*MockRepository)(nil).GetAllQueues), arg0)
}

// GetApplicationByID mocks base method.
func (m *MockRepository) GetApplicationByID(arg0 context.Context, arg1 string) (*model.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationByID", arg0, arg1)
	ret0, _ := ret[0].(*model.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationByID indicates an expected call of GetApplicationByID.
func (mr *MockRepositoryMockRecorder) GetApplicationByID(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationByID", reflect.TypeOf((*MockRepository)(nil).GetApplicationByID), arg0, arg1)
}

// GetApplicationsHistory mocks base method.
func (m *MockRepository) GetApplicationsHistory(arg0 context.Context, arg1 HistoryFilters) ([]*dao.ApplicationHistoryDAOInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationsHistory", arg0, arg1)
	ret0, _ := ret[0].([]*dao.ApplicationHistoryDAOInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationsHistory indicates an expected call of GetApplicationsHistory.
func (mr *MockRepositoryMockRecorder) GetApplicationsHistory(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationsHistory", reflect.TypeOf((*MockRepository)(nil).GetApplicationsHistory), arg0, arg1)
}

// GetAppsPerPartitionPerQueue mocks base method.
func (m *MockRepository) GetAppsPerPartitionPerQueue(arg0 context.Context, arg1, arg2 string, arg3 ApplicationFilters) ([]*model.Application, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppsPerPartitionPerQueue", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*model.Application)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAppsPerPartitionPerQueue indicates an expected call of GetAppsPerPartitionPerQueue.
func (mr *MockRepositoryMockRecorder) GetAppsPerPartitionPerQueue(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppsPerPartitionPerQueue", reflect.TypeOf((*MockRepository)(nil).GetAppsPerPartitionPerQueue), arg0, arg1, arg2, arg3)
}

// GetContainersHistory mocks base method.
func (m *MockRepository) GetContainersHistory(arg0 context.Context, arg1 HistoryFilters) ([]*dao.ContainerHistoryDAOInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainersHistory", arg0, arg1)
	ret0, _ := ret[0].([]*dao.ContainerHistoryDAOInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContainersHistory indicates an expected call of GetContainersHistory.
func (mr *MockRepositoryMockRecorder) GetContainersHistory(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainersHistory", reflect.TypeOf((*MockRepository)(nil).GetContainersHistory), arg0, arg1)
}

// GetNodeByID mocks base method.
func (m *MockRepository) GetNodeByID(arg0 context.Context, arg1 string) (*model.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeByID", arg0, arg1)
	ret0, _ := ret[0].(*model.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodeByID indicates an expected call of GetNodeByID.
func (mr *MockRepositoryMockRecorder) GetNodeByID(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeByID", reflect.TypeOf((*MockRepository)(nil).GetNodeByID), arg0, arg1)
}

// GetNodesPerPartition mocks base method.
func (m *MockRepository) GetNodesPerPartition(arg0 context.Context, arg1 string, arg2 NodeFilters) ([]*model.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodesPerPartition", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*model.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodesPerPartition indicates an expected call of GetNodesPerPartition.
func (mr *MockRepositoryMockRecorder) GetNodesPerPartition(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodesPerPartition", reflect.TypeOf((*MockRepository)(nil).GetNodesPerPartition), arg0, arg1, arg2)
}

// GetPartitionByID mocks base method.
func (m *MockRepository) GetPartitionByID(arg0 context.Context, arg1 string) (*model.Partition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPartitionByID", arg0, arg1)
	ret0, _ := ret[0].(*model.Partition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPartitionByID indicates an expected call of GetPartitionByID.
func (mr *MockRepositoryMockRecorder) GetPartitionByID(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPartitionByID", reflect.TypeOf((*MockRepository)(nil).GetPartitionByID), arg0, arg1)
}

// GetQueue mocks base method.
func (m *MockRepository) GetQueue(arg0 context.Context, arg1 string) (*model.Queue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetQueue", arg0, arg1)
	ret0, _ := ret[0].(*model.Queue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetQueue indicates an expected call of GetQueue.
func (mr *MockRepositoryMockRecorder) GetQueue(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQueue", reflect.TypeOf((*MockRepository)(nil).GetQueue), arg0, arg1)
}

// GetQueuesInPartition mocks base method.
func (m *MockRepository) GetQueuesInPartition(arg0 context.Context, arg1 string) ([]*model.Queue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetQueuesInPartition", arg0, arg1)
	ret0, _ := ret[0].([]*model.Queue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetQueuesInPartition indicates an expected call of GetQueuesInPartition.
func (mr *MockRepositoryMockRecorder) GetQueuesInPartition(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQueuesInPartition", reflect.TypeOf((*MockRepository)(nil).GetQueuesInPartition), arg0, arg1)
}

// InsertApplication mocks base method.
func (m *MockRepository) InsertApplication(arg0 context.Context, arg1 *model.Application) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertApplication", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertApplication indicates an expected call of InsertApplication.
func (mr *MockRepositoryMockRecorder) InsertApplication(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertApplication", reflect.TypeOf((*MockRepository)(nil).InsertApplication), arg0, arg1)
}

// InsertNode mocks base method.
func (m *MockRepository) InsertNode(arg0 context.Context, arg1 *model.Node) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertNode", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertNode indicates an expected call of InsertNode.
func (mr *MockRepositoryMockRecorder) InsertNode(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertNode", reflect.TypeOf((*MockRepository)(nil).InsertNode), arg0, arg1)
}

// InsertPartition mocks base method.
func (m *MockRepository) InsertPartition(arg0 context.Context, arg1 *model.Partition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertPartition", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertPartition indicates an expected call of InsertPartition.
func (mr *MockRepositoryMockRecorder) InsertPartition(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertPartition", reflect.TypeOf((*MockRepository)(nil).InsertPartition), arg0, arg1)
}

// InsertQueue mocks base method.
func (m *MockRepository) InsertQueue(arg0 context.Context, arg1 *model.Queue) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertQueue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertQueue indicates an expected call of InsertQueue.
func (mr *MockRepositoryMockRecorder) InsertQueue(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertQueue", reflect.TypeOf((*MockRepository)(nil).InsertQueue), arg0, arg1)
}

// UpdateApplication mocks base method.
func (m *MockRepository) UpdateApplication(arg0 context.Context, arg1 *model.Application) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateApplication", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateApplication indicates an expected call of UpdateApplication.
func (mr *MockRepositoryMockRecorder) UpdateApplication(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateApplication", reflect.TypeOf((*MockRepository)(nil).UpdateApplication), arg0, arg1)
}

// UpdateHistory mocks base method.
func (m *MockRepository) UpdateHistory(arg0 context.Context, arg1 []*dao.ApplicationHistoryDAOInfo, arg2 []*dao.ContainerHistoryDAOInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateHistory", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateHistory indicates an expected call of UpdateHistory.
func (mr *MockRepositoryMockRecorder) UpdateHistory(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateHistory", reflect.TypeOf((*MockRepository)(nil).UpdateHistory), arg0, arg1, arg2)
}

// UpdateNode mocks base method.
func (m *MockRepository) UpdateNode(arg0 context.Context, arg1 *model.Node) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNode", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNode indicates an expected call of UpdateNode.
func (mr *MockRepositoryMockRecorder) UpdateNode(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNode", reflect.TypeOf((*MockRepository)(nil).UpdateNode), arg0, arg1)
}

// UpdatePartition mocks base method.
func (m *MockRepository) UpdatePartition(arg0 context.Context, arg1 *model.Partition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePartition", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePartition indicates an expected call of UpdatePartition.
func (mr *MockRepositoryMockRecorder) UpdatePartition(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePartition", reflect.TypeOf((*MockRepository)(nil).UpdatePartition), arg0, arg1)
}

// UpdateQueue mocks base method.
func (m *MockRepository) UpdateQueue(arg0 context.Context, arg1 *model.Queue) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateQueue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateQueue indicates an expected call of UpdateQueue.
func (mr *MockRepositoryMockRecorder) UpdateQueue(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateQueue", reflect.TypeOf((*MockRepository)(nil).UpdateQueue), arg0, arg1)
}
