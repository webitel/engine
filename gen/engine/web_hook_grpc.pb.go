// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: web_hook.proto

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
	WebHookService_CreateWebHook_FullMethodName = "/engine.WebHookService/CreateWebHook"
	WebHookService_SearchWebHook_FullMethodName = "/engine.WebHookService/SearchWebHook"
	WebHookService_ReadWebHook_FullMethodName   = "/engine.WebHookService/ReadWebHook"
	WebHookService_PatchWebHook_FullMethodName  = "/engine.WebHookService/PatchWebHook"
	WebHookService_UpdateWebHook_FullMethodName = "/engine.WebHookService/UpdateWebHook"
	WebHookService_DeleteWebHook_FullMethodName = "/engine.WebHookService/DeleteWebHook"
)

// WebHookServiceClient is the client API for WebHookService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WebHookServiceClient interface {
	// Create WebHook
	CreateWebHook(ctx context.Context, in *CreateWebHookRequest, opts ...grpc.CallOption) (*WebHook, error)
	// List of WebHook
	SearchWebHook(ctx context.Context, in *SearchWebHookRequest, opts ...grpc.CallOption) (*ListWebHook, error)
	// WebHook item
	ReadWebHook(ctx context.Context, in *ReadWebHookRequest, opts ...grpc.CallOption) (*WebHook, error)
	// Patch WebHook
	PatchWebHook(ctx context.Context, in *PatchWebHookRequest, opts ...grpc.CallOption) (*WebHook, error)
	// Update WebHook
	UpdateWebHook(ctx context.Context, in *UpdateWebHookRequest, opts ...grpc.CallOption) (*WebHook, error)
	// Remove WebHook
	DeleteWebHook(ctx context.Context, in *DeleteWebHookRequest, opts ...grpc.CallOption) (*WebHook, error)
}

type webHookServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWebHookServiceClient(cc grpc.ClientConnInterface) WebHookServiceClient {
	return &webHookServiceClient{cc}
}

func (c *webHookServiceClient) CreateWebHook(ctx context.Context, in *CreateWebHookRequest, opts ...grpc.CallOption) (*WebHook, error) {
	out := new(WebHook)
	err := c.cc.Invoke(ctx, WebHookService_CreateWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webHookServiceClient) SearchWebHook(ctx context.Context, in *SearchWebHookRequest, opts ...grpc.CallOption) (*ListWebHook, error) {
	out := new(ListWebHook)
	err := c.cc.Invoke(ctx, WebHookService_SearchWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webHookServiceClient) ReadWebHook(ctx context.Context, in *ReadWebHookRequest, opts ...grpc.CallOption) (*WebHook, error) {
	out := new(WebHook)
	err := c.cc.Invoke(ctx, WebHookService_ReadWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webHookServiceClient) PatchWebHook(ctx context.Context, in *PatchWebHookRequest, opts ...grpc.CallOption) (*WebHook, error) {
	out := new(WebHook)
	err := c.cc.Invoke(ctx, WebHookService_PatchWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webHookServiceClient) UpdateWebHook(ctx context.Context, in *UpdateWebHookRequest, opts ...grpc.CallOption) (*WebHook, error) {
	out := new(WebHook)
	err := c.cc.Invoke(ctx, WebHookService_UpdateWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *webHookServiceClient) DeleteWebHook(ctx context.Context, in *DeleteWebHookRequest, opts ...grpc.CallOption) (*WebHook, error) {
	out := new(WebHook)
	err := c.cc.Invoke(ctx, WebHookService_DeleteWebHook_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WebHookServiceServer is the server API for WebHookService service.
// All implementations must embed UnimplementedWebHookServiceServer
// for forward compatibility
type WebHookServiceServer interface {
	// Create WebHook
	CreateWebHook(context.Context, *CreateWebHookRequest) (*WebHook, error)
	// List of WebHook
	SearchWebHook(context.Context, *SearchWebHookRequest) (*ListWebHook, error)
	// WebHook item
	ReadWebHook(context.Context, *ReadWebHookRequest) (*WebHook, error)
	// Patch WebHook
	PatchWebHook(context.Context, *PatchWebHookRequest) (*WebHook, error)
	// Update WebHook
	UpdateWebHook(context.Context, *UpdateWebHookRequest) (*WebHook, error)
	// Remove WebHook
	DeleteWebHook(context.Context, *DeleteWebHookRequest) (*WebHook, error)
	mustEmbedUnimplementedWebHookServiceServer()
}

// UnimplementedWebHookServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWebHookServiceServer struct {
}

func (UnimplementedWebHookServiceServer) CreateWebHook(context.Context, *CreateWebHookRequest) (*WebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) SearchWebHook(context.Context, *SearchWebHookRequest) (*ListWebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) ReadWebHook(context.Context, *ReadWebHookRequest) (*WebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) PatchWebHook(context.Context, *PatchWebHookRequest) (*WebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PatchWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) UpdateWebHook(context.Context, *UpdateWebHookRequest) (*WebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) DeleteWebHook(context.Context, *DeleteWebHookRequest) (*WebHook, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteWebHook not implemented")
}
func (UnimplementedWebHookServiceServer) mustEmbedUnimplementedWebHookServiceServer() {}

// UnsafeWebHookServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WebHookServiceServer will
// result in compilation errors.
type UnsafeWebHookServiceServer interface {
	mustEmbedUnimplementedWebHookServiceServer()
}

func RegisterWebHookServiceServer(s grpc.ServiceRegistrar, srv WebHookServiceServer) {
	s.RegisterService(&WebHookService_ServiceDesc, srv)
}

func _WebHookService_CreateWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).CreateWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_CreateWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).CreateWebHook(ctx, req.(*CreateWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WebHookService_SearchWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).SearchWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_SearchWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).SearchWebHook(ctx, req.(*SearchWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WebHookService_ReadWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).ReadWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_ReadWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).ReadWebHook(ctx, req.(*ReadWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WebHookService_PatchWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).PatchWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_PatchWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).PatchWebHook(ctx, req.(*PatchWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WebHookService_UpdateWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).UpdateWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_UpdateWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).UpdateWebHook(ctx, req.(*UpdateWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WebHookService_DeleteWebHook_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteWebHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WebHookServiceServer).DeleteWebHook(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WebHookService_DeleteWebHook_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WebHookServiceServer).DeleteWebHook(ctx, req.(*DeleteWebHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WebHookService_ServiceDesc is the grpc.ServiceDesc for WebHookService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WebHookService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "engine.WebHookService",
	HandlerType: (*WebHookServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateWebHook",
			Handler:    _WebHookService_CreateWebHook_Handler,
		},
		{
			MethodName: "SearchWebHook",
			Handler:    _WebHookService_SearchWebHook_Handler,
		},
		{
			MethodName: "ReadWebHook",
			Handler:    _WebHookService_ReadWebHook_Handler,
		},
		{
			MethodName: "PatchWebHook",
			Handler:    _WebHookService_PatchWebHook_Handler,
		},
		{
			MethodName: "UpdateWebHook",
			Handler:    _WebHookService_UpdateWebHook_Handler,
		},
		{
			MethodName: "DeleteWebHook",
			Handler:    _WebHookService_DeleteWebHook_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "web_hook.proto",
}
