// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: tensorflow_federated/proto/v0/executor.proto

package v0

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

// ExecutorGroupClient is the client API for ExecutorGroup service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExecutorGroupClient interface {
	// Returns a
	GetExecutor(ctx context.Context, in *GetExecutorRequest, opts ...grpc.CallOption) (*GetExecutorResponse, error)
	// Creates a value in the executor and returns a reference to it that can be
	// supplied as an argument to other methods.
	CreateValue(ctx context.Context, in *CreateValueRequest, opts ...grpc.CallOption) (*CreateValueResponse, error)
	// Creates a call in the executor and returns a reference to the result.
	CreateCall(ctx context.Context, in *CreateCallRequest, opts ...grpc.CallOption) (*CreateCallResponse, error)
	// Creates a struct of values in the executor and returns a reference to it.
	CreateStruct(ctx context.Context, in *CreateStructRequest, opts ...grpc.CallOption) (*CreateStructResponse, error)
	// Creates a selection from an executor value and returns a reference to it.
	CreateSelection(ctx context.Context, in *CreateSelectionRequest, opts ...grpc.CallOption) (*CreateSelectionResponse, error)
	// Causes a value in the executor to get computed, and sends back the result.
	// WARNING: Unlike all other methods in this API, this may be a long-running
	// call (it will block until the value becomes available).
	Compute(ctx context.Context, in *ComputeRequest, opts ...grpc.CallOption) (*ComputeResponse, error)
	// Causes one or more values in the executor to get disposed of (no longer
	// available for future calls).
	Dispose(ctx context.Context, in *DisposeRequest, opts ...grpc.CallOption) (*DisposeResponse, error)
	// Causes an executor to be disposed of (no longer available for future
	// calls).
	DisposeExecutor(ctx context.Context, in *DisposeExecutorRequest, opts ...grpc.CallOption) (*DisposeExecutorResponse, error)
}

type executorGroupClient struct {
	cc grpc.ClientConnInterface
}

func NewExecutorGroupClient(cc grpc.ClientConnInterface) ExecutorGroupClient {
	return &executorGroupClient{cc}
}

func (c *executorGroupClient) GetExecutor(ctx context.Context, in *GetExecutorRequest, opts ...grpc.CallOption) (*GetExecutorResponse, error) {
	out := new(GetExecutorResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/GetExecutor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) CreateValue(ctx context.Context, in *CreateValueRequest, opts ...grpc.CallOption) (*CreateValueResponse, error) {
	out := new(CreateValueResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/CreateValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) CreateCall(ctx context.Context, in *CreateCallRequest, opts ...grpc.CallOption) (*CreateCallResponse, error) {
	out := new(CreateCallResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/CreateCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) CreateStruct(ctx context.Context, in *CreateStructRequest, opts ...grpc.CallOption) (*CreateStructResponse, error) {
	out := new(CreateStructResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/CreateStruct", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) CreateSelection(ctx context.Context, in *CreateSelectionRequest, opts ...grpc.CallOption) (*CreateSelectionResponse, error) {
	out := new(CreateSelectionResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/CreateSelection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) Compute(ctx context.Context, in *ComputeRequest, opts ...grpc.CallOption) (*ComputeResponse, error) {
	out := new(ComputeResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/Compute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) Dispose(ctx context.Context, in *DisposeRequest, opts ...grpc.CallOption) (*DisposeResponse, error) {
	out := new(DisposeResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/Dispose", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorGroupClient) DisposeExecutor(ctx context.Context, in *DisposeExecutorRequest, opts ...grpc.CallOption) (*DisposeExecutorResponse, error) {
	out := new(DisposeExecutorResponse)
	err := c.cc.Invoke(ctx, "/tensorflow_federated.v0.ExecutorGroup/DisposeExecutor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExecutorGroupServer is the server API for ExecutorGroup service.
// All implementations must embed UnimplementedExecutorGroupServer
// for forward compatibility
type ExecutorGroupServer interface {
	// Returns a
	GetExecutor(context.Context, *GetExecutorRequest) (*GetExecutorResponse, error)
	// Creates a value in the executor and returns a reference to it that can be
	// supplied as an argument to other methods.
	CreateValue(context.Context, *CreateValueRequest) (*CreateValueResponse, error)
	// Creates a call in the executor and returns a reference to the result.
	CreateCall(context.Context, *CreateCallRequest) (*CreateCallResponse, error)
	// Creates a struct of values in the executor and returns a reference to it.
	CreateStruct(context.Context, *CreateStructRequest) (*CreateStructResponse, error)
	// Creates a selection from an executor value and returns a reference to it.
	CreateSelection(context.Context, *CreateSelectionRequest) (*CreateSelectionResponse, error)
	// Causes a value in the executor to get computed, and sends back the result.
	// WARNING: Unlike all other methods in this API, this may be a long-running
	// call (it will block until the value becomes available).
	Compute(context.Context, *ComputeRequest) (*ComputeResponse, error)
	// Causes one or more values in the executor to get disposed of (no longer
	// available for future calls).
	Dispose(context.Context, *DisposeRequest) (*DisposeResponse, error)
	// Causes an executor to be disposed of (no longer available for future
	// calls).
	DisposeExecutor(context.Context, *DisposeExecutorRequest) (*DisposeExecutorResponse, error)
	mustEmbedUnimplementedExecutorGroupServer()
}

// UnimplementedExecutorGroupServer must be embedded to have forward compatible implementations.
type UnimplementedExecutorGroupServer struct {
}

func (UnimplementedExecutorGroupServer) GetExecutor(context.Context, *GetExecutorRequest) (*GetExecutorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExecutor not implemented")
}
func (UnimplementedExecutorGroupServer) CreateValue(context.Context, *CreateValueRequest) (*CreateValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateValue not implemented")
}
func (UnimplementedExecutorGroupServer) CreateCall(context.Context, *CreateCallRequest) (*CreateCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCall not implemented")
}
func (UnimplementedExecutorGroupServer) CreateStruct(context.Context, *CreateStructRequest) (*CreateStructResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateStruct not implemented")
}
func (UnimplementedExecutorGroupServer) CreateSelection(context.Context, *CreateSelectionRequest) (*CreateSelectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSelection not implemented")
}
func (UnimplementedExecutorGroupServer) Compute(context.Context, *ComputeRequest) (*ComputeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Compute not implemented")
}
func (UnimplementedExecutorGroupServer) Dispose(context.Context, *DisposeRequest) (*DisposeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Dispose not implemented")
}
func (UnimplementedExecutorGroupServer) DisposeExecutor(context.Context, *DisposeExecutorRequest) (*DisposeExecutorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisposeExecutor not implemented")
}
func (UnimplementedExecutorGroupServer) mustEmbedUnimplementedExecutorGroupServer() {}

// UnsafeExecutorGroupServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExecutorGroupServer will
// result in compilation errors.
type UnsafeExecutorGroupServer interface {
	mustEmbedUnimplementedExecutorGroupServer()
}

func RegisterExecutorGroupServer(s grpc.ServiceRegistrar, srv ExecutorGroupServer) {
	s.RegisterService(&ExecutorGroup_ServiceDesc, srv)
}

func _ExecutorGroup_GetExecutor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetExecutorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).GetExecutor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/GetExecutor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).GetExecutor(ctx, req.(*GetExecutorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_CreateValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).CreateValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/CreateValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).CreateValue(ctx, req.(*CreateValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_CreateCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).CreateCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/CreateCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).CreateCall(ctx, req.(*CreateCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_CreateStruct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateStructRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).CreateStruct(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/CreateStruct",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).CreateStruct(ctx, req.(*CreateStructRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_CreateSelection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSelectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).CreateSelection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/CreateSelection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).CreateSelection(ctx, req.(*CreateSelectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_Compute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ComputeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).Compute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/Compute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).Compute(ctx, req.(*ComputeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_Dispose_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisposeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).Dispose(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/Dispose",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).Dispose(ctx, req.(*DisposeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorGroup_DisposeExecutor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisposeExecutorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorGroupServer).DisposeExecutor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow_federated.v0.ExecutorGroup/DisposeExecutor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorGroupServer).DisposeExecutor(ctx, req.(*DisposeExecutorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExecutorGroup_ServiceDesc is the grpc.ServiceDesc for ExecutorGroup service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ExecutorGroup_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tensorflow_federated.v0.ExecutorGroup",
	HandlerType: (*ExecutorGroupServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetExecutor",
			Handler:    _ExecutorGroup_GetExecutor_Handler,
		},
		{
			MethodName: "CreateValue",
			Handler:    _ExecutorGroup_CreateValue_Handler,
		},
		{
			MethodName: "CreateCall",
			Handler:    _ExecutorGroup_CreateCall_Handler,
		},
		{
			MethodName: "CreateStruct",
			Handler:    _ExecutorGroup_CreateStruct_Handler,
		},
		{
			MethodName: "CreateSelection",
			Handler:    _ExecutorGroup_CreateSelection_Handler,
		},
		{
			MethodName: "Compute",
			Handler:    _ExecutorGroup_Compute_Handler,
		},
		{
			MethodName: "Dispose",
			Handler:    _ExecutorGroup_Dispose_Handler,
		},
		{
			MethodName: "DisposeExecutor",
			Handler:    _ExecutorGroup_DisposeExecutor_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tensorflow_federated/proto/v0/executor.proto",
}
