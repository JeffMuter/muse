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
	TranscriptService_GetTranscriptSummary_FullMethodName = "/transcript.TranscriptService/GetTranscriptSummary"
)

// TranscriptServiceClient is the client API for TranscriptService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TranscriptServiceClient interface {
	GetTranscriptSummary(ctx context.Context, in *TranscriptRequest, opts ...grpc.CallOption) (*TranscriptSummaryResponse, error)
}

type transcriptServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTranscriptServiceClient(cc grpc.ClientConnInterface) TranscriptServiceClient {
	return &transcriptServiceClient{cc}
}

func (c *transcriptServiceClient) GetTranscriptSummary(ctx context.Context, in *TranscriptRequest, opts ...grpc.CallOption) (*TranscriptSummaryResponse, error) {
	out := new(TranscriptSummaryResponse)
	err := c.cc.Invoke(ctx, TranscriptService_GetTranscriptSummary_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TranscriptServiceServer is the server API for TranscriptService service.
// All implementations must embed UnimplementedTranscriptServiceServer
// for forward compatibility
type TranscriptServiceServer interface {
	GetTranscriptSummary(context.Context, *TranscriptRequest) (*TranscriptSummaryResponse, error)
	mustEmbedUnimplementedTranscriptServiceServer()
}

// UnimplementedTranscriptServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTranscriptServiceServer struct {
}

func (UnimplementedTranscriptServiceServer) GetTranscriptSummary(context.Context, *TranscriptRequest) (*TranscriptSummaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTranscriptSummary not implemented")
}
func (UnimplementedTranscriptServiceServer) mustEmbedUnimplementedTranscriptServiceServer() {}

// UnsafeTranscriptServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TranscriptServiceServer will
// result in compilation errors.
type UnsafeTranscriptServiceServer interface {
	mustEmbedUnimplementedTranscriptServiceServer()
}

func RegisterTranscriptServiceServer(s grpc.ServiceRegistrar, srv TranscriptServiceServer) {
	s.RegisterService(&TranscriptService_ServiceDesc, srv)
}

func _TranscriptService_GetTranscriptSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TranscriptRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranscriptServiceServer).GetTranscriptSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranscriptService_GetTranscriptSummary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranscriptServiceServer).GetTranscriptSummary(ctx, req.(*TranscriptRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TranscriptService_ServiceDesc is the grpc.ServiceDesc for TranscriptService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TranscriptService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "transcript.TranscriptService",
	HandlerType: (*TranscriptServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTranscriptSummary",
			Handler:    _TranscriptService_GetTranscriptSummary_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/data.proto",
}
