// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.17.1
// source: db.proto

package gen

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
	DB_Tx_FullMethodName = "/database.DB/Tx"
)

// DBClient is the client API for DB service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DBClient interface {
	Tx(ctx context.Context, in *Cursor, opts ...grpc.CallOption) (DB_TxClient, error)
}

type dBClient struct {
	cc grpc.ClientConnInterface
}

func NewDBClient(cc grpc.ClientConnInterface) DBClient {
	return &dBClient{cc}
}

func (c *dBClient) Tx(ctx context.Context, in *Cursor, opts ...grpc.CallOption) (DB_TxClient, error) {
	stream, err := c.cc.NewStream(ctx, &DB_ServiceDesc.Streams[0], DB_Tx_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &dBTxClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DB_TxClient interface {
	Recv() (*Pair, error)
	grpc.ClientStream
}

type dBTxClient struct {
	grpc.ClientStream
}

func (x *dBTxClient) Recv() (*Pair, error) {
	m := new(Pair)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DBServer is the server API for DB service.
// All implementations should embed UnimplementedDBServer
// for forward compatibility
type DBServer interface {
	Tx(*Cursor, DB_TxServer) error
}

// UnimplementedDBServer should be embedded to have forward compatible implementations.
type UnimplementedDBServer struct {
}

func (UnimplementedDBServer) Tx(*Cursor, DB_TxServer) error {
	return status.Errorf(codes.Unimplemented, "method Tx not implemented")
}

// UnsafeDBServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DBServer will
// result in compilation errors.
type UnsafeDBServer interface {
	mustEmbedUnimplementedDBServer()
}

func RegisterDBServer(s grpc.ServiceRegistrar, srv DBServer) {
	s.RegisterService(&DB_ServiceDesc, srv)
}

func _DB_Tx_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Cursor)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DBServer).Tx(m, &dBTxServer{stream})
}

type DB_TxServer interface {
	Send(*Pair) error
	grpc.ServerStream
}

type dBTxServer struct {
	grpc.ServerStream
}

func (x *dBTxServer) Send(m *Pair) error {
	return x.ServerStream.SendMsg(m)
}

// DB_ServiceDesc is the grpc.ServiceDesc for DB service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DB_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "database.DB",
	HandlerType: (*DBServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Tx",
			Handler:       _DB_Tx_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "db.proto",
}