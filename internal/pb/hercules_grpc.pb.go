// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.19.6
// source: hercules.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	HerculesService_Health_FullMethodName                 = "/hercules.v1.HerculesService/Health"
	HerculesService_ListAnalyses_FullMethodName           = "/hercules.v1.HerculesService/ListAnalyses"
	HerculesService_SubmitAnalysis_FullMethodName         = "/hercules.v1.HerculesService/SubmitAnalysis"
	HerculesService_GetAnalysisStatus_FullMethodName      = "/hercules.v1.HerculesService/GetAnalysisStatus"
	HerculesService_StreamAnalysisProgress_FullMethodName = "/hercules.v1.HerculesService/StreamAnalysisProgress"
)

// HerculesServiceClient is the client API for HerculesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Hercules gRPC Service
type HerculesServiceClient interface {
	// Health check endpoint
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error)
	// List available analysis types
	ListAnalyses(ctx context.Context, in *ListAnalysesRequest, opts ...grpc.CallOption) (*ListAnalysesResponse, error)
	// Submit an analysis request
	SubmitAnalysis(ctx context.Context, in *SubmitAnalysisRequest, opts ...grpc.CallOption) (*SubmitAnalysisResponse, error)
	// Check analysis status
	GetAnalysisStatus(ctx context.Context, in *GetAnalysisStatusRequest, opts ...grpc.CallOption) (*GetAnalysisStatusResponse, error)
	// Stream analysis progress (optional)
	StreamAnalysisProgress(ctx context.Context, in *StreamAnalysisProgressRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamAnalysisProgressResponse], error)
}

type herculesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHerculesServiceClient(cc grpc.ClientConnInterface) HerculesServiceClient {
	return &herculesServiceClient{cc}
}

func (c *herculesServiceClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HealthResponse)
	err := c.cc.Invoke(ctx, HerculesService_Health_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *herculesServiceClient) ListAnalyses(ctx context.Context, in *ListAnalysesRequest, opts ...grpc.CallOption) (*ListAnalysesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListAnalysesResponse)
	err := c.cc.Invoke(ctx, HerculesService_ListAnalyses_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *herculesServiceClient) SubmitAnalysis(ctx context.Context, in *SubmitAnalysisRequest, opts ...grpc.CallOption) (*SubmitAnalysisResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SubmitAnalysisResponse)
	err := c.cc.Invoke(ctx, HerculesService_SubmitAnalysis_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *herculesServiceClient) GetAnalysisStatus(ctx context.Context, in *GetAnalysisStatusRequest, opts ...grpc.CallOption) (*GetAnalysisStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAnalysisStatusResponse)
	err := c.cc.Invoke(ctx, HerculesService_GetAnalysisStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *herculesServiceClient) StreamAnalysisProgress(ctx context.Context, in *StreamAnalysisProgressRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamAnalysisProgressResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &HerculesService_ServiceDesc.Streams[0], HerculesService_StreamAnalysisProgress_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamAnalysisProgressRequest, StreamAnalysisProgressResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type HerculesService_StreamAnalysisProgressClient = grpc.ServerStreamingClient[StreamAnalysisProgressResponse]

// HerculesServiceServer is the server API for HerculesService service.
// All implementations must embed UnimplementedHerculesServiceServer
// for forward compatibility.
//
// Hercules gRPC Service
type HerculesServiceServer interface {
	// Health check endpoint
	Health(context.Context, *HealthRequest) (*HealthResponse, error)
	// List available analysis types
	ListAnalyses(context.Context, *ListAnalysesRequest) (*ListAnalysesResponse, error)
	// Submit an analysis request
	SubmitAnalysis(context.Context, *SubmitAnalysisRequest) (*SubmitAnalysisResponse, error)
	// Check analysis status
	GetAnalysisStatus(context.Context, *GetAnalysisStatusRequest) (*GetAnalysisStatusResponse, error)
	// Stream analysis progress (optional)
	StreamAnalysisProgress(*StreamAnalysisProgressRequest, grpc.ServerStreamingServer[StreamAnalysisProgressResponse]) error
	mustEmbedUnimplementedHerculesServiceServer()
}

// UnimplementedHerculesServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedHerculesServiceServer struct{}

func (UnimplementedHerculesServiceServer) Health(context.Context, *HealthRequest) (*HealthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Health not implemented")
}
func (UnimplementedHerculesServiceServer) ListAnalyses(context.Context, *ListAnalysesRequest) (*ListAnalysesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAnalyses not implemented")
}
func (UnimplementedHerculesServiceServer) SubmitAnalysis(context.Context, *SubmitAnalysisRequest) (*SubmitAnalysisResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitAnalysis not implemented")
}
func (UnimplementedHerculesServiceServer) GetAnalysisStatus(context.Context, *GetAnalysisStatusRequest) (*GetAnalysisStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAnalysisStatus not implemented")
}
func (UnimplementedHerculesServiceServer) StreamAnalysisProgress(*StreamAnalysisProgressRequest, grpc.ServerStreamingServer[StreamAnalysisProgressResponse]) error {
	return status.Errorf(codes.Unimplemented, "method StreamAnalysisProgress not implemented")
}
func (UnimplementedHerculesServiceServer) mustEmbedUnimplementedHerculesServiceServer() {}
func (UnimplementedHerculesServiceServer) testEmbeddedByValue()                         {}

// UnsafeHerculesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HerculesServiceServer will
// result in compilation errors.
type UnsafeHerculesServiceServer interface {
	mustEmbedUnimplementedHerculesServiceServer()
}

func RegisterHerculesServiceServer(s grpc.ServiceRegistrar, srv HerculesServiceServer) {
	// If the following call pancis, it indicates UnimplementedHerculesServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&HerculesService_ServiceDesc, srv)
}

func _HerculesService_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HerculesServiceServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HerculesService_Health_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HerculesServiceServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HerculesService_ListAnalyses_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAnalysesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HerculesServiceServer).ListAnalyses(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HerculesService_ListAnalyses_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HerculesServiceServer).ListAnalyses(ctx, req.(*ListAnalysesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HerculesService_SubmitAnalysis_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitAnalysisRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HerculesServiceServer).SubmitAnalysis(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HerculesService_SubmitAnalysis_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HerculesServiceServer).SubmitAnalysis(ctx, req.(*SubmitAnalysisRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HerculesService_GetAnalysisStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAnalysisStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HerculesServiceServer).GetAnalysisStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HerculesService_GetAnalysisStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HerculesServiceServer).GetAnalysisStatus(ctx, req.(*GetAnalysisStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HerculesService_StreamAnalysisProgress_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamAnalysisProgressRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HerculesServiceServer).StreamAnalysisProgress(m, &grpc.GenericServerStream[StreamAnalysisProgressRequest, StreamAnalysisProgressResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type HerculesService_StreamAnalysisProgressServer = grpc.ServerStreamingServer[StreamAnalysisProgressResponse]

// HerculesService_ServiceDesc is the grpc.ServiceDesc for HerculesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HerculesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "hercules.v1.HerculesService",
	HandlerType: (*HerculesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _HerculesService_Health_Handler,
		},
		{
			MethodName: "ListAnalyses",
			Handler:    _HerculesService_ListAnalyses_Handler,
		},
		{
			MethodName: "SubmitAnalysis",
			Handler:    _HerculesService_SubmitAnalysis_Handler,
		},
		{
			MethodName: "GetAnalysisStatus",
			Handler:    _HerculesService_GetAnalysisStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamAnalysisProgress",
			Handler:       _HerculesService_StreamAnalysisProgress_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "hercules.proto",
}
