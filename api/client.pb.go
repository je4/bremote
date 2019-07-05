// Code generated by protoc-gen-go. DO NOT EDIT.
// source: client.proto

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

type BrowserInitFlag struct {
	Name string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	// Types that are valid to be assigned to Value:
	//	*BrowserInitFlag_Strval
	//	*BrowserInitFlag_Bval
	//	*BrowserInitFlag_Nil
	Value                isBrowserInitFlag_Value `protobuf_oneof:"value"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *BrowserInitFlag) Reset()         { *m = BrowserInitFlag{} }
func (m *BrowserInitFlag) String() string { return proto.CompactTextString(m) }
func (*BrowserInitFlag) ProtoMessage()    {}
func (*BrowserInitFlag) Descriptor() ([]byte, []int) {
	return fileDescriptor_014de31d7ac8c57c, []int{0}
}

func (m *BrowserInitFlag) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BrowserInitFlag.Unmarshal(m, b)
}
func (m *BrowserInitFlag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BrowserInitFlag.Marshal(b, m, deterministic)
}
func (m *BrowserInitFlag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BrowserInitFlag.Merge(m, src)
}
func (m *BrowserInitFlag) XXX_Size() int {
	return xxx_messageInfo_BrowserInitFlag.Size(m)
}
func (m *BrowserInitFlag) XXX_DiscardUnknown() {
	xxx_messageInfo_BrowserInitFlag.DiscardUnknown(m)
}

var xxx_messageInfo_BrowserInitFlag proto.InternalMessageInfo

func (m *BrowserInitFlag) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type isBrowserInitFlag_Value interface {
	isBrowserInitFlag_Value()
}

type BrowserInitFlag_Strval struct {
	Strval string `protobuf:"bytes,2,opt,name=strval,proto3,oneof"`
}

type BrowserInitFlag_Bval struct {
	Bval bool `protobuf:"varint,3,opt,name=bval,proto3,oneof"`
}

type BrowserInitFlag_Nil struct {
	Nil bool `protobuf:"varint,4,opt,name=nil,proto3,oneof"`
}

func (*BrowserInitFlag_Strval) isBrowserInitFlag_Value() {}

func (*BrowserInitFlag_Bval) isBrowserInitFlag_Value() {}

func (*BrowserInitFlag_Nil) isBrowserInitFlag_Value() {}

func (m *BrowserInitFlag) GetValue() isBrowserInitFlag_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *BrowserInitFlag) GetStrval() string {
	if x, ok := m.GetValue().(*BrowserInitFlag_Strval); ok {
		return x.Strval
	}
	return ""
}

func (m *BrowserInitFlag) GetBval() bool {
	if x, ok := m.GetValue().(*BrowserInitFlag_Bval); ok {
		return x.Bval
	}
	return false
}

func (m *BrowserInitFlag) GetNil() bool {
	if x, ok := m.GetValue().(*BrowserInitFlag_Nil); ok {
		return x.Nil
	}
	return false
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*BrowserInitFlag) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*BrowserInitFlag_Strval)(nil),
		(*BrowserInitFlag_Bval)(nil),
		(*BrowserInitFlag_Nil)(nil),
	}
}

type NavigateParam struct {
	Url                  string   `protobuf:"bytes,1,opt,name=Url,proto3" json:"Url,omitempty"`
	NextStatus           string   `protobuf:"bytes,2,opt,name=NextStatus,proto3" json:"NextStatus,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NavigateParam) Reset()         { *m = NavigateParam{} }
func (m *NavigateParam) String() string { return proto.CompactTextString(m) }
func (*NavigateParam) ProtoMessage()    {}
func (*NavigateParam) Descriptor() ([]byte, []int) {
	return fileDescriptor_014de31d7ac8c57c, []int{1}
}

func (m *NavigateParam) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NavigateParam.Unmarshal(m, b)
}
func (m *NavigateParam) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NavigateParam.Marshal(b, m, deterministic)
}
func (m *NavigateParam) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NavigateParam.Merge(m, src)
}
func (m *NavigateParam) XXX_Size() int {
	return xxx_messageInfo_NavigateParam.Size(m)
}
func (m *NavigateParam) XXX_DiscardUnknown() {
	xxx_messageInfo_NavigateParam.DiscardUnknown(m)
}

var xxx_messageInfo_NavigateParam proto.InternalMessageInfo

func (m *NavigateParam) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *NavigateParam) GetNextStatus() string {
	if m != nil {
		return m.NextStatus
	}
	return ""
}

type BrowserInitFlags struct {
	Flags                []*BrowserInitFlag `protobuf:"bytes,1,rep,name=flags,proto3" json:"flags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *BrowserInitFlags) Reset()         { *m = BrowserInitFlags{} }
func (m *BrowserInitFlags) String() string { return proto.CompactTextString(m) }
func (*BrowserInitFlags) ProtoMessage()    {}
func (*BrowserInitFlags) Descriptor() ([]byte, []int) {
	return fileDescriptor_014de31d7ac8c57c, []int{2}
}

func (m *BrowserInitFlags) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BrowserInitFlags.Unmarshal(m, b)
}
func (m *BrowserInitFlags) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BrowserInitFlags.Marshal(b, m, deterministic)
}
func (m *BrowserInitFlags) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BrowserInitFlags.Merge(m, src)
}
func (m *BrowserInitFlags) XXX_Size() int {
	return xxx_messageInfo_BrowserInitFlags.Size(m)
}
func (m *BrowserInitFlags) XXX_DiscardUnknown() {
	xxx_messageInfo_BrowserInitFlags.DiscardUnknown(m)
}

var xxx_messageInfo_BrowserInitFlags proto.InternalMessageInfo

func (m *BrowserInitFlags) GetFlags() []*BrowserInitFlag {
	if m != nil {
		return m.Flags
	}
	return nil
}

func init() {
	proto.RegisterType((*BrowserInitFlag)(nil), "info_age.bremote.BrowserInitFlag")
	proto.RegisterType((*NavigateParam)(nil), "info_age.bremote.NavigateParam")
	proto.RegisterType((*BrowserInitFlags)(nil), "info_age.bremote.BrowserInitFlags")
}

func init() { proto.RegisterFile("client.proto", fileDescriptor_014de31d7ac8c57c) }

var fileDescriptor_014de31d7ac8c57c = []byte{
	// 388 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0x5d, 0x8b, 0xd3, 0x50,
	0x10, 0x4d, 0x4c, 0x76, 0xdd, 0xce, 0xee, 0xb2, 0x65, 0x10, 0x09, 0x59, 0xd0, 0x9a, 0xa7, 0xbe,
	0x98, 0x85, 0xf5, 0xc1, 0x17, 0x61, 0xb1, 0xcb, 0xea, 0x8a, 0x52, 0x4a, 0x82, 0x2f, 0xbe, 0xc8,
	0x4d, 0x9d, 0xc6, 0x0b, 0xc9, 0xbd, 0xe1, 0x66, 0x92, 0xda, 0x9f, 0xe4, 0xbf, 0x94, 0x7c, 0x54,
	0x6a, 0x4b, 0xea, 0xdb, 0xcc, 0xdc, 0x39, 0x67, 0x0e, 0xe7, 0x70, 0xe1, 0x62, 0x99, 0x49, 0x52,
	0x1c, 0x16, 0x46, 0xb3, 0xc6, 0xb1, 0x54, 0x2b, 0xfd, 0x5d, 0xa4, 0x14, 0x26, 0x86, 0x72, 0xcd,
	0xe4, 0x5f, 0xa7, 0x5a, 0xa7, 0x19, 0xdd, 0xb4, 0xef, 0x49, 0xb5, 0xba, 0xa1, 0xbc, 0xe0, 0x4d,
	0xb7, 0xee, 0x9f, 0xf3, 0xa6, 0xa0, 0xb2, 0x6b, 0x02, 0x03, 0x57, 0x33, 0xa3, 0xd7, 0x25, 0x99,
	0x4f, 0x4a, 0xf2, 0x87, 0x4c, 0xa4, 0x88, 0xe0, 0xce, 0x45, 0x4e, 0x9e, 0x3d, 0xb1, 0xa7, 0xa3,
	0xa8, 0xad, 0xd1, 0x83, 0xd3, 0x92, 0x4d, 0x2d, 0x32, 0xef, 0x49, 0x33, 0x7d, 0xb4, 0xa2, 0xbe,
	0xc7, 0x67, 0xe0, 0x26, 0xcd, 0xdc, 0x99, 0xd8, 0xd3, 0xb3, 0x47, 0x2b, 0x6a, 0x3b, 0x44, 0x70,
	0x94, 0xcc, 0x3c, 0xb7, 0x1f, 0x36, 0xcd, 0xec, 0x29, 0x9c, 0xd4, 0x22, 0xab, 0x28, 0x78, 0x0f,
	0x97, 0x73, 0x51, 0xcb, 0x54, 0x30, 0x2d, 0x84, 0x11, 0x39, 0x8e, 0xc1, 0xf9, 0x6a, 0xb2, 0xfe,
	0x60, 0x53, 0xe2, 0x0b, 0x80, 0x39, 0xfd, 0xe2, 0x98, 0x05, 0x57, 0x65, 0x77, 0x33, 0xda, 0x99,
	0x04, 0x9f, 0x61, 0xbc, 0x27, 0xbb, 0xc4, 0xb7, 0x70, 0xb2, 0x6a, 0x0a, 0xcf, 0x9e, 0x38, 0xd3,
	0xf3, 0xdb, 0x57, 0xe1, 0xbe, 0x2d, 0xe1, 0x1e, 0x24, 0xea, 0xf6, 0x6f, 0x7f, 0x3b, 0x70, 0x79,
	0xdf, 0x1a, 0x1a, 0x93, 0xa9, 0xe5, 0x92, 0xf0, 0x1d, 0xb8, 0x0b, 0xa9, 0x52, 0xf4, 0x0e, 0x39,
	0x62, 0x36, 0x52, 0xa5, 0xfe, 0xe0, 0x4b, 0x60, 0xe1, 0x17, 0xb8, 0x88, 0x59, 0x18, 0xee, 0xcf,
	0x61, 0xf0, 0x5f, 0x25, 0xa5, 0xff, 0x3c, 0xec, 0x22, 0x0b, 0xb7, 0x91, 0x85, 0x0f, 0x4d, 0x64,
	0x81, 0x85, 0x0f, 0x70, 0xb6, 0x75, 0x0b, 0x5f, 0x1e, 0x32, 0xfd, 0xe3, 0xe4, 0x11, 0x9a, 0x7b,
	0xb8, 0x8a, 0x7f, 0x56, 0xfc, 0x43, 0xaf, 0xd5, 0x56, 0xd7, 0xc0, 0xf2, 0x11, 0x92, 0x3b, 0x18,
	0x7d, 0xa4, 0x3e, 0x83, 0x41, 0xf8, 0x31, 0x6b, 0xee, 0x60, 0x14, 0xff, 0x25, 0x18, 0x76, 0x77,
	0x50, 0xc1, 0x2c, 0x80, 0x6b, 0x45, 0xdc, 0x02, 0x5f, 0xef, 0x02, 0xbb, 0x0f, 0xf1, 0xcd, 0x11,
	0x85, 0x4c, 0x4e, 0x5b, 0xd4, 0x9b, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3d, 0x2a, 0x7d, 0x10,
	0x26, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ClientServiceClient is the client API for ClientService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ClientServiceClient interface {
	Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
	StartBrowser(ctx context.Context, in *BrowserInitFlags, opts ...grpc.CallOption) (*empty.Empty, error)
	Navigate(ctx context.Context, in *NavigateParam, opts ...grpc.CallOption) (*empty.Empty, error)
	ShutdownBrowser(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	GetStatus(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*String, error)
	SetStatus(ctx context.Context, in *String, opts ...grpc.CallOption) (*empty.Empty, error)
}

type clientServiceClient struct {
	cc *grpc.ClientConn
}

func NewClientServiceClient(cc *grpc.ClientConn) ClientServiceClient {
	return &clientServiceClient{cc}
}

func (c *clientServiceClient) Ping(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error) {
	out := new(String)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) StartBrowser(ctx context.Context, in *BrowserInitFlags, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/StartBrowser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) Navigate(ctx context.Context, in *NavigateParam, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/Navigate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) ShutdownBrowser(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/ShutdownBrowser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) GetStatus(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*String, error) {
	out := new(String)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientServiceClient) SetStatus(ctx context.Context, in *String, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/info_age.bremote.ClientService/SetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientServiceServer is the server API for ClientService service.
type ClientServiceServer interface {
	Ping(context.Context, *String) (*String, error)
	StartBrowser(context.Context, *BrowserInitFlags) (*empty.Empty, error)
	Navigate(context.Context, *NavigateParam) (*empty.Empty, error)
	ShutdownBrowser(context.Context, *empty.Empty) (*empty.Empty, error)
	GetStatus(context.Context, *empty.Empty) (*String, error)
	SetStatus(context.Context, *String) (*empty.Empty, error)
}

// UnimplementedClientServiceServer can be embedded to have forward compatible implementations.
type UnimplementedClientServiceServer struct {
}

func (*UnimplementedClientServiceServer) Ping(ctx context.Context, req *String) (*String, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedClientServiceServer) StartBrowser(ctx context.Context, req *BrowserInitFlags) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartBrowser not implemented")
}
func (*UnimplementedClientServiceServer) Navigate(ctx context.Context, req *NavigateParam) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Navigate not implemented")
}
func (*UnimplementedClientServiceServer) ShutdownBrowser(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShutdownBrowser not implemented")
}
func (*UnimplementedClientServiceServer) GetStatus(ctx context.Context, req *empty.Empty) (*String, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (*UnimplementedClientServiceServer) SetStatus(ctx context.Context, req *String) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetStatus not implemented")
}

func RegisterClientServiceServer(s *grpc.Server, srv ClientServiceServer) {
	s.RegisterService(&_ClientService_serviceDesc, srv)
}

func _ClientService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).Ping(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_StartBrowser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BrowserInitFlags)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).StartBrowser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/StartBrowser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).StartBrowser(ctx, req.(*BrowserInitFlags))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_Navigate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NavigateParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).Navigate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/Navigate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).Navigate(ctx, req.(*NavigateParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_ShutdownBrowser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).ShutdownBrowser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/ShutdownBrowser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).ShutdownBrowser(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).GetStatus(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientService_SetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientServiceServer).SetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/info_age.bremote.ClientService/SetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientServiceServer).SetStatus(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

var _ClientService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "info_age.bremote.ClientService",
	HandlerType: (*ClientServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ClientService_Ping_Handler,
		},
		{
			MethodName: "StartBrowser",
			Handler:    _ClientService_StartBrowser_Handler,
		},
		{
			MethodName: "Navigate",
			Handler:    _ClientService_Navigate_Handler,
		},
		{
			MethodName: "ShutdownBrowser",
			Handler:    _ClientService_ShutdownBrowser_Handler,
		},
		{
			MethodName: "GetStatus",
			Handler:    _ClientService_GetStatus_Handler,
		},
		{
			MethodName: "SetStatus",
			Handler:    _ClientService_SetStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "client.proto",
}
