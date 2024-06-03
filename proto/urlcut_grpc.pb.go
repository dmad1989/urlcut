// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.0
// source: urlcut.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	UrlCut_CutterJson_FullMethodName      = "/urlcut.UrlCut/CutterJson"
	UrlCut_Cutter_FullMethodName          = "/urlcut.UrlCut/Cutter"
	UrlCut_Redirect_FullMethodName        = "/urlcut.UrlCut/Redirect"
	UrlCut_Ping_FullMethodName            = "/urlcut.UrlCut/Ping"
	UrlCut_CutterJsonBatch_FullMethodName = "/urlcut.UrlCut/CutterJsonBatch"
	UrlCut_UserUrls_FullMethodName        = "/urlcut.UrlCut/UserUrls"
	UrlCut_DeleteUsersUrls_FullMethodName = "/urlcut.UrlCut/DeleteUsersUrls"
	UrlCut_Stats_FullMethodName           = "/urlcut.UrlCut/Stats"
)

// UrlCutClient is the client API for UrlCut service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UrlCutClient interface {
	CutterJson(ctx context.Context, in *CutterRequest, opts ...grpc.CallOption) (*CutterResponse, error)
	Cutter(ctx context.Context, in *CutterRequest, opts ...grpc.CallOption) (*CutterResponse, error)
	Redirect(ctx context.Context, in *RedirectRequest, opts ...grpc.CallOption) (*RedirectResponse, error)
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CutterJsonBatch(ctx context.Context, in *CutterJsonBatchRequest, opts ...grpc.CallOption) (*CutterJsonBatchResponse, error)
	UserUrls(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserUrlsResponse, error)
	DeleteUsersUrls(ctx context.Context, in *DeleteUserUrlsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Stats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatsResponse, error)
}

type urlCutClient struct {
	cc grpc.ClientConnInterface
}

func NewUrlCutClient(cc grpc.ClientConnInterface) UrlCutClient {
	return &urlCutClient{cc}
}

func (c *urlCutClient) CutterJson(ctx context.Context, in *CutterRequest, opts ...grpc.CallOption) (*CutterResponse, error) {
	out := new(CutterResponse)
	err := c.cc.Invoke(ctx, UrlCut_CutterJson_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) Cutter(ctx context.Context, in *CutterRequest, opts ...grpc.CallOption) (*CutterResponse, error) {
	out := new(CutterResponse)
	err := c.cc.Invoke(ctx, UrlCut_Cutter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) Redirect(ctx context.Context, in *RedirectRequest, opts ...grpc.CallOption) (*RedirectResponse, error) {
	out := new(RedirectResponse)
	err := c.cc.Invoke(ctx, UrlCut_Redirect_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UrlCut_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) CutterJsonBatch(ctx context.Context, in *CutterJsonBatchRequest, opts ...grpc.CallOption) (*CutterJsonBatchResponse, error) {
	out := new(CutterJsonBatchResponse)
	err := c.cc.Invoke(ctx, UrlCut_CutterJsonBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) UserUrls(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserUrlsResponse, error) {
	out := new(UserUrlsResponse)
	err := c.cc.Invoke(ctx, UrlCut_UserUrls_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) DeleteUsersUrls(ctx context.Context, in *DeleteUserUrlsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UrlCut_DeleteUsersUrls_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *urlCutClient) Stats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatsResponse, error) {
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, UrlCut_Stats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UrlCutServer is the server API for UrlCut service.
// All implementations must embed UnimplementedUrlCutServer
// for forward compatibility
type UrlCutServer interface {
	CutterJson(context.Context, *CutterRequest) (*CutterResponse, error)
	Cutter(context.Context, *CutterRequest) (*CutterResponse, error)
	Redirect(context.Context, *RedirectRequest) (*RedirectResponse, error)
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	CutterJsonBatch(context.Context, *CutterJsonBatchRequest) (*CutterJsonBatchResponse, error)
	UserUrls(context.Context, *emptypb.Empty) (*UserUrlsResponse, error)
	DeleteUsersUrls(context.Context, *DeleteUserUrlsRequest) (*emptypb.Empty, error)
	Stats(context.Context, *emptypb.Empty) (*StatsResponse, error)
	mustEmbedUnimplementedUrlCutServer()
}

// UnimplementedUrlCutServer must be embedded to have forward compatible implementations.
type UnimplementedUrlCutServer struct {
}

func (UnimplementedUrlCutServer) CutterJson(context.Context, *CutterRequest) (*CutterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CutterJson not implemented")
}
func (UnimplementedUrlCutServer) Cutter(context.Context, *CutterRequest) (*CutterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Cutter not implemented")
}
func (UnimplementedUrlCutServer) Redirect(context.Context, *RedirectRequest) (*RedirectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Redirect not implemented")
}
func (UnimplementedUrlCutServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedUrlCutServer) CutterJsonBatch(context.Context, *CutterJsonBatchRequest) (*CutterJsonBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CutterJsonBatch not implemented")
}
func (UnimplementedUrlCutServer) UserUrls(context.Context, *emptypb.Empty) (*UserUrlsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserUrls not implemented")
}
func (UnimplementedUrlCutServer) DeleteUsersUrls(context.Context, *DeleteUserUrlsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUsersUrls not implemented")
}
func (UnimplementedUrlCutServer) Stats(context.Context, *emptypb.Empty) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
func (UnimplementedUrlCutServer) mustEmbedUnimplementedUrlCutServer() {}

// UnsafeUrlCutServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UrlCutServer will
// result in compilation errors.
type UnsafeUrlCutServer interface {
	mustEmbedUnimplementedUrlCutServer()
}

func RegisterUrlCutServer(s grpc.ServiceRegistrar, srv UrlCutServer) {
	s.RegisterService(&UrlCut_ServiceDesc, srv)
}

func _UrlCut_CutterJson_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CutterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).CutterJson(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_CutterJson_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).CutterJson(ctx, req.(*CutterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_Cutter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CutterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).Cutter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_Cutter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).Cutter(ctx, req.(*CutterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_Redirect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RedirectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).Redirect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_Redirect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).Redirect(ctx, req.(*RedirectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_CutterJsonBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CutterJsonBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).CutterJsonBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_CutterJsonBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).CutterJsonBatch(ctx, req.(*CutterJsonBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_UserUrls_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).UserUrls(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_UserUrls_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).UserUrls(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_DeleteUsersUrls_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserUrlsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).DeleteUsersUrls(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_DeleteUsersUrls_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).DeleteUsersUrls(ctx, req.(*DeleteUserUrlsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UrlCut_Stats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UrlCutServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UrlCut_Stats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UrlCutServer).Stats(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// UrlCut_ServiceDesc is the grpc.ServiceDesc for UrlCut service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UrlCut_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "urlcut.UrlCut",
	HandlerType: (*UrlCutServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CutterJson",
			Handler:    _UrlCut_CutterJson_Handler,
		},
		{
			MethodName: "Cutter",
			Handler:    _UrlCut_Cutter_Handler,
		},
		{
			MethodName: "Redirect",
			Handler:    _UrlCut_Redirect_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _UrlCut_Ping_Handler,
		},
		{
			MethodName: "CutterJsonBatch",
			Handler:    _UrlCut_CutterJsonBatch_Handler,
		},
		{
			MethodName: "UserUrls",
			Handler:    _UrlCut_UserUrls_Handler,
		},
		{
			MethodName: "DeleteUsersUrls",
			Handler:    _UrlCut_DeleteUsersUrls_Handler,
		},
		{
			MethodName: "Stats",
			Handler:    _UrlCut_Stats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "urlcut.proto",
}