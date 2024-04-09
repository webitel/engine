// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: contacts/upload.proto

package contacts

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Upload_UploadMedia_FullMethodName = "/webitel.contacts.Upload/UploadMedia"
)

// UploadClient is the client API for Upload service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UploadClient interface {
	// Upload an image or photo
	UploadMedia(ctx context.Context, opts ...grpc.CallOption) (Upload_UploadMediaClient, error)
}

type uploadClient struct {
	cc grpc.ClientConnInterface
}

func NewUploadClient(cc grpc.ClientConnInterface) UploadClient {
	return &uploadClient{cc}
}

func (c *uploadClient) UploadMedia(ctx context.Context, opts ...grpc.CallOption) (Upload_UploadMediaClient, error) {
	stream, err := c.cc.NewStream(ctx, &Upload_ServiceDesc.Streams[0], Upload_UploadMedia_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &uploadUploadMediaClient{stream}
	return x, nil
}

type Upload_UploadMediaClient interface {
	Send(*UploadMediaRequest) error
	Recv() (*UploadMediaResponse, error)
	grpc.ClientStream
}

type uploadUploadMediaClient struct {
	grpc.ClientStream
}

func (x *uploadUploadMediaClient) Send(m *UploadMediaRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *uploadUploadMediaClient) Recv() (*UploadMediaResponse, error) {
	m := new(UploadMediaResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UploadServer is the server API for Upload service.
// All implementations must embed UnimplementedUploadServer
// for forward compatibility
type UploadServer interface {
	// Upload an image or photo
	UploadMedia(Upload_UploadMediaServer) error
	mustEmbedUnimplementedUploadServer()
}

// UnimplementedUploadServer must be embedded to have forward compatible implementations.
type UnimplementedUploadServer struct {
}

func (UnimplementedUploadServer) UploadMedia(Upload_UploadMediaServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadMedia not implemented")
}
func (UnimplementedUploadServer) mustEmbedUnimplementedUploadServer() {}

// UnsafeUploadServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UploadServer will
// result in compilation errors.
type UnsafeUploadServer interface {
	mustEmbedUnimplementedUploadServer()
}

func RegisterUploadServer(s grpc.ServiceRegistrar, srv UploadServer) {
	s.RegisterService(&Upload_ServiceDesc, srv)
}

func _Upload_UploadMedia_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(UploadServer).UploadMedia(&uploadUploadMediaServer{stream})
}

type Upload_UploadMediaServer interface {
	Send(*UploadMediaResponse) error
	Recv() (*UploadMediaRequest, error)
	grpc.ServerStream
}

type uploadUploadMediaServer struct {
	grpc.ServerStream
}

func (x *uploadUploadMediaServer) Send(m *UploadMediaResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *uploadUploadMediaServer) Recv() (*UploadMediaRequest, error) {
	m := new(UploadMediaRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Upload_ServiceDesc is the grpc.ServiceDesc for Upload service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Upload_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "webitel.contacts.Upload",
	HandlerType: (*UploadServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadMedia",
			Handler:       _Upload_UploadMedia_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "contacts/upload.proto",
}
