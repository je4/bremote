// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package api

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type ProxySessionType int32

const (
	ProxySessionType_Undefined  ProxySessionType = 0
	ProxySessionType_Client     ProxySessionType = 1
	ProxySessionType_Controller ProxySessionType = 2
)

var ProxySessionType_name = map[int32]string{
	0: "Undefined",
	1: "Client",
	2: "Controller",
}

var ProxySessionType_value = map[string]int32{
	"Undefined":  0,
	"Client":     1,
	"Controller": 2,
}

func (x ProxySessionType) String() string {
	return proto.EnumName(ProxySessionType_name, int32(x))
}

func (ProxySessionType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

type String struct {
	Value                string   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *String) Reset()         { *m = String{} }
func (m *String) String() string { return proto.CompactTextString(m) }
func (*String) ProtoMessage()    {}
func (*String) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

func (m *String) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_String.Unmarshal(m, b)
}
func (m *String) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_String.Marshal(b, m, deterministic)
}
func (m *String) XXX_Merge(src proto.Message) {
	xxx_messageInfo_String.Merge(m, src)
}
func (m *String) XXX_Size() int {
	return xxx_messageInfo_String.Size(m)
}
func (m *String) XXX_DiscardUnknown() {
	xxx_messageInfo_String.DiscardUnknown(m)
}

var xxx_messageInfo_String proto.InternalMessageInfo

func (m *String) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Boolean struct {
	Value                bool     `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Boolean) Reset()         { *m = Boolean{} }
func (m *Boolean) String() string { return proto.CompactTextString(m) }
func (*Boolean) ProtoMessage()    {}
func (*Boolean) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{1}
}

func (m *Boolean) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Boolean.Unmarshal(m, b)
}
func (m *Boolean) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Boolean.Marshal(b, m, deterministic)
}
func (m *Boolean) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Boolean.Merge(m, src)
}
func (m *Boolean) XXX_Size() int {
	return xxx_messageInfo_Boolean.Size(m)
}
func (m *Boolean) XXX_DiscardUnknown() {
	xxx_messageInfo_Boolean.DiscardUnknown(m)
}

var xxx_messageInfo_Boolean proto.InternalMessageInfo

func (m *Boolean) GetValue() bool {
	if m != nil {
		return m.Value
	}
	return false
}

type ProxyClient struct {
	Type                 ProxySessionType `protobuf:"varint,1,opt,name=type,proto3,enum=info_age.bremote.ProxySessionType" json:"type,omitempty"`
	Instance             string           `protobuf:"bytes,2,opt,name=instance,proto3" json:"instance,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *ProxyClient) Reset()         { *m = ProxyClient{} }
func (m *ProxyClient) String() string { return proto.CompactTextString(m) }
func (*ProxyClient) ProtoMessage()    {}
func (*ProxyClient) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{2}
}

func (m *ProxyClient) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProxyClient.Unmarshal(m, b)
}
func (m *ProxyClient) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProxyClient.Marshal(b, m, deterministic)
}
func (m *ProxyClient) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProxyClient.Merge(m, src)
}
func (m *ProxyClient) XXX_Size() int {
	return xxx_messageInfo_ProxyClient.Size(m)
}
func (m *ProxyClient) XXX_DiscardUnknown() {
	xxx_messageInfo_ProxyClient.DiscardUnknown(m)
}

var xxx_messageInfo_ProxyClient proto.InternalMessageInfo

func (m *ProxyClient) GetType() ProxySessionType {
	if m != nil {
		return m.Type
	}
	return ProxySessionType_Undefined
}

func (m *ProxyClient) GetInstance() string {
	if m != nil {
		return m.Instance
	}
	return ""
}

func init() {
	proto.RegisterEnum("info_age.bremote.ProxySessionType", ProxySessionType_name, ProxySessionType_value)
	proto.RegisterType((*String)(nil), "info_age.bremote.String")
	proto.RegisterType((*Boolean)(nil), "info_age.bremote.Boolean")
	proto.RegisterType((*ProxyClient)(nil), "info_age.bremote.ProxyClient")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 226 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x41, 0x4b, 0x03, 0x31,
	0x10, 0x46, 0xdd, 0xaa, 0x6b, 0x3b, 0xc5, 0x12, 0x82, 0x87, 0xa2, 0xa0, 0xb2, 0x27, 0x11, 0xdc,
	0x83, 0x82, 0x37, 0x2f, 0xed, 0x1f, 0x90, 0x56, 0x2f, 0x5e, 0x24, 0xb5, 0x5f, 0x4b, 0x20, 0xce,
	0x84, 0x64, 0x2c, 0xee, 0xbf, 0x17, 0x53, 0x14, 0xed, 0xf1, 0x83, 0x37, 0xbc, 0xc7, 0xd0, 0x50,
	0xbb, 0x88, 0xdc, 0xc6, 0x24, 0x2a, 0xd6, 0x78, 0x5e, 0xc9, 0xab, 0x5b, 0xa3, 0x5d, 0x24, 0xbc,
	0x8b, 0xa2, 0x39, 0xa7, 0x7a, 0xae, 0xc9, 0xf3, 0xda, 0x9e, 0xd0, 0xe1, 0xc6, 0x85, 0x0f, 0x8c,
	0xab, 0xcb, 0xea, 0x6a, 0x30, 0xdb, 0x8e, 0xe6, 0x82, 0x8e, 0x26, 0x22, 0x01, 0x8e, 0xff, 0x03,
	0xfd, 0x1f, 0xc0, 0xd1, 0xf0, 0x31, 0xc9, 0x67, 0x37, 0x0d, 0x1e, 0xac, 0xf6, 0x9e, 0x0e, 0xbe,
	0x85, 0x85, 0x19, 0xdd, 0x36, 0xed, 0xae, 0xb0, 0x2d, 0xf0, 0x1c, 0x39, 0x7b, 0xe1, 0xa7, 0x2e,
	0x62, 0x56, 0x78, 0x7b, 0x4a, 0x7d, 0xcf, 0x59, 0x1d, 0xbf, 0x61, 0xdc, 0x2b, 0x01, 0xbf, 0xfb,
	0xfa, 0x81, 0xcc, 0xee, 0x95, 0x3d, 0xa6, 0xc1, 0x33, 0x2f, 0xb1, 0xf2, 0x8c, 0xa5, 0xd9, 0xb3,
	0x44, 0xf5, 0x36, 0xc0, 0x54, 0x76, 0x44, 0x34, 0x15, 0xd6, 0x24, 0x21, 0x20, 0x99, 0xde, 0xa4,
	0xa1, 0x33, 0x86, 0x96, 0x92, 0x9b, 0xbf, 0x25, 0x19, 0x69, 0x83, 0xf4, 0xb2, 0xef, 0xa2, 0x5f,
	0xd4, 0xe5, 0x3f, 0x77, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x44, 0x2a, 0xf5, 0xcf, 0x2e, 0x01,
	0x00, 0x00,
}
