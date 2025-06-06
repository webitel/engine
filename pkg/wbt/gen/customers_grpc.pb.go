// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: customers.proto

package api

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
	Customers_ServerInfo_FullMethodName     = "/api.Customers/ServerInfo"
	Customers_GetCustomer_FullMethodName    = "/api.Customers/GetCustomer"
	Customers_UpdateCustomer_FullMethodName = "/api.Customers/UpdateCustomer"
	Customers_LicenseUsage_FullMethodName   = "/api.Customers/LicenseUsage"
	Customers_LicenseUsers_FullMethodName   = "/api.Customers/LicenseUsers"
)

// CustomersClient is the client API for Customers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CustomersClient interface {
	ServerInfo(ctx context.Context, in *ServerInfoRequest, opts ...grpc.CallOption) (*ServerInfoResponse, error)
	// rpc GetCertificate(CertificateUsageRequest) returns (CertificateUsageResponse) {}
	GetCustomer(ctx context.Context, in *GetCustomerRequest, opts ...grpc.CallOption) (*GetCustomerResponse, error)
	UpdateCustomer(ctx context.Context, in *UpdateCustomerRequest, opts ...grpc.CallOption) (*UpdateCustomerResponse, error)
	LicenseUsage(ctx context.Context, in *LicenseUsageRequest, opts ...grpc.CallOption) (*LicenseUsageResponse, error)
	LicenseUsers(ctx context.Context, in *LicenseUsersRequest, opts ...grpc.CallOption) (*LicenseUsersResponse, error)
}

type customersClient struct {
	cc grpc.ClientConnInterface
}

func NewCustomersClient(cc grpc.ClientConnInterface) CustomersClient {
	return &customersClient{cc}
}

func (c *customersClient) ServerInfo(ctx context.Context, in *ServerInfoRequest, opts ...grpc.CallOption) (*ServerInfoResponse, error) {
	out := new(ServerInfoResponse)
	err := c.cc.Invoke(ctx, Customers_ServerInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customersClient) GetCustomer(ctx context.Context, in *GetCustomerRequest, opts ...grpc.CallOption) (*GetCustomerResponse, error) {
	out := new(GetCustomerResponse)
	err := c.cc.Invoke(ctx, Customers_GetCustomer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customersClient) UpdateCustomer(ctx context.Context, in *UpdateCustomerRequest, opts ...grpc.CallOption) (*UpdateCustomerResponse, error) {
	out := new(UpdateCustomerResponse)
	err := c.cc.Invoke(ctx, Customers_UpdateCustomer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customersClient) LicenseUsage(ctx context.Context, in *LicenseUsageRequest, opts ...grpc.CallOption) (*LicenseUsageResponse, error) {
	out := new(LicenseUsageResponse)
	err := c.cc.Invoke(ctx, Customers_LicenseUsage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customersClient) LicenseUsers(ctx context.Context, in *LicenseUsersRequest, opts ...grpc.CallOption) (*LicenseUsersResponse, error) {
	out := new(LicenseUsersResponse)
	err := c.cc.Invoke(ctx, Customers_LicenseUsers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CustomersServer is the server API for Customers service.
// All implementations must embed UnimplementedCustomersServer
// for forward compatibility
type CustomersServer interface {
	ServerInfo(context.Context, *ServerInfoRequest) (*ServerInfoResponse, error)
	// rpc GetCertificate(CertificateUsageRequest) returns (CertificateUsageResponse) {}
	GetCustomer(context.Context, *GetCustomerRequest) (*GetCustomerResponse, error)
	UpdateCustomer(context.Context, *UpdateCustomerRequest) (*UpdateCustomerResponse, error)
	LicenseUsage(context.Context, *LicenseUsageRequest) (*LicenseUsageResponse, error)
	LicenseUsers(context.Context, *LicenseUsersRequest) (*LicenseUsersResponse, error)
	mustEmbedUnimplementedCustomersServer()
}

// UnimplementedCustomersServer must be embedded to have forward compatible implementations.
type UnimplementedCustomersServer struct {
}

func (UnimplementedCustomersServer) ServerInfo(context.Context, *ServerInfoRequest) (*ServerInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerInfo not implemented")
}
func (UnimplementedCustomersServer) GetCustomer(context.Context, *GetCustomerRequest) (*GetCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCustomer not implemented")
}
func (UnimplementedCustomersServer) UpdateCustomer(context.Context, *UpdateCustomerRequest) (*UpdateCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCustomer not implemented")
}
func (UnimplementedCustomersServer) LicenseUsage(context.Context, *LicenseUsageRequest) (*LicenseUsageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LicenseUsage not implemented")
}
func (UnimplementedCustomersServer) LicenseUsers(context.Context, *LicenseUsersRequest) (*LicenseUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LicenseUsers not implemented")
}
func (UnimplementedCustomersServer) mustEmbedUnimplementedCustomersServer() {}

// UnsafeCustomersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CustomersServer will
// result in compilation errors.
type UnsafeCustomersServer interface {
	mustEmbedUnimplementedCustomersServer()
}

func RegisterCustomersServer(s grpc.ServiceRegistrar, srv CustomersServer) {
	s.RegisterService(&Customers_ServiceDesc, srv)
}

func _Customers_ServerInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomersServer).ServerInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Customers_ServerInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomersServer).ServerInfo(ctx, req.(*ServerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customers_GetCustomer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCustomerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomersServer).GetCustomer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Customers_GetCustomer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomersServer).GetCustomer(ctx, req.(*GetCustomerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customers_UpdateCustomer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCustomerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomersServer).UpdateCustomer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Customers_UpdateCustomer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomersServer).UpdateCustomer(ctx, req.(*UpdateCustomerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customers_LicenseUsage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LicenseUsageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomersServer).LicenseUsage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Customers_LicenseUsage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomersServer).LicenseUsage(ctx, req.(*LicenseUsageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customers_LicenseUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LicenseUsersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomersServer).LicenseUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Customers_LicenseUsers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomersServer).LicenseUsers(ctx, req.(*LicenseUsersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Customers_ServiceDesc is the grpc.ServiceDesc for Customers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Customers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.Customers",
	HandlerType: (*CustomersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ServerInfo",
			Handler:    _Customers_ServerInfo_Handler,
		},
		{
			MethodName: "GetCustomer",
			Handler:    _Customers_GetCustomer_Handler,
		},
		{
			MethodName: "UpdateCustomer",
			Handler:    _Customers_UpdateCustomer_Handler,
		},
		{
			MethodName: "LicenseUsage",
			Handler:    _Customers_LicenseUsage_Handler,
		},
		{
			MethodName: "LicenseUsers",
			Handler:    _Customers_LicenseUsers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "customers.proto",
}
