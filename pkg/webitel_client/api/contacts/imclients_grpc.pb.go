// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.6
// source: contacts/imclients.proto

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
	IMClients_LocateIMClient_FullMethodName  = "/webitel.contacts.IMClients/LocateIMClient"
	IMClients_SearchIMClients_FullMethodName = "/webitel.contacts.IMClients/SearchIMClients"
	IMClients_CreateIMClients_FullMethodName = "/webitel.contacts.IMClients/CreateIMClients"
	IMClients_UpdateIMClients_FullMethodName = "/webitel.contacts.IMClients/UpdateIMClients"
	IMClients_UpdateIMClient_FullMethodName  = "/webitel.contacts.IMClients/UpdateIMClient"
	IMClients_DeleteIMClients_FullMethodName = "/webitel.contacts.IMClients/DeleteIMClients"
	IMClients_DeleteIMClient_FullMethodName  = "/webitel.contacts.IMClients/DeleteIMClient"
)

// IMClientsClient is the client API for IMClients service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IMClientsClient interface {
	// Locate the IM client link.
	LocateIMClient(ctx context.Context, in *LocateIMClientRequest, opts ...grpc.CallOption) (*IMClient, error)
	// Search IM client links
	SearchIMClients(ctx context.Context, in *SearchIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error)
	// Link IM client(s) with the Contact
	CreateIMClients(ctx context.Context, in *CreateIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error)
	// Reset the Contact's IM clients to fit given data set.
	UpdateIMClients(ctx context.Context, in *UpdateIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error)
	// Update the Contact's IM client link
	UpdateIMClient(ctx context.Context, in *UpdateIMClientRequest, opts ...grpc.CallOption) (*IMClient, error)
	// Remove the Contact's IM client link(s)
	DeleteIMClients(ctx context.Context, in *DeleteIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error)
	// Remove the Contact's IM client link
	DeleteIMClient(ctx context.Context, in *DeleteIMClientRequest, opts ...grpc.CallOption) (*IMClient, error)
}

type iMClientsClient struct {
	cc grpc.ClientConnInterface
}

func NewIMClientsClient(cc grpc.ClientConnInterface) IMClientsClient {
	return &iMClientsClient{cc}
}

func (c *iMClientsClient) LocateIMClient(ctx context.Context, in *LocateIMClientRequest, opts ...grpc.CallOption) (*IMClient, error) {
	out := new(IMClient)
	err := c.cc.Invoke(ctx, IMClients_LocateIMClient_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) SearchIMClients(ctx context.Context, in *SearchIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error) {
	out := new(IMClientList)
	err := c.cc.Invoke(ctx, IMClients_SearchIMClients_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) CreateIMClients(ctx context.Context, in *CreateIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error) {
	out := new(IMClientList)
	err := c.cc.Invoke(ctx, IMClients_CreateIMClients_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) UpdateIMClients(ctx context.Context, in *UpdateIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error) {
	out := new(IMClientList)
	err := c.cc.Invoke(ctx, IMClients_UpdateIMClients_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) UpdateIMClient(ctx context.Context, in *UpdateIMClientRequest, opts ...grpc.CallOption) (*IMClient, error) {
	out := new(IMClient)
	err := c.cc.Invoke(ctx, IMClients_UpdateIMClient_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) DeleteIMClients(ctx context.Context, in *DeleteIMClientsRequest, opts ...grpc.CallOption) (*IMClientList, error) {
	out := new(IMClientList)
	err := c.cc.Invoke(ctx, IMClients_DeleteIMClients_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMClientsClient) DeleteIMClient(ctx context.Context, in *DeleteIMClientRequest, opts ...grpc.CallOption) (*IMClient, error) {
	out := new(IMClient)
	err := c.cc.Invoke(ctx, IMClients_DeleteIMClient_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IMClientsServer is the server API for IMClients service.
// All implementations must embed UnimplementedIMClientsServer
// for forward compatibility
type IMClientsServer interface {
	// Locate the IM client link.
	LocateIMClient(context.Context, *LocateIMClientRequest) (*IMClient, error)
	// Search IM client links
	SearchIMClients(context.Context, *SearchIMClientsRequest) (*IMClientList, error)
	// Link IM client(s) with the Contact
	CreateIMClients(context.Context, *CreateIMClientsRequest) (*IMClientList, error)
	// Reset the Contact's IM clients to fit given data set.
	UpdateIMClients(context.Context, *UpdateIMClientsRequest) (*IMClientList, error)
	// Update the Contact's IM client link
	UpdateIMClient(context.Context, *UpdateIMClientRequest) (*IMClient, error)
	// Remove the Contact's IM client link(s)
	DeleteIMClients(context.Context, *DeleteIMClientsRequest) (*IMClientList, error)
	// Remove the Contact's IM client link
	DeleteIMClient(context.Context, *DeleteIMClientRequest) (*IMClient, error)
	mustEmbedUnimplementedIMClientsServer()
}

// UnimplementedIMClientsServer must be embedded to have forward compatible implementations.
type UnimplementedIMClientsServer struct {
}

func (UnimplementedIMClientsServer) LocateIMClient(context.Context, *LocateIMClientRequest) (*IMClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LocateIMClient not implemented")
}
func (UnimplementedIMClientsServer) SearchIMClients(context.Context, *SearchIMClientsRequest) (*IMClientList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchIMClients not implemented")
}
func (UnimplementedIMClientsServer) CreateIMClients(context.Context, *CreateIMClientsRequest) (*IMClientList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateIMClients not implemented")
}
func (UnimplementedIMClientsServer) UpdateIMClients(context.Context, *UpdateIMClientsRequest) (*IMClientList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateIMClients not implemented")
}
func (UnimplementedIMClientsServer) UpdateIMClient(context.Context, *UpdateIMClientRequest) (*IMClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateIMClient not implemented")
}
func (UnimplementedIMClientsServer) DeleteIMClients(context.Context, *DeleteIMClientsRequest) (*IMClientList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteIMClients not implemented")
}
func (UnimplementedIMClientsServer) DeleteIMClient(context.Context, *DeleteIMClientRequest) (*IMClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteIMClient not implemented")
}
func (UnimplementedIMClientsServer) mustEmbedUnimplementedIMClientsServer() {}

// UnsafeIMClientsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IMClientsServer will
// result in compilation errors.
type UnsafeIMClientsServer interface {
	mustEmbedUnimplementedIMClientsServer()
}

func RegisterIMClientsServer(s grpc.ServiceRegistrar, srv IMClientsServer) {
	s.RegisterService(&IMClients_ServiceDesc, srv)
}

func _IMClients_LocateIMClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocateIMClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).LocateIMClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_LocateIMClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).LocateIMClient(ctx, req.(*LocateIMClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_SearchIMClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchIMClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).SearchIMClients(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_SearchIMClients_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).SearchIMClients(ctx, req.(*SearchIMClientsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_CreateIMClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateIMClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).CreateIMClients(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_CreateIMClients_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).CreateIMClients(ctx, req.(*CreateIMClientsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_UpdateIMClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateIMClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).UpdateIMClients(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_UpdateIMClients_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).UpdateIMClients(ctx, req.(*UpdateIMClientsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_UpdateIMClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateIMClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).UpdateIMClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_UpdateIMClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).UpdateIMClient(ctx, req.(*UpdateIMClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_DeleteIMClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIMClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).DeleteIMClients(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_DeleteIMClients_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).DeleteIMClients(ctx, req.(*DeleteIMClientsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMClients_DeleteIMClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIMClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMClientsServer).DeleteIMClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMClients_DeleteIMClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMClientsServer).DeleteIMClient(ctx, req.(*DeleteIMClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IMClients_ServiceDesc is the grpc.ServiceDesc for IMClients service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IMClients_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "webitel.contacts.IMClients",
	HandlerType: (*IMClientsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LocateIMClient",
			Handler:    _IMClients_LocateIMClient_Handler,
		},
		{
			MethodName: "SearchIMClients",
			Handler:    _IMClients_SearchIMClients_Handler,
		},
		{
			MethodName: "CreateIMClients",
			Handler:    _IMClients_CreateIMClients_Handler,
		},
		{
			MethodName: "UpdateIMClients",
			Handler:    _IMClients_UpdateIMClients_Handler,
		},
		{
			MethodName: "UpdateIMClient",
			Handler:    _IMClients_UpdateIMClient_Handler,
		},
		{
			MethodName: "DeleteIMClients",
			Handler:    _IMClients_DeleteIMClients_Handler,
		},
		{
			MethodName: "DeleteIMClient",
			Handler:    _IMClients_DeleteIMClient_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contacts/imclients.proto",
}
