// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: write.proto

package protoWriteV1

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
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

type WriteRequest struct {
	Record               []byte   `protobuf:"bytes,1,opt,name=record,proto3" json:"record,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WriteRequest) Reset()         { *m = WriteRequest{} }
func (m *WriteRequest) String() string { return proto.CompactTextString(m) }
func (*WriteRequest) ProtoMessage()    {}
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_67966b2b12a73214, []int{0}
}
func (m *WriteRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WriteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WriteRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WriteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WriteRequest.Merge(m, src)
}
func (m *WriteRequest) XXX_Size() int {
	return m.Size()
}
func (m *WriteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_WriteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_WriteRequest proto.InternalMessageInfo

func (m *WriteRequest) GetRecord() []byte {
	if m != nil {
		return m.Record
	}
	return nil
}

type WriteResponse struct {
	Err                  string   `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WriteResponse) Reset()         { *m = WriteResponse{} }
func (m *WriteResponse) String() string { return proto.CompactTextString(m) }
func (*WriteResponse) ProtoMessage()    {}
func (*WriteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_67966b2b12a73214, []int{1}
}
func (m *WriteResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WriteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WriteResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WriteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WriteResponse.Merge(m, src)
}
func (m *WriteResponse) XXX_Size() int {
	return m.Size()
}
func (m *WriteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_WriteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_WriteResponse proto.InternalMessageInfo

func (m *WriteResponse) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

func init() {
	proto.RegisterType((*WriteRequest)(nil), "protoWriteV1.WriteRequest")
	proto.RegisterType((*WriteResponse)(nil), "protoWriteV1.WriteResponse")
}

func init() { proto.RegisterFile("write.proto", fileDescriptor_67966b2b12a73214) }

var fileDescriptor_67966b2b12a73214 = []byte{
	// 163 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2e, 0x2f, 0xca, 0x2c,
	0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x01, 0x53, 0xe1, 0x20, 0x91, 0x30, 0x43,
	0x25, 0x35, 0x2e, 0x1e, 0x30, 0x33, 0x28, 0xb5, 0xb0, 0x34, 0xb5, 0xb8, 0x44, 0x48, 0x8c, 0x8b,
	0xad, 0x28, 0x35, 0x39, 0xbf, 0x28, 0x45, 0x82, 0x51, 0x81, 0x51, 0x83, 0x27, 0x08, 0xca, 0x53,
	0x52, 0xe4, 0xe2, 0x85, 0xaa, 0x2b, 0x2e, 0xc8, 0xcf, 0x2b, 0x4e, 0x15, 0x12, 0xe0, 0x62, 0x4e,
	0x2d, 0x2a, 0x02, 0xab, 0xe2, 0x0c, 0x02, 0x31, 0x8d, 0xc2, 0xa0, 0x46, 0x05, 0xa7, 0x16, 0x95,
	0x65, 0x26, 0xa7, 0x0a, 0xb9, 0x71, 0xb1, 0x82, 0xf9, 0x42, 0x52, 0x7a, 0xc8, 0x56, 0xea, 0x21,
	0xdb, 0x27, 0x25, 0x8d, 0x55, 0x0e, 0x62, 0x87, 0x12, 0x83, 0x06, 0xa3, 0x01, 0xa3, 0x93, 0xc0,
	0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0x38, 0xe3, 0xb1, 0x1c,
	0x43, 0x12, 0x1b, 0x58, 0x8f, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x48, 0xda, 0x57, 0x19, 0xd8,
	0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// WriteServiceClient is the client API for WriteService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type WriteServiceClient interface {
	Write(ctx context.Context, opts ...grpc.CallOption) (WriteService_WriteClient, error)
}

type writeServiceClient struct {
	cc *grpc.ClientConn
}

func NewWriteServiceClient(cc *grpc.ClientConn) WriteServiceClient {
	return &writeServiceClient{cc}
}

func (c *writeServiceClient) Write(ctx context.Context, opts ...grpc.CallOption) (WriteService_WriteClient, error) {
	stream, err := c.cc.NewStream(ctx, &_WriteService_serviceDesc.Streams[0], "/protoWriteV1.WriteService/Write", opts...)
	if err != nil {
		return nil, err
	}
	x := &writeServiceWriteClient{stream}
	return x, nil
}

type WriteService_WriteClient interface {
	Send(*WriteRequest) error
	Recv() (*WriteResponse, error)
	grpc.ClientStream
}

type writeServiceWriteClient struct {
	grpc.ClientStream
}

func (x *writeServiceWriteClient) Send(m *WriteRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *writeServiceWriteClient) Recv() (*WriteResponse, error) {
	m := new(WriteResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// WriteServiceServer is the server API for WriteService service.
type WriteServiceServer interface {
	Write(WriteService_WriteServer) error
}

// UnimplementedWriteServiceServer can be embedded to have forward compatible implementations.
type UnimplementedWriteServiceServer struct {
}

func (*UnimplementedWriteServiceServer) Write(srv WriteService_WriteServer) error {
	return status.Errorf(codes.Unimplemented, "method Write not implemented")
}

func RegisterWriteServiceServer(s *grpc.Server, srv WriteServiceServer) {
	s.RegisterService(&_WriteService_serviceDesc, srv)
}

func _WriteService_Write_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(WriteServiceServer).Write(&writeServiceWriteServer{stream})
}

type WriteService_WriteServer interface {
	Send(*WriteResponse) error
	Recv() (*WriteRequest, error)
	grpc.ServerStream
}

type writeServiceWriteServer struct {
	grpc.ServerStream
}

func (x *writeServiceWriteServer) Send(m *WriteResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *writeServiceWriteServer) Recv() (*WriteRequest, error) {
	m := new(WriteRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _WriteService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protoWriteV1.WriteService",
	HandlerType: (*WriteServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Write",
			Handler:       _WriteService_Write_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "write.proto",
}

func (m *WriteRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WriteRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WriteRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Record) > 0 {
		i -= len(m.Record)
		copy(dAtA[i:], m.Record)
		i = encodeVarintWrite(dAtA, i, uint64(len(m.Record)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *WriteResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WriteResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WriteResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Err) > 0 {
		i -= len(m.Err)
		copy(dAtA[i:], m.Err)
		i = encodeVarintWrite(dAtA, i, uint64(len(m.Err)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintWrite(dAtA []byte, offset int, v uint64) int {
	offset -= sovWrite(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *WriteRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Record)
	if l > 0 {
		n += 1 + l + sovWrite(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *WriteResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Err)
	if l > 0 {
		n += 1 + l + sovWrite(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovWrite(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozWrite(x uint64) (n int) {
	return sovWrite(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *WriteRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWrite
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: WriteRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WriteRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Record", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrite
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthWrite
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthWrite
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Record = append(m.Record[:0], dAtA[iNdEx:postIndex]...)
			if m.Record == nil {
				m.Record = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipWrite(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthWrite
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthWrite
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *WriteResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWrite
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: WriteResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WriteResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Err", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrite
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthWrite
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthWrite
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Err = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipWrite(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthWrite
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthWrite
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipWrite(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowWrite
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowWrite
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowWrite
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthWrite
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupWrite
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthWrite
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthWrite        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowWrite          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupWrite = fmt.Errorf("proto: unexpected end of group")
)