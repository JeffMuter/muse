// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.28.3
// source: proto/data.proto

package proto

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
	ParrotService_SendAlertData_FullMethodName = "/dataservice.ParrotService/SendAlertData"
)

// ParrotServiceClient is the client API for ParrotService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ParrotServiceClient interface {
	SendAlertData(ctx context.Context, in *AlertDataRequest, opts ...grpc.CallOption) (*AlertDataResponse, error)
}

type parrotServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewParrotServiceClient(cc grpc.ClientConnInterface) ParrotServiceClient {
	return &parrotServiceClient{cc}
}

func (c *parrotServiceClient) SendAlertData(ctx context.Context, in *AlertDataRequest, opts ...grpc.CallOption) (*AlertDataResponse, error) {
	out := new(AlertDataResponse)
	err := c.cc.Invoke(ctx, ParrotService_SendAlertData_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ParrotServiceServer is the server API for ParrotService service.
// All implementations must embed UnimplementedParrotServiceServer
// for forward compatibility
type ParrotServiceServer interface {
	SendAlertData(context.Context, *AlertDataRequest) (*AlertDataResponse, error)
	mustEmbedUnimplementedParrotServiceServer()
}

// UnimplementedParrotServiceServer must be embedded to have forward compatible implementations.
type UnimplementedParrotServiceServer struct {
}

func (UnimplementedParrotServiceServer) SendAlertData(context.Context, *AlertDataRequest) (*AlertDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendAlertData not implemented")
}
func (UnimplementedParrotServiceServer) mustEmbedUnimplementedParrotServiceServer() {}

// UnsafeParrotServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ParrotServiceServer will
// result in compilation errors.
type UnsafeParrotServiceServer interface {
	mustEmbedUnimplementedParrotServiceServer()
}

func RegisterParrotServiceServer(s grpc.ServiceRegistrar, srv ParrotServiceServer) {
	s.RegisterService(&ParrotService_ServiceDesc, srv)
}

func _ParrotService_SendAlertData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AlertDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ParrotServiceServer).SendAlertData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ParrotService_SendAlertData_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ParrotServiceServer).SendAlertData(ctx, req.(*AlertDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ParrotService_ServiceDesc is the grpc.ServiceDesc for ParrotService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ParrotService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dataservice.ParrotService",
	HandlerType: (*ParrotServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendAlertData",
			Handler:    _ParrotService_SendAlertData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/data.proto",
}
