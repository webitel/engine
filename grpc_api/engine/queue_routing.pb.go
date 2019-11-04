// Code generated by protoc-gen-go. DO NOT EDIT.
// source: queue_routing.proto

package engine

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type DeleteQueueRoutingRequest struct {
	QueueId              int64    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	DomainId             int64    `protobuf:"varint,2,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	Id                   int64    `protobuf:"varint,3,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteQueueRoutingRequest) Reset()         { *m = DeleteQueueRoutingRequest{} }
func (m *DeleteQueueRoutingRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteQueueRoutingRequest) ProtoMessage()    {}
func (*DeleteQueueRoutingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{0}
}

func (m *DeleteQueueRoutingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteQueueRoutingRequest.Unmarshal(m, b)
}
func (m *DeleteQueueRoutingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteQueueRoutingRequest.Marshal(b, m, deterministic)
}
func (m *DeleteQueueRoutingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteQueueRoutingRequest.Merge(m, src)
}
func (m *DeleteQueueRoutingRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteQueueRoutingRequest.Size(m)
}
func (m *DeleteQueueRoutingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteQueueRoutingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteQueueRoutingRequest proto.InternalMessageInfo

func (m *DeleteQueueRoutingRequest) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *DeleteQueueRoutingRequest) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

func (m *DeleteQueueRoutingRequest) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type UpdateQueueRoutingRequest struct {
	Id                   int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId              int64    `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Pattern              string   `protobuf:"bytes,3,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Priority             int32    `protobuf:"varint,4,opt,name=priority,proto3" json:"priority,omitempty"`
	Disabled             bool     `protobuf:"varint,5,opt,name=disabled,proto3" json:"disabled,omitempty"`
	DomainId             int64    `protobuf:"varint,6,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateQueueRoutingRequest) Reset()         { *m = UpdateQueueRoutingRequest{} }
func (m *UpdateQueueRoutingRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateQueueRoutingRequest) ProtoMessage()    {}
func (*UpdateQueueRoutingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{1}
}

func (m *UpdateQueueRoutingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateQueueRoutingRequest.Unmarshal(m, b)
}
func (m *UpdateQueueRoutingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateQueueRoutingRequest.Marshal(b, m, deterministic)
}
func (m *UpdateQueueRoutingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateQueueRoutingRequest.Merge(m, src)
}
func (m *UpdateQueueRoutingRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateQueueRoutingRequest.Size(m)
}
func (m *UpdateQueueRoutingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateQueueRoutingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateQueueRoutingRequest proto.InternalMessageInfo

func (m *UpdateQueueRoutingRequest) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *UpdateQueueRoutingRequest) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *UpdateQueueRoutingRequest) GetPattern() string {
	if m != nil {
		return m.Pattern
	}
	return ""
}

func (m *UpdateQueueRoutingRequest) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *UpdateQueueRoutingRequest) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *UpdateQueueRoutingRequest) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

type ReadQueueRoutingRequest struct {
	QueueId              int64    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	DomainId             int64    `protobuf:"varint,2,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	Id                   int64    `protobuf:"varint,3,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadQueueRoutingRequest) Reset()         { *m = ReadQueueRoutingRequest{} }
func (m *ReadQueueRoutingRequest) String() string { return proto.CompactTextString(m) }
func (*ReadQueueRoutingRequest) ProtoMessage()    {}
func (*ReadQueueRoutingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{2}
}

func (m *ReadQueueRoutingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadQueueRoutingRequest.Unmarshal(m, b)
}
func (m *ReadQueueRoutingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadQueueRoutingRequest.Marshal(b, m, deterministic)
}
func (m *ReadQueueRoutingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadQueueRoutingRequest.Merge(m, src)
}
func (m *ReadQueueRoutingRequest) XXX_Size() int {
	return xxx_messageInfo_ReadQueueRoutingRequest.Size(m)
}
func (m *ReadQueueRoutingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadQueueRoutingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReadQueueRoutingRequest proto.InternalMessageInfo

func (m *ReadQueueRoutingRequest) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *ReadQueueRoutingRequest) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

func (m *ReadQueueRoutingRequest) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type SearchQueueRoutingRequest struct {
	QueueId              int64    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	DomainId             int64    `protobuf:"varint,2,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	Size                 int32    `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Page                 int32    `protobuf:"varint,4,opt,name=page,proto3" json:"page,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SearchQueueRoutingRequest) Reset()         { *m = SearchQueueRoutingRequest{} }
func (m *SearchQueueRoutingRequest) String() string { return proto.CompactTextString(m) }
func (*SearchQueueRoutingRequest) ProtoMessage()    {}
func (*SearchQueueRoutingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{3}
}

func (m *SearchQueueRoutingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchQueueRoutingRequest.Unmarshal(m, b)
}
func (m *SearchQueueRoutingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchQueueRoutingRequest.Marshal(b, m, deterministic)
}
func (m *SearchQueueRoutingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchQueueRoutingRequest.Merge(m, src)
}
func (m *SearchQueueRoutingRequest) XXX_Size() int {
	return xxx_messageInfo_SearchQueueRoutingRequest.Size(m)
}
func (m *SearchQueueRoutingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchQueueRoutingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SearchQueueRoutingRequest proto.InternalMessageInfo

func (m *SearchQueueRoutingRequest) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *SearchQueueRoutingRequest) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

func (m *SearchQueueRoutingRequest) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *SearchQueueRoutingRequest) GetPage() int32 {
	if m != nil {
		return m.Page
	}
	return 0
}

type ListQueueRouting struct {
	Items                []*QueueRouting `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ListQueueRouting) Reset()         { *m = ListQueueRouting{} }
func (m *ListQueueRouting) String() string { return proto.CompactTextString(m) }
func (*ListQueueRouting) ProtoMessage()    {}
func (*ListQueueRouting) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{4}
}

func (m *ListQueueRouting) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListQueueRouting.Unmarshal(m, b)
}
func (m *ListQueueRouting) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListQueueRouting.Marshal(b, m, deterministic)
}
func (m *ListQueueRouting) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListQueueRouting.Merge(m, src)
}
func (m *ListQueueRouting) XXX_Size() int {
	return xxx_messageInfo_ListQueueRouting.Size(m)
}
func (m *ListQueueRouting) XXX_DiscardUnknown() {
	xxx_messageInfo_ListQueueRouting.DiscardUnknown(m)
}

var xxx_messageInfo_ListQueueRouting proto.InternalMessageInfo

func (m *ListQueueRouting) GetItems() []*QueueRouting {
	if m != nil {
		return m.Items
	}
	return nil
}

type CreateQueueRoutingRequest struct {
	QueueId              int64    `protobuf:"varint,1,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Pattern              string   `protobuf:"bytes,2,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Priority             int32    `protobuf:"varint,3,opt,name=priority,proto3" json:"priority,omitempty"`
	Disabled             bool     `protobuf:"varint,4,opt,name=disabled,proto3" json:"disabled,omitempty"`
	DomainId             int64    `protobuf:"varint,5,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateQueueRoutingRequest) Reset()         { *m = CreateQueueRoutingRequest{} }
func (m *CreateQueueRoutingRequest) String() string { return proto.CompactTextString(m) }
func (*CreateQueueRoutingRequest) ProtoMessage()    {}
func (*CreateQueueRoutingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{5}
}

func (m *CreateQueueRoutingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateQueueRoutingRequest.Unmarshal(m, b)
}
func (m *CreateQueueRoutingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateQueueRoutingRequest.Marshal(b, m, deterministic)
}
func (m *CreateQueueRoutingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateQueueRoutingRequest.Merge(m, src)
}
func (m *CreateQueueRoutingRequest) XXX_Size() int {
	return xxx_messageInfo_CreateQueueRoutingRequest.Size(m)
}
func (m *CreateQueueRoutingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateQueueRoutingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateQueueRoutingRequest proto.InternalMessageInfo

func (m *CreateQueueRoutingRequest) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *CreateQueueRoutingRequest) GetPattern() string {
	if m != nil {
		return m.Pattern
	}
	return ""
}

func (m *CreateQueueRoutingRequest) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *CreateQueueRoutingRequest) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *CreateQueueRoutingRequest) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

type QueueRouting struct {
	Id                   int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	QueueId              int64    `protobuf:"varint,2,opt,name=queue_id,json=queueId,proto3" json:"queue_id,omitempty"`
	Pattern              string   `protobuf:"bytes,3,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Priority             int32    `protobuf:"varint,4,opt,name=priority,proto3" json:"priority,omitempty"`
	Disabled             bool     `protobuf:"varint,5,opt,name=disabled,proto3" json:"disabled,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueueRouting) Reset()         { *m = QueueRouting{} }
func (m *QueueRouting) String() string { return proto.CompactTextString(m) }
func (*QueueRouting) ProtoMessage()    {}
func (*QueueRouting) Descriptor() ([]byte, []int) {
	return fileDescriptor_086384a2ff08799d, []int{6}
}

func (m *QueueRouting) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueueRouting.Unmarshal(m, b)
}
func (m *QueueRouting) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueueRouting.Marshal(b, m, deterministic)
}
func (m *QueueRouting) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueueRouting.Merge(m, src)
}
func (m *QueueRouting) XXX_Size() int {
	return xxx_messageInfo_QueueRouting.Size(m)
}
func (m *QueueRouting) XXX_DiscardUnknown() {
	xxx_messageInfo_QueueRouting.DiscardUnknown(m)
}

var xxx_messageInfo_QueueRouting proto.InternalMessageInfo

func (m *QueueRouting) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *QueueRouting) GetQueueId() int64 {
	if m != nil {
		return m.QueueId
	}
	return 0
}

func (m *QueueRouting) GetPattern() string {
	if m != nil {
		return m.Pattern
	}
	return ""
}

func (m *QueueRouting) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *QueueRouting) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func init() {
	proto.RegisterType((*DeleteQueueRoutingRequest)(nil), "engine.DeleteQueueRoutingRequest")
	proto.RegisterType((*UpdateQueueRoutingRequest)(nil), "engine.UpdateQueueRoutingRequest")
	proto.RegisterType((*ReadQueueRoutingRequest)(nil), "engine.ReadQueueRoutingRequest")
	proto.RegisterType((*SearchQueueRoutingRequest)(nil), "engine.SearchQueueRoutingRequest")
	proto.RegisterType((*ListQueueRouting)(nil), "engine.ListQueueRouting")
	proto.RegisterType((*CreateQueueRoutingRequest)(nil), "engine.CreateQueueRoutingRequest")
	proto.RegisterType((*QueueRouting)(nil), "engine.QueueRouting")
}

func init() { proto.RegisterFile("queue_routing.proto", fileDescriptor_086384a2ff08799d) }

var fileDescriptor_086384a2ff08799d = []byte{
	// 514 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x55, 0x41, 0x6b, 0xd4, 0x40,
	0x14, 0x66, 0xb2, 0x9b, 0xed, 0xf6, 0x29, 0x52, 0x5e, 0x05, 0x93, 0x28, 0xb8, 0xe6, 0xe2, 0xba,
	0x94, 0x0d, 0xae, 0x82, 0xe0, 0xc1, 0x8b, 0x5e, 0x0a, 0x5e, 0x4c, 0xf1, 0x5c, 0xa6, 0x99, 0x47,
	0x1c, 0x48, 0x93, 0x74, 0x32, 0x2b, 0xe8, 0x52, 0xc4, 0x0a, 0x5e, 0x3c, 0xfa, 0x1b, 0xbc, 0x7a,
	0xf3, 0x97, 0xf8, 0x17, 0xfc, 0x21, 0x92, 0x49, 0xd3, 0xdd, 0xec, 0x36, 0x4b, 0x17, 0x94, 0xde,
	0xf2, 0x66, 0x1e, 0xef, 0xfb, 0xde, 0xf7, 0xbd, 0x79, 0x81, 0xdd, 0x93, 0x29, 0x4d, 0xe9, 0x50,
	0x65, 0x53, 0x2d, 0xd3, 0x78, 0x9c, 0xab, 0x4c, 0x67, 0xd8, 0xa3, 0x34, 0x96, 0x29, 0x79, 0xf7,
	0xe2, 0x2c, 0x8b, 0x13, 0x0a, 0x78, 0x2e, 0x03, 0x9e, 0xa6, 0x99, 0xe6, 0x5a, 0x66, 0x69, 0x51,
	0x65, 0xf9, 0x11, 0xb8, 0xaf, 0x28, 0x21, 0x4d, 0x6f, 0xca, 0x12, 0x61, 0x55, 0x21, 0xa4, 0x93,
	0x29, 0x15, 0x1a, 0x5d, 0xe8, 0x57, 0x95, 0xa5, 0x70, 0xd8, 0x80, 0x0d, 0x3b, 0xe1, 0x96, 0x89,
	0xf7, 0x05, 0xde, 0x85, 0x6d, 0x91, 0x1d, 0x73, 0x99, 0x96, 0x77, 0x96, 0xb9, 0xeb, 0x57, 0x07,
	0xfb, 0x02, 0x6f, 0x81, 0x25, 0x85, 0xd3, 0x31, 0xa7, 0x96, 0x14, 0xfe, 0x2f, 0x06, 0xee, 0xdb,
	0x5c, 0xf0, 0xcb, 0x51, 0xaa, 0x6c, 0x56, 0x67, 0x37, 0x50, 0xad, 0x26, 0xaa, 0x03, 0x5b, 0x39,
	0xd7, 0x9a, 0x54, 0x6a, 0xaa, 0x6f, 0x87, 0x75, 0x88, 0x1e, 0xf4, 0x73, 0x25, 0x33, 0x25, 0xf5,
	0x07, 0xa7, 0x3b, 0x60, 0x43, 0x3b, 0xbc, 0x88, 0xcb, 0x3b, 0x21, 0x0b, 0x7e, 0x94, 0x90, 0x70,
	0xec, 0x01, 0x1b, 0xf6, 0xc3, 0x8b, 0xb8, 0xd9, 0x47, 0xaf, 0xd9, 0x87, 0xcf, 0xe1, 0x4e, 0x48,
	0x5c, 0xfc, 0x4f, 0x69, 0x66, 0xe0, 0x1e, 0x10, 0x57, 0xd1, 0xbb, 0x7f, 0x09, 0x82, 0xd0, 0x2d,
	0xe4, 0x47, 0x32, 0x30, 0x76, 0x68, 0xbe, 0xcb, 0xb3, 0x9c, 0xc7, 0x74, 0x2e, 0x8e, 0xf9, 0xf6,
	0x5f, 0xc0, 0xce, 0x6b, 0x59, 0xe8, 0x45, 0x68, 0x1c, 0x81, 0x2d, 0x35, 0x1d, 0x17, 0x0e, 0x1b,
	0x74, 0x86, 0x37, 0x26, 0xb7, 0xc7, 0xd5, 0x18, 0x8d, 0x1b, 0xfc, 0xaa, 0x14, 0xff, 0x07, 0x03,
	0xf7, 0xa5, 0x22, 0xbe, 0xf1, 0xf4, 0x2c, 0xf8, 0x68, 0xb5, 0xfb, 0xd8, 0x59, 0xe3, 0x63, 0x77,
	0x9d, 0x8f, 0xf6, 0x92, 0x8f, 0xdf, 0x18, 0xdc, 0x6c, 0x34, 0x79, 0x9d, 0x23, 0x37, 0xf9, 0x69,
	0xc3, 0xee, 0x22, 0x9b, 0x03, 0x52, 0xef, 0x65, 0x44, 0xf8, 0x99, 0x01, 0xae, 0xaa, 0x89, 0x0f,
	0x6a, 0x07, 0x5a, 0x95, 0xf6, 0x2e, 0x35, 0xc9, 0x9f, 0x9c, 0xfd, 0xfe, 0xf3, 0xdd, 0xda, 0xf3,
	0x1f, 0x06, 0x11, 0x4f, 0x92, 0xc3, 0x88, 0x52, 0x4d, 0x2a, 0x30, 0x7d, 0x15, 0xc1, 0xac, 0xee,
	0xf7, 0x34, 0x38, 0x5f, 0x1b, 0xc5, 0x73, 0x36, 0xc2, 0x33, 0x06, 0xb8, 0x3a, 0x8f, 0x73, 0x0e,
	0xad, 0xb3, 0xea, 0x39, 0x75, 0xca, 0xf2, 0x44, 0xf9, 0x81, 0xe1, 0xf1, 0x08, 0xaf, 0xca, 0x03,
	0x3f, 0xc1, 0xce, 0xf2, 0xb3, 0xc3, 0xfb, 0x75, 0xf9, 0x96, 0x07, 0xd9, 0xa2, 0xc1, 0x53, 0x83,
	0x3d, 0xc6, 0xbd, 0x2b, 0x62, 0x07, 0x33, 0x29, 0x4e, 0xf1, 0x2b, 0x03, 0x5c, 0xdd, 0x57, 0x73,
	0x15, 0x5a, 0x77, 0x59, 0x0b, 0x8b, 0x67, 0x86, 0xc5, 0x63, 0x6f, 0x23, 0x16, 0xa5, 0x1d, 0x5f,
	0x18, 0xe0, 0xea, 0x7a, 0x9e, 0x13, 0x69, 0x5d, 0xdd, 0xeb, 0xe5, 0x18, 0x6d, 0x44, 0xe4, 0xa8,
	0x67, 0x7e, 0x15, 0x4f, 0xfe, 0x06, 0x00, 0x00, 0xff, 0xff, 0xe8, 0x54, 0x9a, 0x21, 0x67, 0x06,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueueRoutingServiceClient is the client API for QueueRoutingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueueRoutingServiceClient interface {
	// CreateQueueRouting
	CreateQueueRouting(ctx context.Context, in *CreateQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error)
	// SearchQueueRouting
	SearchQueueRouting(ctx context.Context, in *SearchQueueRoutingRequest, opts ...grpc.CallOption) (*ListQueueRouting, error)
	// ReadQueueRouting
	ReadQueueRouting(ctx context.Context, in *ReadQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error)
	// UpdateQueueRouting
	UpdateQueueRouting(ctx context.Context, in *UpdateQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error)
	// DeleteQueueRouting
	DeleteQueueRouting(ctx context.Context, in *DeleteQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error)
}

type queueRoutingServiceClient struct {
	cc *grpc.ClientConn
}

func NewQueueRoutingServiceClient(cc *grpc.ClientConn) QueueRoutingServiceClient {
	return &queueRoutingServiceClient{cc}
}

func (c *queueRoutingServiceClient) CreateQueueRouting(ctx context.Context, in *CreateQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error) {
	out := new(QueueRouting)
	err := c.cc.Invoke(ctx, "/engine.QueueRoutingService/CreateQueueRouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueRoutingServiceClient) SearchQueueRouting(ctx context.Context, in *SearchQueueRoutingRequest, opts ...grpc.CallOption) (*ListQueueRouting, error) {
	out := new(ListQueueRouting)
	err := c.cc.Invoke(ctx, "/engine.QueueRoutingService/SearchQueueRouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueRoutingServiceClient) ReadQueueRouting(ctx context.Context, in *ReadQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error) {
	out := new(QueueRouting)
	err := c.cc.Invoke(ctx, "/engine.QueueRoutingService/ReadQueueRouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueRoutingServiceClient) UpdateQueueRouting(ctx context.Context, in *UpdateQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error) {
	out := new(QueueRouting)
	err := c.cc.Invoke(ctx, "/engine.QueueRoutingService/UpdateQueueRouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueRoutingServiceClient) DeleteQueueRouting(ctx context.Context, in *DeleteQueueRoutingRequest, opts ...grpc.CallOption) (*QueueRouting, error) {
	out := new(QueueRouting)
	err := c.cc.Invoke(ctx, "/engine.QueueRoutingService/DeleteQueueRouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueueRoutingServiceServer is the server API for QueueRoutingService service.
type QueueRoutingServiceServer interface {
	// CreateQueueRouting
	CreateQueueRouting(context.Context, *CreateQueueRoutingRequest) (*QueueRouting, error)
	// SearchQueueRouting
	SearchQueueRouting(context.Context, *SearchQueueRoutingRequest) (*ListQueueRouting, error)
	// ReadQueueRouting
	ReadQueueRouting(context.Context, *ReadQueueRoutingRequest) (*QueueRouting, error)
	// UpdateQueueRouting
	UpdateQueueRouting(context.Context, *UpdateQueueRoutingRequest) (*QueueRouting, error)
	// DeleteQueueRouting
	DeleteQueueRouting(context.Context, *DeleteQueueRoutingRequest) (*QueueRouting, error)
}

// UnimplementedQueueRoutingServiceServer can be embedded to have forward compatible implementations.
type UnimplementedQueueRoutingServiceServer struct {
}

func (*UnimplementedQueueRoutingServiceServer) CreateQueueRouting(ctx context.Context, req *CreateQueueRoutingRequest) (*QueueRouting, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueueRouting not implemented")
}
func (*UnimplementedQueueRoutingServiceServer) SearchQueueRouting(ctx context.Context, req *SearchQueueRoutingRequest) (*ListQueueRouting, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchQueueRouting not implemented")
}
func (*UnimplementedQueueRoutingServiceServer) ReadQueueRouting(ctx context.Context, req *ReadQueueRoutingRequest) (*QueueRouting, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadQueueRouting not implemented")
}
func (*UnimplementedQueueRoutingServiceServer) UpdateQueueRouting(ctx context.Context, req *UpdateQueueRoutingRequest) (*QueueRouting, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateQueueRouting not implemented")
}
func (*UnimplementedQueueRoutingServiceServer) DeleteQueueRouting(ctx context.Context, req *DeleteQueueRoutingRequest) (*QueueRouting, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteQueueRouting not implemented")
}

func RegisterQueueRoutingServiceServer(s *grpc.Server, srv QueueRoutingServiceServer) {
	s.RegisterService(&_QueueRoutingService_serviceDesc, srv)
}

func _QueueRoutingService_CreateQueueRouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateQueueRoutingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueRoutingServiceServer).CreateQueueRouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.QueueRoutingService/CreateQueueRouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueRoutingServiceServer).CreateQueueRouting(ctx, req.(*CreateQueueRoutingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueRoutingService_SearchQueueRouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchQueueRoutingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueRoutingServiceServer).SearchQueueRouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.QueueRoutingService/SearchQueueRouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueRoutingServiceServer).SearchQueueRouting(ctx, req.(*SearchQueueRoutingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueRoutingService_ReadQueueRouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadQueueRoutingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueRoutingServiceServer).ReadQueueRouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.QueueRoutingService/ReadQueueRouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueRoutingServiceServer).ReadQueueRouting(ctx, req.(*ReadQueueRoutingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueRoutingService_UpdateQueueRouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateQueueRoutingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueRoutingServiceServer).UpdateQueueRouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.QueueRoutingService/UpdateQueueRouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueRoutingServiceServer).UpdateQueueRouting(ctx, req.(*UpdateQueueRoutingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueRoutingService_DeleteQueueRouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteQueueRoutingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueRoutingServiceServer).DeleteQueueRouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.QueueRoutingService/DeleteQueueRouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueRoutingServiceServer).DeleteQueueRouting(ctx, req.(*DeleteQueueRoutingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _QueueRoutingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "engine.QueueRoutingService",
	HandlerType: (*QueueRoutingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateQueueRouting",
			Handler:    _QueueRoutingService_CreateQueueRouting_Handler,
		},
		{
			MethodName: "SearchQueueRouting",
			Handler:    _QueueRoutingService_SearchQueueRouting_Handler,
		},
		{
			MethodName: "ReadQueueRouting",
			Handler:    _QueueRoutingService_ReadQueueRouting_Handler,
		},
		{
			MethodName: "UpdateQueueRouting",
			Handler:    _QueueRoutingService_UpdateQueueRouting_Handler,
		},
		{
			MethodName: "DeleteQueueRouting",
			Handler:    _QueueRoutingService_DeleteQueueRouting_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "queue_routing.proto",
}