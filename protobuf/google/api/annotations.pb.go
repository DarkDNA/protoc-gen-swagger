// Code generated by protoc-gen-go.
// source: google/api/annotations.proto
// DO NOT EDIT!

package google_api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

var E_Http = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MethodOptions)(nil),
	ExtensionType: (*HttpRule)(nil),
	Field:         72295728,
	Name:          "google.api.http",
	Tag:           "bytes,72295728,opt,name=http",
}

func init() {
	proto.RegisterExtension(E_Http)
}

var fileDescriptor1 = []byte{
	// 164 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x92, 0x49, 0xcf, 0xcf, 0x4f,
	0xcf, 0x49, 0xd5, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f, 0xcc, 0xcb, 0xcb, 0x2f, 0x49, 0x2c, 0xc9, 0xcc,
	0xcf, 0x2b, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x82, 0xc8, 0xea, 0x01, 0x65, 0xa5,
	0x44, 0x91, 0x54, 0x66, 0x94, 0x94, 0x14, 0x40, 0x94, 0x48, 0x29, 0x40, 0x85, 0xc1, 0xbc, 0xa4,
	0xd2, 0x34, 0xfd, 0x94, 0xd4, 0xe2, 0xe4, 0xa2, 0xcc, 0x82, 0x92, 0xfc, 0x22, 0x88, 0x0a, 0x2b,
	0x57, 0x2e, 0x16, 0x90, 0x7a, 0x21, 0x39, 0x3d, 0xa8, 0x69, 0x30, 0xa5, 0x7a, 0xbe, 0xa9, 0x25,
	0x19, 0xf9, 0x29, 0xfe, 0x05, 0x60, 0x2b, 0x25, 0x36, 0x9c, 0xda, 0xa3, 0xa4, 0xc0, 0xa8, 0xc1,
	0x6d, 0x24, 0xa2, 0x87, 0xb0, 0x56, 0xcf, 0x03, 0xa8, 0x35, 0xa8, 0x34, 0x27, 0xd5, 0x49, 0x85,
	0x8b, 0x2f, 0x39, 0x3f, 0x17, 0x49, 0xca, 0x49, 0xc0, 0x11, 0xe1, 0xe0, 0x00, 0x90, 0x99, 0x01,
	0x8c, 0x49, 0x6c, 0x60, 0xc3, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x36, 0x43, 0xa8, 0xf7,
	0xd8, 0x00, 0x00, 0x00,
}
