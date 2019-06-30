// Code generated by protoc-gen-go. DO NOT EDIT.
// source: controller.proto

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

type NewClientParam struct {
	Client               string   `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	Status               string   `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NewClientParam) Reset()         { *m = NewClientParam{} }
func (m *NewClientParam) String() string { return proto.CompactTextString(m) }
func (*NewClientParam) ProtoMessage()    {}
func (*NewClientParam) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed7f10298fa1d90f, []int{0}
}

func (m *NewClientParam) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NewClientParam.Unmarshal(m, b)
}
func (m *NewClientParam) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NewClientParam.Marshal(b, m, deterministic)
}
func (m *NewClientParam) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NewClientParam.Merge(m, src)
}
func (m *NewClientParam) XXX_Size() int {
	return xxx_messageInfo_NewClientParam.Size(m)
}
func (m *NewClientParam) XXX_DiscardUnknown() {
	xxx_messageInfo_NewClientParam.DiscardUnknown(m)
}

var xxx_messageInfo_NewClientParam proto.InternalMessageInfo

func (m *NewClientParam) GetClient() string {
	if m != nil {
		return m.Client
	}
	return ""
}

func (m *NewClientParam) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func init() {
	proto.RegisterType((*NewClientParam)(nil), "info_age.bremote.NewClientParam")
}

func init() { proto.RegisterFile("controller.proto", fileDescriptor_ed7f10298fa1d90f) }

var fileDescriptor_ed7f10298fa1d90f = []byte{
	// 247 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0x4f, 0x4b, 0x03, 0x31,
	0x10, 0xc5, 0xbb, 0x2a, 0x85, 0x8e, 0x22, 0x35, 0x87, 0xb2, 0xac, 0x07, 0xcb, 0x1e, 0xc4, 0x8b,
	0x29, 0xe8, 0xd5, 0x83, 0xb4, 0x88, 0x37, 0x29, 0xed, 0xcd, 0x8b, 0x64, 0x97, 0x69, 0x08, 0x6c,
	0x32, 0x21, 0x3b, 0xad, 0xf4, 0xcb, 0xfa, 0x59, 0xa4, 0x49, 0xad, 0xff, 0xd8, 0xe3, 0x9b, 0x99,
	0xf7, 0xf8, 0x3d, 0x06, 0x86, 0x35, 0x39, 0x0e, 0xd4, 0x34, 0x18, 0xa4, 0x0f, 0xc4, 0x24, 0x86,
	0xc6, 0xad, 0xe8, 0x4d, 0x69, 0x94, 0x55, 0x40, 0x4b, 0x8c, 0xc5, 0xa5, 0x26, 0xd2, 0x0d, 0x4e,
	0xe2, 0xbe, 0x5a, 0xaf, 0x26, 0x68, 0x3d, 0x6f, 0xd3, 0x79, 0x71, 0xca, 0x5b, 0x8f, 0x6d, 0x12,
	0xe5, 0x23, 0x9c, 0xbf, 0xe0, 0xfb, 0xac, 0x31, 0xe8, 0x78, 0xae, 0x82, 0xb2, 0x62, 0x04, 0xfd,
	0x3a, 0xca, 0x3c, 0x1b, 0x67, 0x37, 0x83, 0xc5, 0x5e, 0xed, 0xe6, 0x2d, 0x2b, 0x5e, 0xb7, 0xf9,
	0x51, 0x9a, 0x27, 0x75, 0xf7, 0x91, 0xc1, 0xc5, 0xec, 0x80, 0xb4, 0xc4, 0xb0, 0x31, 0x35, 0x8a,
	0x07, 0x38, 0x99, 0x1b, 0xa7, 0x45, 0x2e, 0xff, 0xc2, 0xc9, 0x25, 0x07, 0xe3, 0x74, 0xd1, 0xb9,
	0x29, 0x7b, 0xe2, 0x19, 0x06, 0x07, 0x2a, 0x31, 0xfe, 0x7f, 0xf8, 0x1b, 0xb9, 0x18, 0xc9, 0xd4,
	0x57, 0x7e, 0xf5, 0x95, 0x4f, 0xbb, 0xbe, 0x65, 0x4f, 0x4c, 0xe1, 0x6c, 0x81, 0x96, 0x36, 0xb8,
	0xcf, 0xea, 0xc6, 0xe9, 0xcc, 0x98, 0x5e, 0xc3, 0x95, 0x43, 0x8e, 0xc6, 0xdb, 0x9f, 0xc6, 0xef,
	0x3f, 0xbc, 0x1e, 0x2b, 0x6f, 0xaa, 0x7e, 0x74, 0xde, 0x7f, 0x06, 0x00, 0x00, 0xff, 0xff, 0x32,
	0x53, 0x25, 0xad, 0xa1, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ControllerServiceClient is the client API for ControllerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ControllerServiceClient interface {
	Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
	NewClient(ctx context.Context, in *NewClientParam, opts ...grpc.CallOption) (*empty.Empty, error)
	RemoveClient(ctx context.Context, in *String, opts ...grpc.CallOption) (*empty.Empty, error)
}

type controllerServiceClient struct {
	cc *grpc.ClientConn
}

func NewControllerServiceClient(cc *grpc.ClientConn) ControllerServiceClient {
	return &controllerServiceClient{cc}
}

func (c *controllerServiceClient) Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error) {
	out := new(String)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ControllerService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) NewClient(ctx context.Context, in *NewClientParam, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ControllerService/NewClient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) RemoveClient(ctx context.Context, in *String, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ControllerService/RemoveClient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ControllerServiceServer is the server API for ControllerService service.
type ControllerServiceServer interface {
	Ping(context.Context, *String) (*String, error)
	NewClient(context.Context, *NewClientParam) (*empty.Empty, error)
	RemoveClient(context.Context, *String) (*empty.Empty, error)
}

// UnimplementedControllerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedControllerServiceServer struct {
}

func (*UnimplementedControllerServiceServer) Ping(ctx context.Context, req *String) (*String, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedControllerServiceServer) NewClient(ctx context.Context, req *NewClientParam) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewClient not implemented")
}
func (*UnimplementedControllerServiceServer) RemoveClient(ctx context.Context, req *String) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveClient not implemented")
}

func RegisterControllerServiceServer(s *grpc.Server, srv ControllerServiceServer) {
	s.RegisterService(&_ControllerService_serviceDesc, srv)
}

func _ControllerService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ControllerService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).Ping(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_NewClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewClientParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).NewClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ControllerService/NewClient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).NewClient(ctx, req.(*NewClientParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_RemoveClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).RemoveClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ControllerService/RemoveClient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).RemoveClient(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

var _ControllerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "info_age.bremote.ControllerService",
	HandlerType: (*ControllerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ControllerService_Ping_Handler,
		},
		{
			MethodName: "NewClient",
			Handler:    _ControllerService_NewClient_Handler,
		},
		{
			MethodName: "RemoveClient",
			Handler:    _ControllerService_RemoveClient_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "controller.proto",
}
