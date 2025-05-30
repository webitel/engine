// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: queue_hook.proto

package engine

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
	QueueHookService_CreateQueueHook_FullMethodName = "/engine.QueueHookService/CreateQueueHook"
	QueueHookService_SearchQueueHook_FullMethodName = "/engine.QueueHookService/SearchQueueHook"
	QueueHookService_ReadQueueHook_FullMethodName   = "/engine.QueueHookService/ReadQueueHook"
	QueueHookService_UpdateQueueHook_FullMethodName = "/engine.QueueHookService/UpdateQueueHook"
	QueueHookService_PatchQueueHook_FullMethodName  = "/engine.QueueHookService/PatchQueueHook"
	QueueHookService_DeleteQueueHook_FullMethodName = "/engine.QueueHookService/DeleteQueueHook"
)

// QueueHookServiceClient is the client API for QueueHookService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueueHookServiceClient interface {
	CreateQueueHook(ctx context.Context, in *CreateQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error)
	SearchQueueHook(ctx context.Context, in *SearchQueueHookRequest, opts ...grpc.CallOption) (*ListQueueHook, error)
	ReadQueueHook(ctx context.Context, in *ReadQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error)
	UpdateQueueHook(ctx context.Context, in *UpdateQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error)
	PatchQueueHook(ctx context.Context, in *PatchQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error)
	DeleteQueueHook(ctx context.Context, in *DeleteQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error)
}

type queueHookServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewQueueHookServiceClient(cc grpc.ClientConnInterface) QueueHookServiceClient {
	return &queueHookServiceClient{cc}
}

func (c *queueHookServiceClient) CreateQueueHook(ctx context.Context, in *CreateQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error) {
	out := new(QueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_CreateQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueHookServiceClient) SearchQueueHook(ctx context.Context, in *SearchQueueHookRequest, opts ...grpc.CallOption) (*ListQueueHook, error) {
	out := new(ListQueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_SearchQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueHookServiceClient) ReadQueueHook(ctx context.Context, in *ReadQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error) {
	out := new(QueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_ReadQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueHookServiceClient) UpdateQueueHook(ctx context.Context, in *UpdateQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error) {
	out := new(QueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_UpdateQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueHookServiceClient) PatchQueueHook(ctx context.Context, in *PatchQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error) {
	out := new(QueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_PatchQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queueHookServiceClient) DeleteQueueHook(ctx context.Context, in *DeleteQueueHookRequest, opts ...grpc.CallOption) (*QueueHook, error) {
	out := new(QueueHook)
	err := c.cc.Invoke(ctx, QueueHookService_DeleteQueueHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueueHookServiceServer is the server API for QueueHookService service.
// All implementations must embed UnimplementedQueueHookServiceServer
// for forward compatibility
type QueueHookServiceServer interface {
	CreateQueueHook(context.Context, *CreateQueueHookRequest) (*QueueHook, error)
	SearchQueueHook(context.Context, *SearchQueueHookRequest) (*ListQueueHook, error)
	ReadQueueHook(context.Context, *ReadQueueHookRequest) (*QueueHook, error)
	UpdateQueueHook(context.Context, *UpdateQueueHookRequest) (*QueueHook, error)
	PatchQueueHook(context.Context, *PatchQueueHookRequest) (*QueueHook, error)
	DeleteQueueHook(context.Context, *DeleteQueueHookRequest) (*QueueHook, error)
	mustEmbedUnimplementedQueueHookServiceServer()
}

// UnimplementedQueueHookServiceServer must be embedded to have forward compatible implementations.
type UnimplementedQueueHookServiceServer struct {
}

func (UnimplementedQueueHookServiceServer) CreateQueueHook(context.Context, *CreateQueueHookRequest) (*QueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) SearchQueueHook(context.Context, *SearchQueueHookRequest) (*ListQueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) ReadQueueHook(context.Context, *ReadQueueHookRequest) (*QueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) UpdateQueueHook(context.Context, *UpdateQueueHookRequest) (*QueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) PatchQueueHook(context.Context, *PatchQueueHookRequest) (*QueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PatchQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) DeleteQueueHook(context.Context, *DeleteQueueHookRequest) (*QueueHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteQueueHook not implemented")
}
func (UnimplementedQueueHookServiceServer) mustEmbedUnimplementedQueueHookServiceServer() {}

// UnsafeQueueHookServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueueHookServiceServer will
// result in compilation errors.
type UnsafeQueueHookServiceServer interface {
	mustEmbedUnimplementedQueueHookServiceServer()
}

func RegisterQueueHookServiceServer(s grpc.ServiceRegistrar, srv QueueHookServiceServer) {
	s.RegisterService(&QueueHookService_ServiceDesc, srv)
}

func _QueueHookService_CreateQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).CreateQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_CreateQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).CreateQueueHook(ctx, req.(*CreateQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueHookService_SearchQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).SearchQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_SearchQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).SearchQueueHook(ctx, req.(*SearchQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueHookService_ReadQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).ReadQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_ReadQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).ReadQueueHook(ctx, req.(*ReadQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueHookService_UpdateQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).UpdateQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_UpdateQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).UpdateQueueHook(ctx, req.(*UpdateQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueHookService_PatchQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).PatchQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_PatchQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).PatchQueueHook(ctx, req.(*PatchQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueueHookService_DeleteQueueHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteQueueHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueueHookServiceServer).DeleteQueueHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueueHookService_DeleteQueueHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueueHookServiceServer).DeleteQueueHook(ctx, req.(*DeleteQueueHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// QueueHookService_ServiceDesc is the grpc.ServiceDesc for QueueHookService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var QueueHookService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "engine.QueueHookService",
	HandlerType: (*QueueHookServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateQueueHook",
			Handler:    _QueueHookService_CreateQueueHook_Handler,
		},
		{
			MethodName: "SearchQueueHook",
			Handler:    _QueueHookService_SearchQueueHook_Handler,
		},
		{
			MethodName: "ReadQueueHook",
			Handler:    _QueueHookService_ReadQueueHook_Handler,
		},
		{
			MethodName: "UpdateQueueHook",
			Handler:    _QueueHookService_UpdateQueueHook_Handler,
		},
		{
			MethodName: "PatchQueueHook",
			Handler:    _QueueHookService_PatchQueueHook_Handler,
		},
		{
			MethodName: "DeleteQueueHook",
			Handler:    _QueueHookService_DeleteQueueHook_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "queue_hook.proto",
}
