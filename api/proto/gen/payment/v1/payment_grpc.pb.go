// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: payment/v1/payment.proto

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH

package pmtv1

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
	WechatPaymentService_NativePrepay_FullMethodName = "/pmt.v1.WechatPaymentService/NativePrepay"
	WechatPaymentService_GetPayment_FullMethodName   = "/pmt.v1.WechatPaymentService/GetPayment"
)

// WechatPaymentServiceClient is the client API for WechatPaymentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WechatPaymentServiceClient interface {
	//  这个设计是认为，Prepay 的请求应该是不同的支付方式都是一样的
	// 但是我们认为响应会是不一样的
	// buf:lint:ignore RPC_REQUEST_STANDARD_NAME
	NativePrepay(ctx context.Context, in *PrepayRequest, opts ...grpc.CallOption) (*NativePrepayResponse, error)
	GetPayment(ctx context.Context, in *GetPaymentRequest, opts ...grpc.CallOption) (*GetPaymentResponse, error)
}

type wechatPaymentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWechatPaymentServiceClient(cc grpc.ClientConnInterface) WechatPaymentServiceClient {
	return &wechatPaymentServiceClient{cc}
}

func (c *wechatPaymentServiceClient) NativePrepay(ctx context.Context, in *PrepayRequest, opts ...grpc.CallOption) (*NativePrepayResponse, error) {
	out := new(NativePrepayResponse)
	err := c.cc.Invoke(ctx, WechatPaymentService_NativePrepay_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wechatPaymentServiceClient) GetPayment(ctx context.Context, in *GetPaymentRequest, opts ...grpc.CallOption) (*GetPaymentResponse, error) {
	out := new(GetPaymentResponse)
	err := c.cc.Invoke(ctx, WechatPaymentService_GetPayment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WechatPaymentServiceServer is the server API for WechatPaymentService service.
// All implementations must embed UnimplementedWechatPaymentServiceServer
// for forward compatibility
type WechatPaymentServiceServer interface {
	//  这个设计是认为，Prepay 的请求应该是不同的支付方式都是一样的
	// 但是我们认为响应会是不一样的
	// buf:lint:ignore RPC_REQUEST_STANDARD_NAME
	NativePrepay(context.Context, *PrepayRequest) (*NativePrepayResponse, error)
	GetPayment(context.Context, *GetPaymentRequest) (*GetPaymentResponse, error)
	mustEmbedUnimplementedWechatPaymentServiceServer()
}

// UnimplementedWechatPaymentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWechatPaymentServiceServer struct {
}

func (UnimplementedWechatPaymentServiceServer) NativePrepay(context.Context, *PrepayRequest) (*NativePrepayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NativePrepay not implemented")
}
func (UnimplementedWechatPaymentServiceServer) GetPayment(context.Context, *GetPaymentRequest) (*GetPaymentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPayment not implemented")
}
func (UnimplementedWechatPaymentServiceServer) mustEmbedUnimplementedWechatPaymentServiceServer() {}

// UnsafeWechatPaymentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WechatPaymentServiceServer will
// result in compilation errors.
type UnsafeWechatPaymentServiceServer interface {
	mustEmbedUnimplementedWechatPaymentServiceServer()
}

func RegisterWechatPaymentServiceServer(s grpc.ServiceRegistrar, srv WechatPaymentServiceServer) {
	s.RegisterService(&WechatPaymentService_ServiceDesc, srv)
}

func _WechatPaymentService_NativePrepay_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepayRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WechatPaymentServiceServer).NativePrepay(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WechatPaymentService_NativePrepay_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WechatPaymentServiceServer).NativePrepay(ctx, req.(*PrepayRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WechatPaymentService_GetPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WechatPaymentServiceServer).GetPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WechatPaymentService_GetPayment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WechatPaymentServiceServer).GetPayment(ctx, req.(*GetPaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WechatPaymentService_ServiceDesc is the grpc.ServiceDesc for WechatPaymentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WechatPaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pmt.v1.WechatPaymentService",
	HandlerType: (*WechatPaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NativePrepay",
			Handler:    _WechatPaymentService_NativePrepay_Handler,
		},
		{
			MethodName: "GetPayment",
			Handler:    _WechatPaymentService_GetPayment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment/v1/payment.proto",
}
