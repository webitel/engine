// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: queue_bucket.proto

package engine

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type DeleteQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId int64 `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
}

func (x *DeleteQueueBucketRequest) Reset() {
	*x = DeleteQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteQueueBucketRequest) ProtoMessage() {}

func (x *DeleteQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*DeleteQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{0}
}

func (x *DeleteQueueBucketRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *DeleteQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

type UpdateQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId  int64   `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Ratio    int32   `protobuf:"varint,3,opt,name=ratio,proto3" json:"ratio,omitempty"`
	Bucket   *Lookup `protobuf:"bytes,4,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Disabled bool    `protobuf:"varint,5,opt,name=disabled,proto3" json:"disabled,omitempty"`
	Priority int32   `protobuf:"varint,6,opt,name=priority,proto3" json:"priority,omitempty"`
}

func (x *UpdateQueueBucketRequest) Reset() {
	*x = UpdateQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateQueueBucketRequest) ProtoMessage() {}

func (x *UpdateQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*UpdateQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateQueueBucketRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

func (x *UpdateQueueBucketRequest) GetRatio() int32 {
	if x != nil {
		return x.Ratio
	}
	return 0
}

func (x *UpdateQueueBucketRequest) GetBucket() *Lookup {
	if x != nil {
		return x.Bucket
	}
	return nil
}

func (x *UpdateQueueBucketRequest) GetDisabled() bool {
	if x != nil {
		return x.Disabled
	}
	return false
}

func (x *UpdateQueueBucketRequest) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

type PatchQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId  int64    `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Ratio    int32    `protobuf:"varint,3,opt,name=ratio,proto3" json:"ratio,omitempty"`
	Bucket   *Lookup  `protobuf:"bytes,4,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Disabled bool     `protobuf:"varint,5,opt,name=disabled,proto3" json:"disabled,omitempty"`
	Priority int32    `protobuf:"varint,6,opt,name=priority,proto3" json:"priority,omitempty"`
	Fields   []string `protobuf:"bytes,7,rep,name=fields,proto3" json:"fields,omitempty"`
}

func (x *PatchQueueBucketRequest) Reset() {
	*x = PatchQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PatchQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PatchQueueBucketRequest) ProtoMessage() {}

func (x *PatchQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PatchQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*PatchQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{2}
}

func (x *PatchQueueBucketRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *PatchQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

func (x *PatchQueueBucketRequest) GetRatio() int32 {
	if x != nil {
		return x.Ratio
	}
	return 0
}

func (x *PatchQueueBucketRequest) GetBucket() *Lookup {
	if x != nil {
		return x.Bucket
	}
	return nil
}

func (x *PatchQueueBucketRequest) GetDisabled() bool {
	if x != nil {
		return x.Disabled
	}
	return false
}

func (x *PatchQueueBucketRequest) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

func (x *PatchQueueBucketRequest) GetFields() []string {
	if x != nil {
		return x.Fields
	}
	return nil
}

type SearchQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueueId int64    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Page    int32    `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty"`
	Size    int32    `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Q       string   `protobuf:"bytes,4,opt,name=q,proto3" json:"q,omitempty"`
	Sort    string   `protobuf:"bytes,5,opt,name=sort,proto3" json:"sort,omitempty"`
	Fields  []string `protobuf:"bytes,6,rep,name=fields,proto3" json:"fields,omitempty"`
	Id      []uint32 `protobuf:"varint,7,rep,packed,name=id,proto3" json:"id,omitempty"`
}

func (x *SearchQueueBucketRequest) Reset() {
	*x = SearchQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchQueueBucketRequest) ProtoMessage() {}

func (x *SearchQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*SearchQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{3}
}

func (x *SearchQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

func (x *SearchQueueBucketRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *SearchQueueBucketRequest) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *SearchQueueBucketRequest) GetQ() string {
	if x != nil {
		return x.Q
	}
	return ""
}

func (x *SearchQueueBucketRequest) GetSort() string {
	if x != nil {
		return x.Sort
	}
	return ""
}

func (x *SearchQueueBucketRequest) GetFields() []string {
	if x != nil {
		return x.Fields
	}
	return nil
}

func (x *SearchQueueBucketRequest) GetId() []uint32 {
	if x != nil {
		return x.Id
	}
	return nil
}

type ListQueueBucket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Next  bool           `protobuf:"varint,1,opt,name=next,proto3" json:"next,omitempty"`
	Items []*QueueBucket `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListQueueBucket) Reset() {
	*x = ListQueueBucket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListQueueBucket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListQueueBucket) ProtoMessage() {}

func (x *ListQueueBucket) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListQueueBucket.ProtoReflect.Descriptor instead.
func (*ListQueueBucket) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{4}
}

func (x *ListQueueBucket) GetNext() bool {
	if x != nil {
		return x.Next
	}
	return false
}

func (x *ListQueueBucket) GetItems() []*QueueBucket {
	if x != nil {
		return x.Items
	}
	return nil
}

type ReadQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId int64 `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
}

func (x *ReadQueueBucketRequest) Reset() {
	*x = ReadQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadQueueBucketRequest) ProtoMessage() {}

func (x *ReadQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*ReadQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{5}
}

func (x *ReadQueueBucketRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ReadQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

type CreateQueueBucketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueueId  int64   `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Ratio    int32   `protobuf:"varint,2,opt,name=ratio,proto3" json:"ratio,omitempty"`
	Bucket   *Lookup `protobuf:"bytes,3,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Disabled bool    `protobuf:"varint,4,opt,name=disabled,proto3" json:"disabled,omitempty"`
	Priority int32   `protobuf:"varint,5,opt,name=priority,proto3" json:"priority,omitempty"`
}

func (x *CreateQueueBucketRequest) Reset() {
	*x = CreateQueueBucketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateQueueBucketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateQueueBucketRequest) ProtoMessage() {}

func (x *CreateQueueBucketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateQueueBucketRequest.ProtoReflect.Descriptor instead.
func (*CreateQueueBucketRequest) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{6}
}

func (x *CreateQueueBucketRequest) GetQueueId() int64 {
	if x != nil {
		return x.QueueId
	}
	return 0
}

func (x *CreateQueueBucketRequest) GetRatio() int32 {
	if x != nil {
		return x.Ratio
	}
	return 0
}

func (x *CreateQueueBucketRequest) GetBucket() *Lookup {
	if x != nil {
		return x.Bucket
	}
	return nil
}

func (x *CreateQueueBucketRequest) GetDisabled() bool {
	if x != nil {
		return x.Disabled
	}
	return false
}

func (x *CreateQueueBucketRequest) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

type QueueBucket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Ratio    int32   `protobuf:"varint,2,opt,name=ratio,proto3" json:"ratio,omitempty"`
	Bucket   *Lookup `protobuf:"bytes,3,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Disabled bool    `protobuf:"varint,4,opt,name=disabled,proto3" json:"disabled,omitempty"`
	Priority int32   `protobuf:"varint,5,opt,name=priority,proto3" json:"priority,omitempty"`
}

func (x *QueueBucket) Reset() {
	*x = QueueBucket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_bucket_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueueBucket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueueBucket) ProtoMessage() {}

func (x *QueueBucket) ProtoReflect() protoreflect.Message {
	mi := &file_queue_bucket_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueueBucket.ProtoReflect.Descriptor instead.
func (*QueueBucket) Descriptor() ([]byte, []int) {
	return file_queue_bucket_proto_rawDescGZIP(), []int{7}
}

func (x *QueueBucket) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *QueueBucket) GetRatio() int32 {
	if x != nil {
		return x.Ratio
	}
	return 0
}

func (x *QueueBucket) GetBucket() *Lookup {
	if x != nil {
		return x.Bucket
	}
	return nil
}

func (x *QueueBucket) GetDisabled() bool {
	if x != nil {
		return x.Disabled
	}
	return false
}

func (x *QueueBucket) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

var File_queue_bucket_proto protoreflect.FileDescriptor

var file_queue_bucket_proto_rawDesc = []byte{
	0x0a, 0x12, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x1a, 0x0b, 0x63, 0x6f,
	0x6e, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x45, 0x0a, 0x18, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x71, 0x75, 0x65, 0x75, 0x65, 0x49, 0x64, 0x22, 0xbb,
	0x01, 0x0a, 0x18, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75,
	0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x12, 0x26, 0x0a, 0x06,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65,
	0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x06, 0x62, 0x75,
	0x63, 0x6b, 0x65, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x22, 0xd2, 0x01, 0x0a,
	0x17, 0x50, 0x61, 0x74, 0x63, 0x68, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x12, 0x26, 0x0a, 0x06, 0x62, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x22, 0xa7, 0x01, 0x0a, 0x18, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x51, 0x75, 0x65, 0x75,
	0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19,
	0x0a, 0x08, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x71, 0x75, 0x65, 0x75, 0x65, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x69, 0x7a,
	0x65, 0x12, 0x0c, 0x0a, 0x01, 0x71, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x71, 0x12,
	0x12, 0x0a, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73,
	0x6f, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x06, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x22, 0x50, 0x0a, 0x0f, 0x4c,
	0x69, 0x73, 0x74, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x6e, 0x65,
	0x78, 0x74, 0x12, 0x29, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22, 0x43, 0x0a,
	0x16, 0x52, 0x65, 0x61, 0x64, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x49, 0x64, 0x22, 0xab, 0x01, 0x0a, 0x18, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x51, 0x75, 0x65,
	0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x19, 0x0a, 0x08, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x07, 0x71, 0x75, 0x65, 0x75, 0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x12, 0x26, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0e, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70,
	0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x64, 0x69, 0x73, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79,
	0x22, 0x93, 0x01, 0x0a, 0x0b, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x05, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x12, 0x26, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e,
	0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1a,
	0x0a, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x08, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72,
	0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x72,
	0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x32, 0x99, 0x06, 0x0a, 0x12, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7d, 0x0a,
	0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b,
	0x65, 0x74, 0x12, 0x20, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x51, 0x75,
	0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22, 0x31, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x2b, 0x3a, 0x01, 0x2a, 0x22, 0x26, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e, 0x74,
	0x65, 0x72, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x12, 0x7e, 0x0a, 0x11,
	0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x12, 0x20, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22, 0x2e, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x28, 0x12, 0x26, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e, 0x74,
	0x65, 0x72, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x12, 0x7b, 0x0a, 0x0f,
	0x52, 0x65, 0x61, 0x64, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12,
	0x1e, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x51, 0x75, 0x65,
	0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75,
	0x63, 0x6b, 0x65, 0x74, 0x22, 0x33, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x2d, 0x12, 0x2b, 0x2f, 0x63,
	0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x62, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x82, 0x01, 0x0a, 0x11, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12,
	0x20, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x51,
	0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22, 0x36, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x30, 0x3a, 0x01,
	0x2a, 0x1a, 0x2b, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x2f,
	0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x69, 0x64,
	0x7d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x80,
	0x01, 0x0a, 0x10, 0x50, 0x61, 0x74, 0x63, 0x68, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x12, 0x1f, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x50, 0x61, 0x74,
	0x63, 0x68, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x51, 0x75,
	0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22, 0x36, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x30, 0x3a, 0x01, 0x2a, 0x32, 0x2b, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e, 0x74,
	0x65, 0x72, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64,
	0x7d, 0x12, 0x7f, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x20, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2e, 0x51, 0x75, 0x65, 0x75, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22, 0x33, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x2d, 0x2a, 0x2b, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x63, 0x65, 0x6e,
	0x74, 0x65, 0x72, 0x2f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x73, 0x2f, 0x7b, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x2f, 0x7b, 0x69,
	0x64, 0x7d, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f,
	0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_queue_bucket_proto_rawDescOnce sync.Once
	file_queue_bucket_proto_rawDescData = file_queue_bucket_proto_rawDesc
)

func file_queue_bucket_proto_rawDescGZIP() []byte {
	file_queue_bucket_proto_rawDescOnce.Do(func() {
		file_queue_bucket_proto_rawDescData = protoimpl.X.CompressGZIP(file_queue_bucket_proto_rawDescData)
	})
	return file_queue_bucket_proto_rawDescData
}

var file_queue_bucket_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_queue_bucket_proto_goTypes = []interface{}{
	(*DeleteQueueBucketRequest)(nil), // 0: engine.DeleteQueueBucketRequest
	(*UpdateQueueBucketRequest)(nil), // 1: engine.UpdateQueueBucketRequest
	(*PatchQueueBucketRequest)(nil),  // 2: engine.PatchQueueBucketRequest
	(*SearchQueueBucketRequest)(nil), // 3: engine.SearchQueueBucketRequest
	(*ListQueueBucket)(nil),          // 4: engine.ListQueueBucket
	(*ReadQueueBucketRequest)(nil),   // 5: engine.ReadQueueBucketRequest
	(*CreateQueueBucketRequest)(nil), // 6: engine.CreateQueueBucketRequest
	(*QueueBucket)(nil),              // 7: engine.QueueBucket
	(*Lookup)(nil),                   // 8: engine.Lookup
}
var file_queue_bucket_proto_depIdxs = []int32{
	8,  // 0: engine.UpdateQueueBucketRequest.bucket:type_name -> engine.Lookup
	8,  // 1: engine.PatchQueueBucketRequest.bucket:type_name -> engine.Lookup
	7,  // 2: engine.ListQueueBucket.items:type_name -> engine.QueueBucket
	8,  // 3: engine.CreateQueueBucketRequest.bucket:type_name -> engine.Lookup
	8,  // 4: engine.QueueBucket.bucket:type_name -> engine.Lookup
	6,  // 5: engine.QueueBucketService.CreateQueueBucket:input_type -> engine.CreateQueueBucketRequest
	3,  // 6: engine.QueueBucketService.SearchQueueBucket:input_type -> engine.SearchQueueBucketRequest
	5,  // 7: engine.QueueBucketService.ReadQueueBucket:input_type -> engine.ReadQueueBucketRequest
	1,  // 8: engine.QueueBucketService.UpdateQueueBucket:input_type -> engine.UpdateQueueBucketRequest
	2,  // 9: engine.QueueBucketService.PatchQueueBucket:input_type -> engine.PatchQueueBucketRequest
	0,  // 10: engine.QueueBucketService.DeleteQueueBucket:input_type -> engine.DeleteQueueBucketRequest
	7,  // 11: engine.QueueBucketService.CreateQueueBucket:output_type -> engine.QueueBucket
	4,  // 12: engine.QueueBucketService.SearchQueueBucket:output_type -> engine.ListQueueBucket
	7,  // 13: engine.QueueBucketService.ReadQueueBucket:output_type -> engine.QueueBucket
	7,  // 14: engine.QueueBucketService.UpdateQueueBucket:output_type -> engine.QueueBucket
	7,  // 15: engine.QueueBucketService.PatchQueueBucket:output_type -> engine.QueueBucket
	7,  // 16: engine.QueueBucketService.DeleteQueueBucket:output_type -> engine.QueueBucket
	11, // [11:17] is the sub-list for method output_type
	5,  // [5:11] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_queue_bucket_proto_init() }
func file_queue_bucket_proto_init() {
	if File_queue_bucket_proto != nil {
		return
	}
	file_const_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_queue_bucket_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PatchQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListQueueBucket); i {
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
		file_queue_bucket_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateQueueBucketRequest); i {
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
		file_queue_bucket_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueueBucket); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_queue_bucket_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_queue_bucket_proto_goTypes,
		DependencyIndexes: file_queue_bucket_proto_depIdxs,
		MessageInfos:      file_queue_bucket_proto_msgTypes,
	}.Build()
	File_queue_bucket_proto = out.File
	file_queue_bucket_proto_rawDesc = nil
	file_queue_bucket_proto_goTypes = nil
	file_queue_bucket_proto_depIdxs = nil
}
