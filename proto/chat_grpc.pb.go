// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.6.1
// source: chat.proto

package chat_server

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Broadcast_CreateStream_FullMethodName     = "/chat.Broadcast/CreateStream"
	Broadcast_BroadcastMessage_FullMethodName = "/chat.Broadcast/BroadcastMessage"
	Broadcast_CreateUser_FullMethodName       = "/chat.Broadcast/CreateUser"
)

// BroadcastClient is the client API for Broadcast service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BroadcastClient interface {
	CreateStream(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (Broadcast_CreateStreamClient, error)
	BroadcastMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*empty.Empty, error)
	CreateUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*empty.Empty, error)
}

type broadcastClient struct {
	cc grpc.ClientConnInterface
}

func NewBroadcastClient(cc grpc.ClientConnInterface) BroadcastClient {
	return &broadcastClient{cc}
}

func (c *broadcastClient) CreateStream(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (Broadcast_CreateStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Broadcast_ServiceDesc.Streams[0], Broadcast_CreateStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &broadcastCreateStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Broadcast_CreateStreamClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type broadcastCreateStreamClient struct {
	grpc.ClientStream
}

func (x *broadcastCreateStreamClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *broadcastClient) BroadcastMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Broadcast_BroadcastMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *broadcastClient) CreateUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Broadcast_CreateUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BroadcastServer is the server API for Broadcast service.
// All implementations must embed UnimplementedBroadcastServer
// for forward compatibility
type BroadcastServer interface {
	CreateStream(*empty.Empty, Broadcast_CreateStreamServer) error
	BroadcastMessage(context.Context, *Message) (*empty.Empty, error)
	CreateUser(context.Context, *User) (*empty.Empty, error)
	mustEmbedUnimplementedBroadcastServer()
}

// UnimplementedBroadcastServer must be embedded to have forward compatible implementations.
type UnimplementedBroadcastServer struct {
}

func (UnimplementedBroadcastServer) CreateStream(*empty.Empty, Broadcast_CreateStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateStream not implemented")
}
func (UnimplementedBroadcastServer) BroadcastMessage(context.Context, *Message) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadcastMessage not implemented")
}
func (UnimplementedBroadcastServer) CreateUser(context.Context, *User) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (UnimplementedBroadcastServer) mustEmbedUnimplementedBroadcastServer() {}

// UnsafeBroadcastServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BroadcastServer will
// result in compilation errors.
type UnsafeBroadcastServer interface {
	mustEmbedUnimplementedBroadcastServer()
}

func RegisterBroadcastServer(s grpc.ServiceRegistrar, srv BroadcastServer) {
	s.RegisterService(&Broadcast_ServiceDesc, srv)
}

func _Broadcast_CreateStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(empty.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BroadcastServer).CreateStream(m, &broadcastCreateStreamServer{stream})
}

type Broadcast_CreateStreamServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type broadcastCreateStreamServer struct {
	grpc.ServerStream
}

func (x *broadcastCreateStreamServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _Broadcast_BroadcastMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastServer).BroadcastMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Broadcast_BroadcastMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastServer).BroadcastMessage(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _Broadcast_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Broadcast_CreateUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastServer).CreateUser(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

// Broadcast_ServiceDesc is the grpc.ServiceDesc for Broadcast service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Broadcast_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.Broadcast",
	HandlerType: (*BroadcastServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BroadcastMessage",
			Handler:    _Broadcast_BroadcastMessage_Handler,
		},
		{
			MethodName: "CreateUser",
			Handler:    _Broadcast_CreateUser_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CreateStream",
			Handler:       _Broadcast_CreateStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "chat.proto",
}
