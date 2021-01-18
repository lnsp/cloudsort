// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.12.4
// source: data.proto

package pb

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TaskType int32

const (
	// SORT is the first stage in a job. During sort,
	// the worker fetch data from S3 and sort it.
	TaskType_SORT TaskType = 0
	// SAMPLE is the second stage. Each worker computes a local histogram
	// using sampled data from its minibatch.
	TaskType_SAMPLE TaskType = 1
	// SHUFFLE_SEND is the third stage in a job. Each worker sends a part of its sorted data
	// to the respective peers.
	TaskType_SHUFFLE_SEND TaskType = 2
	// SHUFFLE_RECV is the fourth stage in a job. Each worker receives sorted data from its peers and
	// combines them into a single file.
	TaskType_SHUFFLE_RECV TaskType = 3
	// FLUSH is the fifth stage and marks the end of the combined SHUFFLE/MERGE stage.
	TaskType_FLUSH TaskType = 4
	// UPLOAD is the final stage. During upload, the merge result is pushed to S3.
	// After UPLOAD, all resources attached to this job are freed.
	TaskType_UPLOAD TaskType = 5
	// ABORT signals abortion of the current job. After ABORT, all resources attached to this job are freed.
	TaskType_ABORT TaskType = 6
)

// Enum value maps for TaskType.
var (
	TaskType_name = map[int32]string{
		0: "SORT",
		1: "SAMPLE",
		2: "SHUFFLE_SEND",
		3: "SHUFFLE_RECV",
		4: "FLUSH",
		5: "UPLOAD",
		6: "ABORT",
	}
	TaskType_value = map[string]int32{
		"SORT":         0,
		"SAMPLE":       1,
		"SHUFFLE_SEND": 2,
		"SHUFFLE_RECV": 3,
		"FLUSH":        4,
		"UPLOAD":       5,
		"ABORT":        6,
	}
)

func (x TaskType) Enum() *TaskType {
	p := new(TaskType)
	*p = x
	return p
}

func (x TaskType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskType) Descriptor() protoreflect.EnumDescriptor {
	return file_data_proto_enumTypes[0].Descriptor()
}

func (TaskType) Type() protoreflect.EnumType {
	return &file_data_proto_enumTypes[0]
}

func (x TaskType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskType.Descriptor instead.
func (TaskType) EnumDescriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{0}
}

type TaskState int32

const (
	TaskState_UNKNOWN     TaskState = 0
	TaskState_ACCEPTED    TaskState = 1
	TaskState_IN_PROGRESS TaskState = 2
	TaskState_FAILED      TaskState = 3
	TaskState_DONE        TaskState = 4
)

// Enum value maps for TaskState.
var (
	TaskState_name = map[int32]string{
		0: "UNKNOWN",
		1: "ACCEPTED",
		2: "IN_PROGRESS",
		3: "FAILED",
		4: "DONE",
	}
	TaskState_value = map[string]int32{
		"UNKNOWN":     0,
		"ACCEPTED":    1,
		"IN_PROGRESS": 2,
		"FAILED":      3,
		"DONE":        4,
	}
)

func (x TaskState) Enum() *TaskState {
	p := new(TaskState)
	*p = x
	return p
}

func (x TaskState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskState) Descriptor() protoreflect.EnumDescriptor {
	return file_data_proto_enumTypes[1].Descriptor()
}

func (TaskState) Type() protoreflect.EnumType {
	return &file_data_proto_enumTypes[1]
}

func (x TaskState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskState.Descriptor instead.
func (TaskState) EnumDescriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{1}
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message   string               `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Progress  float32              `protobuf:"fixed32,2,opt,name=progress,proto3" json:"progress,omitempty"`
	Timestamp *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{0}
}

func (x *Event) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Event) GetProgress() float32 {
	if x != nil {
		return x.Progress
	}
	return 0
}

func (x *Event) GetTimestamp() *timestamp.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

type S3Credentials struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Endpoints       []string `protobuf:"bytes,1,rep,name=endpoints,proto3" json:"endpoints,omitempty"`
	Region          string   `protobuf:"bytes,2,opt,name=region,proto3" json:"region,omitempty"`
	AccessKeyId     string   `protobuf:"bytes,3,opt,name=access_key_id,json=accessKeyId,proto3" json:"access_key_id,omitempty"`
	SecretAccessKey string   `protobuf:"bytes,4,opt,name=secret_access_key,json=secretAccessKey,proto3" json:"secret_access_key,omitempty"`
	BucketId        string   `protobuf:"bytes,5,opt,name=bucket_id,json=bucketId,proto3" json:"bucket_id,omitempty"`
	ObjectKey       string   `protobuf:"bytes,6,opt,name=object_key,json=objectKey,proto3" json:"object_key,omitempty"`
	DisableSsl      bool     `protobuf:"varint,7,opt,name=disable_ssl,json=disableSsl,proto3" json:"disable_ssl,omitempty"`
}

func (x *S3Credentials) Reset() {
	*x = S3Credentials{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *S3Credentials) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*S3Credentials) ProtoMessage() {}

func (x *S3Credentials) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use S3Credentials.ProtoReflect.Descriptor instead.
func (*S3Credentials) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{1}
}

func (x *S3Credentials) GetEndpoints() []string {
	if x != nil {
		return x.Endpoints
	}
	return nil
}

func (x *S3Credentials) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *S3Credentials) GetAccessKeyId() string {
	if x != nil {
		return x.AccessKeyId
	}
	return ""
}

func (x *S3Credentials) GetSecretAccessKey() string {
	if x != nil {
		return x.SecretAccessKey
	}
	return ""
}

func (x *S3Credentials) GetBucketId() string {
	if x != nil {
		return x.BucketId
	}
	return ""
}

func (x *S3Credentials) GetObjectKey() string {
	if x != nil {
		return x.ObjectKey
	}
	return ""
}

func (x *S3Credentials) GetDisableSsl() bool {
	if x != nil {
		return x.DisableSsl
	}
	return false
}

type Peer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address       string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	KeyRangeStart []byte `protobuf:"bytes,2,opt,name=keyRangeStart,proto3" json:"keyRangeStart,omitempty"`
	KeyRangeEnd   []byte `protobuf:"bytes,3,opt,name=keyRangeEnd,proto3" json:"keyRangeEnd,omitempty"`
}

func (x *Peer) Reset() {
	*x = Peer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Peer) ProtoMessage() {}

func (x *Peer) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Peer.ProtoReflect.Descriptor instead.
func (*Peer) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{2}
}

func (x *Peer) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Peer) GetKeyRangeStart() []byte {
	if x != nil {
		return x.KeyRangeStart
	}
	return nil
}

func (x *Peer) GetKeyRangeEnd() []byte {
	if x != nil {
		return x.KeyRangeEnd
	}
	return nil
}

type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Job  string   `protobuf:"bytes,2,opt,name=job,proto3" json:"job,omitempty"`
	Type TaskType `protobuf:"varint,3,opt,name=type,proto3,enum=cloudsort.v1.TaskType" json:"type,omitempty"`
	// Types that are assignable to Details:
	//	*Task_Sort
	//	*Task_Sample
	//	*Task_ShuffleSend
	//	*Task_ShuffleRecv
	//	*Task_Upload
	Details isTask_Details `protobuf_oneof:"details"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{3}
}

func (x *Task) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Task) GetJob() string {
	if x != nil {
		return x.Job
	}
	return ""
}

func (x *Task) GetType() TaskType {
	if x != nil {
		return x.Type
	}
	return TaskType_SORT
}

func (m *Task) GetDetails() isTask_Details {
	if m != nil {
		return m.Details
	}
	return nil
}

func (x *Task) GetSort() *SortTask {
	if x, ok := x.GetDetails().(*Task_Sort); ok {
		return x.Sort
	}
	return nil
}

func (x *Task) GetSample() *SampleTask {
	if x, ok := x.GetDetails().(*Task_Sample); ok {
		return x.Sample
	}
	return nil
}

func (x *Task) GetShuffleSend() *ShuffleSendTask {
	if x, ok := x.GetDetails().(*Task_ShuffleSend); ok {
		return x.ShuffleSend
	}
	return nil
}

func (x *Task) GetShuffleRecv() *ShuffleRecvTask {
	if x, ok := x.GetDetails().(*Task_ShuffleRecv); ok {
		return x.ShuffleRecv
	}
	return nil
}

func (x *Task) GetUpload() *UploadTask {
	if x, ok := x.GetDetails().(*Task_Upload); ok {
		return x.Upload
	}
	return nil
}

type isTask_Details interface {
	isTask_Details()
}

type Task_Sort struct {
	Sort *SortTask `protobuf:"bytes,4,opt,name=sort,proto3,oneof"`
}

type Task_Sample struct {
	Sample *SampleTask `protobuf:"bytes,5,opt,name=sample,proto3,oneof"`
}

type Task_ShuffleSend struct {
	ShuffleSend *ShuffleSendTask `protobuf:"bytes,6,opt,name=shuffle_send,json=shuffleSend,proto3,oneof"`
}

type Task_ShuffleRecv struct {
	ShuffleRecv *ShuffleRecvTask `protobuf:"bytes,7,opt,name=shuffle_recv,json=shuffleRecv,proto3,oneof"`
}

type Task_Upload struct {
	Upload *UploadTask `protobuf:"bytes,8,opt,name=upload,proto3,oneof"`
}

func (*Task_Sort) isTask_Details() {}

func (*Task_Sample) isTask_Details() {}

func (*Task_ShuffleSend) isTask_Details() {}

func (*Task_ShuffleRecv) isTask_Details() {}

func (*Task_Upload) isTask_Details() {}

type SortTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Credentials *S3Credentials `protobuf:"bytes,1,opt,name=credentials,proto3" json:"credentials,omitempty"`
	RangeStart  int64          `protobuf:"varint,2,opt,name=range_start,json=rangeStart,proto3" json:"range_start,omitempty"`
	RangeEnd    int64          `protobuf:"varint,3,opt,name=range_end,json=rangeEnd,proto3" json:"range_end,omitempty"`
}

func (x *SortTask) Reset() {
	*x = SortTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SortTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SortTask) ProtoMessage() {}

func (x *SortTask) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SortTask.ProtoReflect.Descriptor instead.
func (*SortTask) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{4}
}

func (x *SortTask) GetCredentials() *S3Credentials {
	if x != nil {
		return x.Credentials
	}
	return nil
}

func (x *SortTask) GetRangeStart() int64 {
	if x != nil {
		return x.RangeStart
	}
	return 0
}

func (x *SortTask) GetRangeEnd() int64 {
	if x != nil {
		return x.RangeEnd
	}
	return 0
}

type SampleTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NumberOfPeers int64 `protobuf:"varint,1,opt,name=number_of_peers,json=numberOfPeers,proto3" json:"number_of_peers,omitempty"`
}

func (x *SampleTask) Reset() {
	*x = SampleTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SampleTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SampleTask) ProtoMessage() {}

func (x *SampleTask) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SampleTask.ProtoReflect.Descriptor instead.
func (*SampleTask) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{5}
}

func (x *SampleTask) GetNumberOfPeers() int64 {
	if x != nil {
		return x.NumberOfPeers
	}
	return 0
}

type ShuffleSendTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Peers []*Peer `protobuf:"bytes,1,rep,name=peers,proto3" json:"peers,omitempty"`
}

func (x *ShuffleSendTask) Reset() {
	*x = ShuffleSendTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShuffleSendTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShuffleSendTask) ProtoMessage() {}

func (x *ShuffleSendTask) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShuffleSendTask.ProtoReflect.Descriptor instead.
func (*ShuffleSendTask) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{6}
}

func (x *ShuffleSendTask) GetPeers() []*Peer {
	if x != nil {
		return x.Peers
	}
	return nil
}

type ShuffleRecvTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NumberOfPeers int64 `protobuf:"varint,1,opt,name=number_of_peers,json=numberOfPeers,proto3" json:"number_of_peers,omitempty"`
}

func (x *ShuffleRecvTask) Reset() {
	*x = ShuffleRecvTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShuffleRecvTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShuffleRecvTask) ProtoMessage() {}

func (x *ShuffleRecvTask) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShuffleRecvTask.ProtoReflect.Descriptor instead.
func (*ShuffleRecvTask) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{7}
}

func (x *ShuffleRecvTask) GetNumberOfPeers() int64 {
	if x != nil {
		return x.NumberOfPeers
	}
	return 0
}

type UploadTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Credentials *S3Credentials `protobuf:"bytes,1,opt,name=credentials,proto3" json:"credentials,omitempty"`
}

func (x *UploadTask) Reset() {
	*x = UploadTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadTask) ProtoMessage() {}

func (x *UploadTask) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadTask.ProtoReflect.Descriptor instead.
func (*UploadTask) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{8}
}

func (x *UploadTask) GetCredentials() *S3Credentials {
	if x != nil {
		return x.Credentials
	}
	return nil
}

type TaskData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Properties:
	//	*TaskData_ShuffleRecvAddr
	Properties isTaskData_Properties `protobuf_oneof:"properties"`
}

func (x *TaskData) Reset() {
	*x = TaskData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskData) ProtoMessage() {}

func (x *TaskData) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskData.ProtoReflect.Descriptor instead.
func (*TaskData) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{9}
}

func (m *TaskData) GetProperties() isTaskData_Properties {
	if m != nil {
		return m.Properties
	}
	return nil
}

func (x *TaskData) GetShuffleRecvAddr() string {
	if x, ok := x.GetProperties().(*TaskData_ShuffleRecvAddr); ok {
		return x.ShuffleRecvAddr
	}
	return ""
}

type isTaskData_Properties interface {
	isTaskData_Properties()
}

type TaskData_ShuffleRecvAddr struct {
	ShuffleRecvAddr string `protobuf:"bytes,3,opt,name=shuffleRecvAddr,proto3,oneof"`
}

func (*TaskData_ShuffleRecvAddr) isTaskData_Properties() {}

var File_data_proto protoreflect.FileDescriptor

var file_data_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x77, 0x0a, 0x05, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1a,
	0x0a, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x22, 0xf2, 0x01, 0x0a, 0x0d, 0x53, 0x33, 0x43, 0x72, 0x65, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x65, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0d,
	0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x49, 0x64,
	0x12, 0x2a, 0x0a, 0x11, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x12, 0x1b, 0x0a, 0x09,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x69, 0x73, 0x61,
	0x62, 0x6c, 0x65, 0x5f, 0x73, 0x73, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x64,
	0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x73, 0x6c, 0x22, 0x68, 0x0a, 0x04, 0x50, 0x65, 0x65,
	0x72, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x6b,
	0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x53, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0d, 0x6b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x12, 0x20, 0x0a, 0x0b, 0x6b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x45, 0x6e, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x6b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65,
	0x45, 0x6e, 0x64, 0x22, 0x81, 0x03, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x6a, 0x6f, 0x62, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6a,
	0x6f, 0x62, 0x12, 0x2a, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x16, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e,
	0x54, 0x61, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x2c,
	0x0a, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x6f, 0x72, 0x74,
	0x54, 0x61, 0x73, 0x6b, 0x48, 0x00, 0x52, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x12, 0x32, 0x0a, 0x06,
	0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x48, 0x00, 0x52, 0x06, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x12, 0x42, 0x0a, 0x0c, 0x73, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65, 0x5f, 0x73, 0x65, 0x6e, 0x64,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x6f,
	0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65, 0x53, 0x65, 0x6e,
	0x64, 0x54, 0x61, 0x73, 0x6b, 0x48, 0x00, 0x52, 0x0b, 0x73, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65,
	0x53, 0x65, 0x6e, 0x64, 0x12, 0x42, 0x0a, 0x0c, 0x73, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65, 0x5f,
	0x72, 0x65, 0x63, 0x76, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x68, 0x75, 0x66, 0x66, 0x6c,
	0x65, 0x52, 0x65, 0x63, 0x76, 0x54, 0x61, 0x73, 0x6b, 0x48, 0x00, 0x52, 0x0b, 0x73, 0x68, 0x75,
	0x66, 0x66, 0x6c, 0x65, 0x52, 0x65, 0x63, 0x76, 0x12, 0x32, 0x0a, 0x06, 0x75, 0x70, 0x6c, 0x6f,
	0x61, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x61,
	0x73, 0x6b, 0x48, 0x00, 0x52, 0x06, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x09, 0x0a, 0x07,
	0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x22, 0x87, 0x01, 0x0a, 0x08, 0x53, 0x6f, 0x72, 0x74,
	0x54, 0x61, 0x73, 0x6b, 0x12, 0x3d, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x33, 0x43, 0x72, 0x65, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x61, 0x6c, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x65, 0x6e,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x45, 0x6e,
	0x64, 0x22, 0x34, 0x0a, 0x0a, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x12,
	0x26, 0x0a, 0x0f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x6f, 0x66, 0x5f, 0x70, 0x65, 0x65,
	0x72, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x4f, 0x66, 0x50, 0x65, 0x65, 0x72, 0x73, 0x22, 0x3b, 0x0a, 0x0f, 0x53, 0x68, 0x75, 0x66, 0x66,
	0x6c, 0x65, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x28, 0x0a, 0x05, 0x70, 0x65,
	0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x05, 0x70,
	0x65, 0x65, 0x72, 0x73, 0x22, 0x39, 0x0a, 0x0f, 0x53, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65, 0x52,
	0x65, 0x63, 0x76, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x5f, 0x6f, 0x66, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x50, 0x65, 0x65, 0x72, 0x73, 0x22,
	0x4b, 0x0a, 0x0a, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x3d, 0x0a,
	0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x33, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x52,
	0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x22, 0x44, 0x0a, 0x08,
	0x54, 0x61, 0x73, 0x6b, 0x44, 0x61, 0x74, 0x61, 0x12, 0x2a, 0x0a, 0x0f, 0x73, 0x68, 0x75, 0x66,
	0x66, 0x6c, 0x65, 0x52, 0x65, 0x63, 0x76, 0x41, 0x64, 0x64, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x0f, 0x73, 0x68, 0x75, 0x66, 0x66, 0x6c, 0x65, 0x52, 0x65, 0x63, 0x76,
	0x41, 0x64, 0x64, 0x72, 0x42, 0x0c, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69,
	0x65, 0x73, 0x2a, 0x66, 0x0a, 0x08, 0x54, 0x61, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08,
	0x0a, 0x04, 0x53, 0x4f, 0x52, 0x54, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x41, 0x4d, 0x50,
	0x4c, 0x45, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x48, 0x55, 0x46, 0x46, 0x4c, 0x45, 0x5f,
	0x53, 0x45, 0x4e, 0x44, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x48, 0x55, 0x46, 0x46, 0x4c,
	0x45, 0x5f, 0x52, 0x45, 0x43, 0x56, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x46, 0x4c, 0x55, 0x53,
	0x48, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x10, 0x05, 0x12,
	0x09, 0x0a, 0x05, 0x41, 0x42, 0x4f, 0x52, 0x54, 0x10, 0x06, 0x2a, 0x4d, 0x0a, 0x09, 0x54, 0x61,
	0x73, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f,
	0x57, 0x4e, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54, 0x45, 0x44,
	0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x49, 0x4e, 0x5f, 0x50, 0x52, 0x4f, 0x47, 0x52, 0x45, 0x53,
	0x53, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x03, 0x12,
	0x08, 0x0a, 0x04, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x04, 0x42, 0x21, 0x5a, 0x1f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x6e, 0x73, 0x70, 0x2f, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x73, 0x6f, 0x72, 0x74, 0x2f, 0x70, 0x62, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_data_proto_rawDescOnce sync.Once
	file_data_proto_rawDescData = file_data_proto_rawDesc
)

func file_data_proto_rawDescGZIP() []byte {
	file_data_proto_rawDescOnce.Do(func() {
		file_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_data_proto_rawDescData)
	})
	return file_data_proto_rawDescData
}

var file_data_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_data_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_data_proto_goTypes = []interface{}{
	(TaskType)(0),               // 0: cloudsort.v1.TaskType
	(TaskState)(0),              // 1: cloudsort.v1.TaskState
	(*Event)(nil),               // 2: cloudsort.v1.Event
	(*S3Credentials)(nil),       // 3: cloudsort.v1.S3Credentials
	(*Peer)(nil),                // 4: cloudsort.v1.Peer
	(*Task)(nil),                // 5: cloudsort.v1.Task
	(*SortTask)(nil),            // 6: cloudsort.v1.SortTask
	(*SampleTask)(nil),          // 7: cloudsort.v1.SampleTask
	(*ShuffleSendTask)(nil),     // 8: cloudsort.v1.ShuffleSendTask
	(*ShuffleRecvTask)(nil),     // 9: cloudsort.v1.ShuffleRecvTask
	(*UploadTask)(nil),          // 10: cloudsort.v1.UploadTask
	(*TaskData)(nil),            // 11: cloudsort.v1.TaskData
	(*timestamp.Timestamp)(nil), // 12: google.protobuf.Timestamp
}
var file_data_proto_depIdxs = []int32{
	12, // 0: cloudsort.v1.Event.timestamp:type_name -> google.protobuf.Timestamp
	0,  // 1: cloudsort.v1.Task.type:type_name -> cloudsort.v1.TaskType
	6,  // 2: cloudsort.v1.Task.sort:type_name -> cloudsort.v1.SortTask
	7,  // 3: cloudsort.v1.Task.sample:type_name -> cloudsort.v1.SampleTask
	8,  // 4: cloudsort.v1.Task.shuffle_send:type_name -> cloudsort.v1.ShuffleSendTask
	9,  // 5: cloudsort.v1.Task.shuffle_recv:type_name -> cloudsort.v1.ShuffleRecvTask
	10, // 6: cloudsort.v1.Task.upload:type_name -> cloudsort.v1.UploadTask
	3,  // 7: cloudsort.v1.SortTask.credentials:type_name -> cloudsort.v1.S3Credentials
	4,  // 8: cloudsort.v1.ShuffleSendTask.peers:type_name -> cloudsort.v1.Peer
	3,  // 9: cloudsort.v1.UploadTask.credentials:type_name -> cloudsort.v1.S3Credentials
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_data_proto_init() }
func file_data_proto_init() {
	if File_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*S3Credentials); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Peer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SortTask); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SampleTask); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShuffleSendTask); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShuffleRecvTask); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadTask); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_data_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*Task_Sort)(nil),
		(*Task_Sample)(nil),
		(*Task_ShuffleSend)(nil),
		(*Task_ShuffleRecv)(nil),
		(*Task_Upload)(nil),
	}
	file_data_proto_msgTypes[9].OneofWrappers = []interface{}{
		(*TaskData_ShuffleRecvAddr)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_data_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_data_proto_goTypes,
		DependencyIndexes: file_data_proto_depIdxs,
		EnumInfos:         file_data_proto_enumTypes,
		MessageInfos:      file_data_proto_msgTypes,
	}.Build()
	File_data_proto = out.File
	file_data_proto_rawDesc = nil
	file_data_proto_goTypes = nil
	file_data_proto_depIdxs = nil
}
