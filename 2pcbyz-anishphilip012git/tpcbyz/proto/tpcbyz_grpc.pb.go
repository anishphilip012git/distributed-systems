// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: proto/tpcbyz.proto

package tpcbyz

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
	PBFTClientService_ReplyMessage_FullMethodName = "/tpcbyz.PBFTClientService/ReplyMessage"
)

// PBFTClientServiceClient is the client API for PBFTClientService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Define a separate service for the client to receive replies
type PBFTClientServiceClient interface {
	ReplyMessage(ctx context.Context, in *ClientReplyMessage, opts ...grpc.CallOption) (*StatusResponse, error)
}

type pBFTClientServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPBFTClientServiceClient(cc grpc.ClientConnInterface) PBFTClientServiceClient {
	return &pBFTClientServiceClient{cc}
}

func (c *pBFTClientServiceClient) ReplyMessage(ctx context.Context, in *ClientReplyMessage, opts ...grpc.CallOption) (*StatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, PBFTClientService_ReplyMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PBFTClientServiceServer is the server API for PBFTClientService service.
// All implementations must embed UnimplementedPBFTClientServiceServer
// for forward compatibility.
//
// Define a separate service for the client to receive replies
type PBFTClientServiceServer interface {
	ReplyMessage(context.Context, *ClientReplyMessage) (*StatusResponse, error)
	mustEmbedUnimplementedPBFTClientServiceServer()
}

// UnimplementedPBFTClientServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPBFTClientServiceServer struct{}

func (UnimplementedPBFTClientServiceServer) ReplyMessage(context.Context, *ClientReplyMessage) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplyMessage not implemented")
}
func (UnimplementedPBFTClientServiceServer) mustEmbedUnimplementedPBFTClientServiceServer() {}
func (UnimplementedPBFTClientServiceServer) testEmbeddedByValue()                           {}

// UnsafePBFTClientServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PBFTClientServiceServer will
// result in compilation errors.
type UnsafePBFTClientServiceServer interface {
	mustEmbedUnimplementedPBFTClientServiceServer()
}

func RegisterPBFTClientServiceServer(s grpc.ServiceRegistrar, srv PBFTClientServiceServer) {
	// If the following call pancis, it indicates UnimplementedPBFTClientServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PBFTClientService_ServiceDesc, srv)
}

func _PBFTClientService_ReplyMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientReplyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PBFTClientServiceServer).ReplyMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PBFTClientService_ReplyMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PBFTClientServiceServer).ReplyMessage(ctx, req.(*ClientReplyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// PBFTClientService_ServiceDesc is the grpc.ServiceDesc for PBFTClientService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PBFTClientService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tpcbyz.PBFTClientService",
	HandlerType: (*PBFTClientServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReplyMessage",
			Handler:    _PBFTClientService_ReplyMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/tpcbyz.proto",
}

const (
	TPCShard_PBFTPrePrepare_FullMethodName     = "/tpcbyz.TPCShard/PBFTPrePrepare"
	TPCShard_PBFTPrepare_FullMethodName        = "/tpcbyz.TPCShard/PBFTPrepare"
	TPCShard_PBFTCommit_FullMethodName         = "/tpcbyz.TPCShard/PBFTCommit"
	TPCShard_ClientRequest_FullMethodName      = "/tpcbyz.TPCShard/ClientRequest"
	TPCShard_ClientRequestTPC_FullMethodName   = "/tpcbyz.TPCShard/ClientRequestTPC"
	TPCShard_CheckHealth_FullMethodName        = "/tpcbyz.TPCShard/CheckHealth"
	TPCShard_TPCPrepare_FullMethodName         = "/tpcbyz.TPCShard/TPCPrepare"
	TPCShard_TPCCooridnatorCall_FullMethodName = "/tpcbyz.TPCShard/TPCCooridnatorCall"
	TPCShard_TPCCommit_FullMethodName          = "/tpcbyz.TPCShard/TPCCommit"
	TPCShard_TPCAbort_FullMethodName           = "/tpcbyz.TPCShard/TPCAbort"
	TPCShard_TPCLock_FullMethodName            = "/tpcbyz.TPCShard/TPCLock"
	TPCShard_TPCUnlock_FullMethodName          = "/tpcbyz.TPCShard/TPCUnlock"
)

// TPCShardClient is the client API for TPCShard service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TPCShardClient interface {
	// PBFT RPCs for intra-shard transactions
	PBFTPrePrepare(ctx context.Context, in *PrePrepareMessage, opts ...grpc.CallOption) (*PrepareMessage, error)
	PBFTPrepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*CommitMessage, error)
	PBFTCommit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	// RPCs for handling transactions
	ClientRequest(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*ClientReplyMessage, error)
	ClientRequestTPC(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*ClientReplyMessage, error)
	// Health check for shards
	CheckHealth(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*HealthResponse, error)
	// Two-Phase Commit (2PC) RPCs for cross-shard transactions using PBFT
	TPCPrepare(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error)
	TPCCooridnatorCall(ctx context.Context, in *TPCCertificate, opts ...grpc.CallOption) (*TransactionResponse, error)
	TPCCommit(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error)
	TPCAbort(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error)
	// Locking and unlocking resources for 2PC
	TPCLock(ctx context.Context, in *TPCLockRequest, opts ...grpc.CallOption) (*TPCLockResponse, error)
	TPCUnlock(ctx context.Context, in *TPCLockRequest, opts ...grpc.CallOption) (*TPCLockResponse, error)
}

type tPCShardClient struct {
	cc grpc.ClientConnInterface
}

func NewTPCShardClient(cc grpc.ClientConnInterface) TPCShardClient {
	return &tPCShardClient{cc}
}

func (c *tPCShardClient) PBFTPrePrepare(ctx context.Context, in *PrePrepareMessage, opts ...grpc.CallOption) (*PrepareMessage, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PrepareMessage)
	err := c.cc.Invoke(ctx, TPCShard_PBFTPrePrepare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) PBFTPrepare(ctx context.Context, in *PrepareRequest, opts ...grpc.CallOption) (*CommitMessage, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CommitMessage)
	err := c.cc.Invoke(ctx, TPCShard_PBFTPrepare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) PBFTCommit(ctx context.Context, in *CommitRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, TPCShard_PBFTCommit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) ClientRequest(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*ClientReplyMessage, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ClientReplyMessage)
	err := c.cc.Invoke(ctx, TPCShard_ClientRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) ClientRequestTPC(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*ClientReplyMessage, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ClientReplyMessage)
	err := c.cc.Invoke(ctx, TPCShard_ClientRequestTPC_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) CheckHealth(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*HealthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HealthResponse)
	err := c.cc.Invoke(ctx, TPCShard_CheckHealth_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCPrepare(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCPrepare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCCooridnatorCall(ctx context.Context, in *TPCCertificate, opts ...grpc.CallOption) (*TransactionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCCooridnatorCall_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCCommit(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCCommit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCAbort(ctx context.Context, in *ClientRequestMessage, opts ...grpc.CallOption) (*TransactionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCAbort_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCLock(ctx context.Context, in *TPCLockRequest, opts ...grpc.CallOption) (*TPCLockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TPCLockResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCLock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tPCShardClient) TPCUnlock(ctx context.Context, in *TPCLockRequest, opts ...grpc.CallOption) (*TPCLockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TPCLockResponse)
	err := c.cc.Invoke(ctx, TPCShard_TPCUnlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TPCShardServer is the server API for TPCShard service.
// All implementations must embed UnimplementedTPCShardServer
// for forward compatibility.
type TPCShardServer interface {
	// PBFT RPCs for intra-shard transactions
	PBFTPrePrepare(context.Context, *PrePrepareMessage) (*PrepareMessage, error)
	PBFTPrepare(context.Context, *PrepareRequest) (*CommitMessage, error)
	PBFTCommit(context.Context, *CommitRequest) (*StatusResponse, error)
	// RPCs for handling transactions
	ClientRequest(context.Context, *ClientRequestMessage) (*ClientReplyMessage, error)
	ClientRequestTPC(context.Context, *ClientRequestMessage) (*ClientReplyMessage, error)
	// Health check for shards
	CheckHealth(context.Context, *Empty) (*HealthResponse, error)
	// Two-Phase Commit (2PC) RPCs for cross-shard transactions using PBFT
	TPCPrepare(context.Context, *ClientRequestMessage) (*TransactionResponse, error)
	TPCCooridnatorCall(context.Context, *TPCCertificate) (*TransactionResponse, error)
	TPCCommit(context.Context, *ClientRequestMessage) (*TransactionResponse, error)
	TPCAbort(context.Context, *ClientRequestMessage) (*TransactionResponse, error)
	// Locking and unlocking resources for 2PC
	TPCLock(context.Context, *TPCLockRequest) (*TPCLockResponse, error)
	TPCUnlock(context.Context, *TPCLockRequest) (*TPCLockResponse, error)
	mustEmbedUnimplementedTPCShardServer()
}

// UnimplementedTPCShardServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTPCShardServer struct{}

func (UnimplementedTPCShardServer) PBFTPrePrepare(context.Context, *PrePrepareMessage) (*PrepareMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PBFTPrePrepare not implemented")
}
func (UnimplementedTPCShardServer) PBFTPrepare(context.Context, *PrepareRequest) (*CommitMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PBFTPrepare not implemented")
}
func (UnimplementedTPCShardServer) PBFTCommit(context.Context, *CommitRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PBFTCommit not implemented")
}
func (UnimplementedTPCShardServer) ClientRequest(context.Context, *ClientRequestMessage) (*ClientReplyMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientRequest not implemented")
}
func (UnimplementedTPCShardServer) ClientRequestTPC(context.Context, *ClientRequestMessage) (*ClientReplyMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientRequestTPC not implemented")
}
func (UnimplementedTPCShardServer) CheckHealth(context.Context, *Empty) (*HealthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckHealth not implemented")
}
func (UnimplementedTPCShardServer) TPCPrepare(context.Context, *ClientRequestMessage) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCPrepare not implemented")
}
func (UnimplementedTPCShardServer) TPCCooridnatorCall(context.Context, *TPCCertificate) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCCooridnatorCall not implemented")
}
func (UnimplementedTPCShardServer) TPCCommit(context.Context, *ClientRequestMessage) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCCommit not implemented")
}
func (UnimplementedTPCShardServer) TPCAbort(context.Context, *ClientRequestMessage) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCAbort not implemented")
}
func (UnimplementedTPCShardServer) TPCLock(context.Context, *TPCLockRequest) (*TPCLockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCLock not implemented")
}
func (UnimplementedTPCShardServer) TPCUnlock(context.Context, *TPCLockRequest) (*TPCLockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TPCUnlock not implemented")
}
func (UnimplementedTPCShardServer) mustEmbedUnimplementedTPCShardServer() {}
func (UnimplementedTPCShardServer) testEmbeddedByValue()                  {}

// UnsafeTPCShardServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TPCShardServer will
// result in compilation errors.
type UnsafeTPCShardServer interface {
	mustEmbedUnimplementedTPCShardServer()
}

func RegisterTPCShardServer(s grpc.ServiceRegistrar, srv TPCShardServer) {
	// If the following call pancis, it indicates UnimplementedTPCShardServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TPCShard_ServiceDesc, srv)
}

func _TPCShard_PBFTPrePrepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrePrepareMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).PBFTPrePrepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_PBFTPrePrepare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).PBFTPrePrepare(ctx, req.(*PrePrepareMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_PBFTPrepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrepareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).PBFTPrepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_PBFTPrepare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).PBFTPrepare(ctx, req.(*PrepareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_PBFTCommit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).PBFTCommit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_PBFTCommit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).PBFTCommit(ctx, req.(*CommitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_ClientRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).ClientRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_ClientRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).ClientRequest(ctx, req.(*ClientRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_ClientRequestTPC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).ClientRequestTPC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_ClientRequestTPC_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).ClientRequestTPC(ctx, req.(*ClientRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_CheckHealth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).CheckHealth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_CheckHealth_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).CheckHealth(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCPrepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCPrepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCPrepare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCPrepare(ctx, req.(*ClientRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCCooridnatorCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TPCCertificate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCCooridnatorCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCCooridnatorCall_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCCooridnatorCall(ctx, req.(*TPCCertificate))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCCommit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCCommit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCCommit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCCommit(ctx, req.(*ClientRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCAbort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCAbort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCAbort_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCAbort(ctx, req.(*ClientRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCLock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TPCLockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCLock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCLock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCLock(ctx, req.(*TPCLockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TPCShard_TPCUnlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TPCLockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TPCShardServer).TPCUnlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TPCShard_TPCUnlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TPCShardServer).TPCUnlock(ctx, req.(*TPCLockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TPCShard_ServiceDesc is the grpc.ServiceDesc for TPCShard service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TPCShard_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tpcbyz.TPCShard",
	HandlerType: (*TPCShardServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PBFTPrePrepare",
			Handler:    _TPCShard_PBFTPrePrepare_Handler,
		},
		{
			MethodName: "PBFTPrepare",
			Handler:    _TPCShard_PBFTPrepare_Handler,
		},
		{
			MethodName: "PBFTCommit",
			Handler:    _TPCShard_PBFTCommit_Handler,
		},
		{
			MethodName: "ClientRequest",
			Handler:    _TPCShard_ClientRequest_Handler,
		},
		{
			MethodName: "ClientRequestTPC",
			Handler:    _TPCShard_ClientRequestTPC_Handler,
		},
		{
			MethodName: "CheckHealth",
			Handler:    _TPCShard_CheckHealth_Handler,
		},
		{
			MethodName: "TPCPrepare",
			Handler:    _TPCShard_TPCPrepare_Handler,
		},
		{
			MethodName: "TPCCooridnatorCall",
			Handler:    _TPCShard_TPCCooridnatorCall_Handler,
		},
		{
			MethodName: "TPCCommit",
			Handler:    _TPCShard_TPCCommit_Handler,
		},
		{
			MethodName: "TPCAbort",
			Handler:    _TPCShard_TPCAbort_Handler,
		},
		{
			MethodName: "TPCLock",
			Handler:    _TPCShard_TPCLock_Handler,
		},
		{
			MethodName: "TPCUnlock",
			Handler:    _TPCShard_TPCUnlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/tpcbyz.proto",
}