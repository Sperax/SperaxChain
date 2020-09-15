// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: engine.proto

package bdls_engine

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// MessageType defines supported message types
type EngineMessageType int32

const (
	// Proposal message
	EngineMessageType_Proposal EngineMessageType = 0
	// Consensus message
	EngineMessageType_Consensus EngineMessageType = 1
)

var EngineMessageType_name = map[int32]string{
	0: "Proposal",
	1: "Consensus",
}

var EngineMessageType_value = map[string]int32{
	"Proposal":  0,
	"Consensus": 1,
}

func (x EngineMessageType) String() string {
	return proto.EnumName(EngineMessageType_name, int32(x))
}

func (EngineMessageType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_770b178c3aab763f, []int{0}
}

// Message defines a engine encapsulate message
type EngineMessage struct {
	// Type of this message
	Type EngineMessageType `protobuf:"varint,1,opt,name=Type,proto3,enum=bdls_engine.EngineMessageType" json:"Type,omitempty"`
	// the Message in bytes
	Message              []byte   `protobuf:"bytes,2,opt,name=Message,proto3" json:"Message,omitempty"`
	Nonce                uint32   `protobuf:"varint,3,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EngineMessage) Reset()         { *m = EngineMessage{} }
func (m *EngineMessage) String() string { return proto.CompactTextString(m) }
func (*EngineMessage) ProtoMessage()    {}
func (*EngineMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_770b178c3aab763f, []int{0}
}
func (m *EngineMessage) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EngineMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EngineMessage.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EngineMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EngineMessage.Merge(m, src)
}
func (m *EngineMessage) XXX_Size() int {
	return m.Size()
}
func (m *EngineMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_EngineMessage.DiscardUnknown(m)
}

var xxx_messageInfo_EngineMessage proto.InternalMessageInfo

func (m *EngineMessage) GetType() EngineMessageType {
	if m != nil {
		return m.Type
	}
	return EngineMessageType_Proposal
}

func (m *EngineMessage) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *EngineMessage) GetNonce() uint32 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func init() {
	proto.RegisterEnum("bdls_engine.EngineMessageType", EngineMessageType_name, EngineMessageType_value)
	proto.RegisterType((*EngineMessage)(nil), "bdls_engine.EngineMessage")
}

func init() { proto.RegisterFile("engine.proto", fileDescriptor_770b178c3aab763f) }

var fileDescriptor_770b178c3aab763f = []byte{
	// 176 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0xcd, 0x4b, 0xcf,
	0xcc, 0x4b, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4e, 0x4a, 0xc9, 0x29, 0x8e, 0x87,
	0x08, 0x29, 0x15, 0x73, 0xf1, 0xba, 0x82, 0x59, 0xbe, 0xa9, 0xc5, 0xc5, 0x89, 0xe9, 0xa9, 0x42,
	0x46, 0x5c, 0x2c, 0x21, 0x95, 0x05, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x7c, 0x46, 0x72, 0x7a,
	0x48, 0x8a, 0xf5, 0x50, 0x54, 0x82, 0x54, 0x05, 0x81, 0xd5, 0x0a, 0x49, 0x70, 0xb1, 0x43, 0x05,
	0x25, 0x98, 0x14, 0x18, 0x35, 0x78, 0x82, 0x60, 0x5c, 0x21, 0x11, 0x2e, 0x56, 0xbf, 0xfc, 0xbc,
	0xe4, 0x54, 0x09, 0x66, 0x05, 0x46, 0x0d, 0xde, 0x20, 0x08, 0x47, 0xcb, 0x80, 0x4b, 0x10, 0xc3,
	0x28, 0x21, 0x1e, 0x2e, 0x8e, 0x80, 0xa2, 0xfc, 0x82, 0xfc, 0xe2, 0xc4, 0x1c, 0x01, 0x06, 0x21,
	0x5e, 0x2e, 0x4e, 0xe7, 0xfc, 0xbc, 0xe2, 0xd4, 0xbc, 0xe2, 0xd2, 0x62, 0x01, 0x46, 0x27, 0x9e,
	0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0, 0x48, 0x8e, 0x31, 0x89, 0x0d, 0xec,
	0x11, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3b, 0x1d, 0x82, 0x17, 0xd8, 0x00, 0x00, 0x00,
}

func (m *EngineMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EngineMessage) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EngineMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Nonce != 0 {
		i = encodeVarintEngine(dAtA, i, uint64(m.Nonce))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Message) > 0 {
		i -= len(m.Message)
		copy(dAtA[i:], m.Message)
		i = encodeVarintEngine(dAtA, i, uint64(len(m.Message)))
		i--
		dAtA[i] = 0x12
	}
	if m.Type != 0 {
		i = encodeVarintEngine(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintEngine(dAtA []byte, offset int, v uint64) int {
	offset -= sovEngine(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EngineMessage) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovEngine(uint64(m.Type))
	}
	l = len(m.Message)
	if l > 0 {
		n += 1 + l + sovEngine(uint64(l))
	}
	if m.Nonce != 0 {
		n += 1 + sovEngine(uint64(m.Nonce))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovEngine(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEngine(x uint64) (n int) {
	return sovEngine(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EngineMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEngine
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
			return fmt.Errorf("proto: EngineMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EngineMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEngine
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= EngineMessageType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEngine
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
				return ErrInvalidLengthEngine
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthEngine
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = append(m.Message[:0], dAtA[iNdEx:postIndex]...)
			if m.Message == nil {
				m.Message = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			m.Nonce = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEngine
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nonce |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEngine(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEngine
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthEngine
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
func skipEngine(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEngine
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
					return 0, ErrIntOverflowEngine
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
					return 0, ErrIntOverflowEngine
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
				return 0, ErrInvalidLengthEngine
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEngine
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEngine
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEngine        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEngine          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEngine = fmt.Errorf("proto: unexpected end of group")
)
