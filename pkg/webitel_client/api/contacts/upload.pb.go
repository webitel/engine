// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.6
// source: contacts/upload.proto

package contacts

import (
	status "google.golang.org/genproto/googleapis/rpc/status"
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

type UploadMediaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Progress
	//
	// Types that are assignable to Input:
	//
	//	*UploadMediaRequest_File
	//	*UploadMediaRequest_Data
	Input isUploadMediaRequest_Input `protobuf_oneof:"input"`
}

func (x *UploadMediaRequest) Reset() {
	*x = UploadMediaRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contacts_upload_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaRequest) ProtoMessage() {}

func (x *UploadMediaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contacts_upload_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaRequest.ProtoReflect.Descriptor instead.
func (*UploadMediaRequest) Descriptor() ([]byte, []int) {
	return file_contacts_upload_proto_rawDescGZIP(), []int{0}
}

func (m *UploadMediaRequest) GetInput() isUploadMediaRequest_Input {
	if m != nil {
		return m.Input
	}
	return nil
}

func (x *UploadMediaRequest) GetFile() *UploadMediaRequest_InputFile {
	if x, ok := x.GetInput().(*UploadMediaRequest_File); ok {
		return x.File
	}
	return nil
}

func (x *UploadMediaRequest) GetData() *UploadMediaRequest_InputData {
	if x, ok := x.GetInput().(*UploadMediaRequest_Data); ok {
		return x.Data
	}
	return nil
}

type isUploadMediaRequest_Input interface {
	isUploadMediaRequest_Input()
}

type UploadMediaRequest_File struct {
	File *UploadMediaRequest_InputFile `protobuf:"bytes,1,opt,name=file,proto3,oneof"`
}

type UploadMediaRequest_Data struct {
	Data *UploadMediaRequest_InputData `protobuf:"bytes,2,opt,name=data,proto3,oneof"`
}

func (*UploadMediaRequest_File) isUploadMediaRequest_Input() {}

func (*UploadMediaRequest_Data) isUploadMediaRequest_Input() {}

type UploadMediaResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Progress
	//
	// Types that are assignable to Output:
	//
	//	*UploadMediaResponse_File
	//	*UploadMediaResponse_Data
	Output isUploadMediaResponse_Output `protobuf_oneof:"output"`
}

func (x *UploadMediaResponse) Reset() {
	*x = UploadMediaResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contacts_upload_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaResponse) ProtoMessage() {}

func (x *UploadMediaResponse) ProtoReflect() protoreflect.Message {
	mi := &file_contacts_upload_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaResponse.ProtoReflect.Descriptor instead.
func (*UploadMediaResponse) Descriptor() ([]byte, []int) {
	return file_contacts_upload_proto_rawDescGZIP(), []int{1}
}

func (m *UploadMediaResponse) GetOutput() isUploadMediaResponse_Output {
	if m != nil {
		return m.Output
	}
	return nil
}

func (x *UploadMediaResponse) GetFile() *MediaFile {
	if x, ok := x.GetOutput().(*UploadMediaResponse_File); ok {
		return x.File
	}
	return nil
}

func (x *UploadMediaResponse) GetData() *UploadMediaResponse_Progress {
	if x, ok := x.GetOutput().(*UploadMediaResponse_Data); ok {
		return x.Data
	}
	return nil
}

type isUploadMediaResponse_Output interface {
	isUploadMediaResponse_Output()
}

type UploadMediaResponse_File struct {
	File *MediaFile `protobuf:"bytes,1,opt,name=file,proto3,oneof"` // START|COMPLETE
}

type UploadMediaResponse_Data struct {
	Data *UploadMediaResponse_Progress `protobuf:"bytes,2,opt,name=data,proto3,oneof"` // PROGRESS
}

func (*UploadMediaResponse_File) isUploadMediaResponse_Output() {}

func (*UploadMediaResponse_Data) isUploadMediaResponse_Output() {}

// File Metadata
type UploadMediaRequest_InputFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type string            `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Size int32             `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	Meta []*MediaAttribute `protobuf:"bytes,3,rep,name=meta,proto3" json:"meta,omitempty"`
}

func (x *UploadMediaRequest_InputFile) Reset() {
	*x = UploadMediaRequest_InputFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contacts_upload_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaRequest_InputFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaRequest_InputFile) ProtoMessage() {}

func (x *UploadMediaRequest_InputFile) ProtoReflect() protoreflect.Message {
	mi := &file_contacts_upload_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaRequest_InputFile.ProtoReflect.Descriptor instead.
func (*UploadMediaRequest_InputFile) Descriptor() ([]byte, []int) {
	return file_contacts_upload_proto_rawDescGZIP(), []int{0, 0}
}

func (x *UploadMediaRequest_InputFile) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *UploadMediaRequest_InputFile) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *UploadMediaRequest_InputFile) GetMeta() []*MediaAttribute {
	if x != nil {
		return x.Meta
	}
	return nil
}

// Multipart Chunk
type UploadMediaRequest_InputData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset uint32         `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Binary []byte         `protobuf:"bytes,2,opt,name=binary,proto3" json:"binary,omitempty"`
	Cancel *status.Status `protobuf:"bytes,3,opt,name=cancel,proto3" json:"cancel,omitempty"`
}

func (x *UploadMediaRequest_InputData) Reset() {
	*x = UploadMediaRequest_InputData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contacts_upload_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaRequest_InputData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaRequest_InputData) ProtoMessage() {}

func (x *UploadMediaRequest_InputData) ProtoReflect() protoreflect.Message {
	mi := &file_contacts_upload_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaRequest_InputData.ProtoReflect.Descriptor instead.
func (*UploadMediaRequest_InputData) Descriptor() ([]byte, []int) {
	return file_contacts_upload_proto_rawDescGZIP(), []int{0, 1}
}

func (x *UploadMediaRequest_InputData) GetOffset() uint32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *UploadMediaRequest_InputData) GetBinary() []byte {
	if x != nil {
		return x.Binary
	}
	return nil
}

func (x *UploadMediaRequest_InputData) GetCancel() *status.Status {
	if x != nil {
		return x.Cancel
	}
	return nil
}

// Upload progress
type UploadMediaResponse_Progress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 0..100
	Percent float32 `protobuf:"fixed32,1,opt,name=percent,proto3" json:"percent,omitempty"`
}

func (x *UploadMediaResponse_Progress) Reset() {
	*x = UploadMediaResponse_Progress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contacts_upload_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMediaResponse_Progress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMediaResponse_Progress) ProtoMessage() {}

func (x *UploadMediaResponse_Progress) ProtoReflect() protoreflect.Message {
	mi := &file_contacts_upload_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMediaResponse_Progress.ProtoReflect.Descriptor instead.
func (*UploadMediaResponse_Progress) Descriptor() ([]byte, []int) {
	return file_contacts_upload_proto_rawDescGZIP(), []int{1, 0}
}

func (x *UploadMediaResponse_Progress) GetPercent() float32 {
	if x != nil {
		return x.Percent
	}
	return 0
}

var File_contacts_upload_proto protoreflect.FileDescriptor

var file_contacts_upload_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c,
	0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x1a, 0x14, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x63, 0x74, 0x73, 0x2f, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfd, 0x02, 0x0a, 0x12, 0x55, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x44, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e,
	0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73,
	0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x48, 0x00, 0x52,
	0x04, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x44, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64,
	0x69, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x48, 0x00, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x69, 0x0a, 0x09, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65,
	0x12, 0x34, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74,
	0x73, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x1a, 0x67, 0x0a, 0x09, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x62,
	0x69, 0x6e, 0x61, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x62, 0x69, 0x6e,
	0x61, 0x72, 0x79, 0x12, 0x2a, 0x0a, 0x06, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x42,
	0x07, 0x0a, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x22, 0xbe, 0x01, 0x0a, 0x13, 0x55, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x31, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74,
	0x73, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x46, 0x69, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x04, 0x66,
	0x69, 0x6c, 0x65, 0x12, 0x44, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2e, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73,
	0x73, 0x48, 0x00, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x24, 0x0a, 0x08, 0x50, 0x72, 0x6f,
	0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x07, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x42,
	0x08, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x32, 0x6a, 0x0a, 0x06, 0x55, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x12, 0x60, 0x0a, 0x0b, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64,
	0x69, 0x61, 0x12, 0x24, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x77, 0x65, 0x62, 0x69, 0x74,
	0x65, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x2e, 0x55, 0x70, 0x6c, 0x6f,
	0x61, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x22, 0x5a, 0x20, 0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c,
	0x2e, 0x67, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73,
	0x3b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_contacts_upload_proto_rawDescOnce sync.Once
	file_contacts_upload_proto_rawDescData = file_contacts_upload_proto_rawDesc
)

func file_contacts_upload_proto_rawDescGZIP() []byte {
	file_contacts_upload_proto_rawDescOnce.Do(func() {
		file_contacts_upload_proto_rawDescData = protoimpl.X.CompressGZIP(file_contacts_upload_proto_rawDescData)
	})
	return file_contacts_upload_proto_rawDescData
}

var file_contacts_upload_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_contacts_upload_proto_goTypes = []interface{}{
	(*UploadMediaRequest)(nil),           // 0: webitel.contacts.UploadMediaRequest
	(*UploadMediaResponse)(nil),          // 1: webitel.contacts.UploadMediaResponse
	(*UploadMediaRequest_InputFile)(nil), // 2: webitel.contacts.UploadMediaRequest.InputFile
	(*UploadMediaRequest_InputData)(nil), // 3: webitel.contacts.UploadMediaRequest.InputData
	(*UploadMediaResponse_Progress)(nil), // 4: webitel.contacts.UploadMediaResponse.Progress
	(*MediaFile)(nil),                    // 5: webitel.contacts.MediaFile
	(*MediaAttribute)(nil),               // 6: webitel.contacts.MediaAttribute
	(*status.Status)(nil),                // 7: google.rpc.Status
}
var file_contacts_upload_proto_depIdxs = []int32{
	2, // 0: webitel.contacts.UploadMediaRequest.file:type_name -> webitel.contacts.UploadMediaRequest.InputFile
	3, // 1: webitel.contacts.UploadMediaRequest.data:type_name -> webitel.contacts.UploadMediaRequest.InputData
	5, // 2: webitel.contacts.UploadMediaResponse.file:type_name -> webitel.contacts.MediaFile
	4, // 3: webitel.contacts.UploadMediaResponse.data:type_name -> webitel.contacts.UploadMediaResponse.Progress
	6, // 4: webitel.contacts.UploadMediaRequest.InputFile.meta:type_name -> webitel.contacts.MediaAttribute
	7, // 5: webitel.contacts.UploadMediaRequest.InputData.cancel:type_name -> google.rpc.Status
	0, // 6: webitel.contacts.Upload.UploadMedia:input_type -> webitel.contacts.UploadMediaRequest
	1, // 7: webitel.contacts.Upload.UploadMedia:output_type -> webitel.contacts.UploadMediaResponse
	7, // [7:8] is the sub-list for method output_type
	6, // [6:7] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_contacts_upload_proto_init() }
func file_contacts_upload_proto_init() {
	if File_contacts_upload_proto != nil {
		return
	}
	file_contacts_media_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_contacts_upload_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaRequest); i {
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
		file_contacts_upload_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaResponse); i {
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
		file_contacts_upload_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaRequest_InputFile); i {
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
		file_contacts_upload_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaRequest_InputData); i {
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
		file_contacts_upload_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMediaResponse_Progress); i {
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
	file_contacts_upload_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*UploadMediaRequest_File)(nil),
		(*UploadMediaRequest_Data)(nil),
	}
	file_contacts_upload_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*UploadMediaResponse_File)(nil),
		(*UploadMediaResponse_Data)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_contacts_upload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_contacts_upload_proto_goTypes,
		DependencyIndexes: file_contacts_upload_proto_depIdxs,
		MessageInfos:      file_contacts_upload_proto_msgTypes,
	}.Build()
	File_contacts_upload_proto = out.File
	file_contacts_upload_proto_rawDesc = nil
	file_contacts_upload_proto_goTypes = nil
	file_contacts_upload_proto_depIdxs = nil
}