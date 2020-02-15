// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dataproxy.proto

package api

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type StringList struct {
	String_              []string `protobuf:"bytes,1,rep,name=string,proto3" json:"string,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StringList) Reset()         { *m = StringList{} }
func (m *StringList) String() string { return proto.CompactTextString(m) }
func (*StringList) ProtoMessage()    {}
func (*StringList) Descriptor() ([]byte, []int) {
	return fileDescriptor_5398b228535d168d, []int{0}
}

func (m *StringList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StringList.Unmarshal(m, b)
}
func (m *StringList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StringList.Marshal(b, m, deterministic)
}
func (m *StringList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StringList.Merge(m, src)
}
func (m *StringList) XXX_Size() int {
	return xxx_messageInfo_StringList.Size(m)
}
func (m *StringList) XXX_DiscardUnknown() {
	xxx_messageInfo_StringList.DiscardUnknown(m)
}

var xxx_messageInfo_StringList proto.InternalMessageInfo

func (m *StringList) GetString_() []string {
	if m != nil {
		return m.String_
	}
	return nil
}

func init() {
	proto.RegisterType((*StringList)(nil), "info_age.bremote.StringList")
}

func init() { proto.RegisterFile("dataproxy.proto", fileDescriptor_5398b228535d168d) }

var fileDescriptor_5398b228535d168d = []byte{
	// 218 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x49, 0x2c, 0x49,
	0x2c, 0x28, 0xca, 0xaf, 0xa8, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0xc8, 0xcc, 0x4b,
	0xcb, 0x8f, 0x4f, 0x4c, 0x4f, 0xd5, 0x4b, 0x2a, 0x4a, 0xcd, 0xcd, 0x2f, 0x49, 0x95, 0x92, 0x4e,
	0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x07, 0xcb, 0x27, 0x95, 0xa6, 0xe9, 0xa7, 0xe6, 0x16, 0x94,
	0x40, 0x95, 0x4b, 0x71, 0x97, 0x54, 0x16, 0xa4, 0x16, 0x43, 0x38, 0x4a, 0x2a, 0x5c, 0x5c, 0xc1,
	0x25, 0x45, 0x99, 0x79, 0xe9, 0x3e, 0x99, 0xc5, 0x25, 0x42, 0x62, 0x5c, 0x6c, 0xc5, 0x60, 0x9e,
	0x04, 0xa3, 0x02, 0xb3, 0x06, 0x67, 0x10, 0x94, 0x67, 0x34, 0x83, 0x91, 0x4b, 0xc0, 0x05, 0x66,
	0x6b, 0x70, 0x6a, 0x51, 0x59, 0x66, 0x72, 0xaa, 0x90, 0x0d, 0x17, 0x4b, 0x40, 0x66, 0x5e, 0xba,
	0x90, 0x84, 0x1e, 0xba, 0xfd, 0x7a, 0x10, 0x23, 0xa5, 0x70, 0xca, 0x28, 0x31, 0x08, 0xb9, 0x71,
	0xf1, 0x04, 0xa7, 0x96, 0x84, 0x67, 0x64, 0x96, 0xa4, 0xe6, 0x80, 0xac, 0x96, 0xc1, 0xa5, 0x16,
	0xe4, 0x30, 0x29, 0x31, 0x3d, 0x88, 0x8f, 0xf4, 0x60, 0x3e, 0xd2, 0x73, 0x05, 0xf9, 0x48, 0x89,
	0xc1, 0x49, 0x89, 0x4b, 0x3a, 0x2f, 0xb5, 0x04, 0xac, 0x59, 0x17, 0x59, 0x73, 0x72, 0x4e, 0x66,
	0x6a, 0x5e, 0x49, 0x14, 0x73, 0x62, 0x41, 0x66, 0x12, 0x1b, 0x58, 0x97, 0x31, 0x20, 0x00, 0x00,
	0xff, 0xff, 0xbc, 0x7c, 0x63, 0xbc, 0x3a, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// DataproxyServiceClient is the client API for DataproxyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DataproxyServiceClient interface {
	Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
	SetWhitelist(ctx context.Context, in *StringList, opts ...grpc.CallOption) (*empty.Empty, error)
}

type dataproxyServiceClient struct {
	cc *grpc.ClientConn
}

func NewDataproxyServiceClient(cc *grpc.ClientConn) DataproxyServiceClient {
	return &dataproxyServiceClient{cc}
}

func (c *dataproxyServiceClient) Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error) {
	out := new(String)
	err := c.cc.Invoke(ctx, "/info_age.bremote.DataproxyService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataproxyServiceClient) SetWhitelist(ctx context.Context, in *StringList, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.DataproxyService/SetWhitelist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataproxyServiceServer is the server API for DataproxyService service.
type DataproxyServiceServer interface {
	Ping(context.Context, *String) (*String, error)
	SetWhitelist(context.Context, *StringList) (*empty.Empty, error)
}

// UnimplementedDataproxyServiceServer can be embedded to have forward compatible implementations.
type UnimplementedDataproxyServiceServer struct {
}

func (*UnimplementedDataproxyServiceServer) Ping(ctx context.Context, req *String) (*String, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedDataproxyServiceServer) SetWhitelist(ctx context.Context, req *StringList) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetWhitelist not implemented")
}

func RegisterDataproxyServiceServer(s *grpc.Server, srv DataproxyServiceServer) {
	s.RegisterService(&_DataproxyService_serviceDesc, srv)
}

func _DataproxyService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataproxyServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.DataproxyService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataproxyServiceServer).Ping(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataproxyService_SetWhitelist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StringList)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataproxyServiceServer).SetWhitelist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.DataproxyService/SetWhitelist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataproxyServiceServer).SetWhitelist(ctx, req.(*StringList))
	}
	return interceptor(ctx, in, info, handler)
}

var _DataproxyService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "info_age.bremote.DataproxyService",
	HandlerType: (*DataproxyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _DataproxyService_Ping_Handler,
		},
		{
			MethodName: "SetWhitelist",
			Handler:    _DataproxyService_SetWhitelist_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dataproxy.proto",
}
