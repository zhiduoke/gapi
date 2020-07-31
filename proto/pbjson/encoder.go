package pbjson

import (
	"encoding/base64"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
)

type protoValue struct {
	x uint64
	b []byte
}

type fieldValue struct {
	assigned bool
	pv       protoValue
	more     []protoValue
}

type Encoder struct {
	buf []byte
	err error
}

func (e *Encoder) Error() error {
	return e.err
}

func (e *Encoder) Bytes() []byte {
	return e.buf
}

func (e *Encoder) Reset() {
	e.buf = e.buf[:0]
	e.err = nil
}

func (e *Encoder) grow(addition int) {
	newbuf := make([]byte, len(e.buf), cap(e.buf)+addition)
	copy(newbuf, e.buf)
	e.buf = newbuf
}

func (e *Encoder) Reserve(addition int) {
	if cap(e.buf)-len(e.buf) < addition {
		e.grow(addition)
	}
}

func (e *Encoder) WriteByte(b byte) error {
	e.buf = append(e.buf, b)
	return nil
}

func (e *Encoder) WriteByte2(b0 byte, b1 byte) {
	e.buf = append(e.buf, b0, b1)
}

func (e *Encoder) WriteString(s string) {
	e.buf = append(e.buf, s...)
}

func (e *Encoder) WriteBytes(s []byte) {
	e.buf = append(e.buf, s...)
}

func (e *Encoder) consume(pb *proto.Buffer, wireType int) (out protoValue) {
	switch wireType {
	case proto.WireVarint:
		out.x, e.err = pb.DecodeVarint()
	case proto.WireFixed64:
		out.x, e.err = pb.DecodeFixed64()
	case proto.WireBytes:
		out.b, e.err = pb.DecodeRawBytes(false)
	case proto.WireFixed32:
		out.x, e.err = pb.DecodeFixed32()
	case proto.WireStartGroup, proto.WireEndGroup:
		e.err = fmt.Errorf("unexpected wire type: %d", wireType)
	default:
		panic("unreachable")
	}
	return
}

func (e *Encoder) emitPackedValue(kind metadata.TypeKind, b []byte, more bool) bool {
	pb := newProtoBuffer(b)
	var decode func() (uint64, error)
	switch wireTypeOfKind[kind] {
	case proto.WireVarint:
		decode = pb.DecodeVarint
	case proto.WireFixed32:
		decode = pb.DecodeFixed32
	case proto.WireFixed64:
		decode = pb.DecodeFixed64
	default:
		panic("unreachable")
	}
	write := writePrimary[kind]
	for {
		x, err := decode()
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				e.err = err
			}
			break
		}
		if more {
			e.WriteByte(',')
		} else {
			more = true
		}
		write(e, x)
	}
	putProtoBuffer(pb)
	return more
}

func (e *Encoder) encodeRepeatedValue(field *metadata.Field, fv *fieldValue) {
	// https://developers.google.cn/protocol-buffers/docs/encoding#optional
	e.WriteByte('[')
	more := false
	wire := wireTypeOfKind[field.Kind]
	pv := &fv.pv
	i := 0
	for {
		if pv.b != nil && wire != proto.WireBytes {
			// https://developers.google.cn/protocol-buffers/docs/encoding#packed
			more = e.emitPackedValue(field.Kind, pv.b, more)
		} else {
			if more {
				e.WriteByte(',')
			} else {
				more = true
			}
			e.encodeValue(field, pv)
		}
		if e.err != nil {
			return
		}
		if i >= len(fv.more) {
			break
		}
		pv = &fv.more[i]
		i++
	}
	e.WriteByte(']')
}

func (e *Encoder) encodeValue(field *metadata.Field, pv *protoValue) {
	switch field.Kind {
	case metadata.StringKind:
		if field.Options.RawData {
			e.WriteByte('"')
			e.WriteBytes(pv.b)
			e.WriteByte('"')
		} else {
			e.WriteSafeString(pv.b)
		}
	case metadata.BytesKind:
		if field.Options.RawData {
			e.WriteBytes(pv.b)
		} else {
			n := base64.StdEncoding.EncodedLen(len(pv.b))
			// grow
			e.Reserve(n)
			m := len(e.buf)
			d := e.buf[m : m+n]
			base64.StdEncoding.Encode(d, pv.b)
			e.buf = e.buf[:m+n]
		}
	case metadata.MessageKind:
		e.EncodeMessage(field.Message, pv.b)
	default:
		writePrimary[field.Kind](e, pv.x)
	}
}

func (e *Encoder) emitMessage(msg *metadata.Message, values []fieldValue) {
	if !msg.Options.Flat {
		e.WriteByte('{')
	}
	more := false
	for i, field := range msg.Fields {
		fv := &values[i]
		if !fv.assigned && field.Options.OmitEmpty {
			continue
		}
		if more {
			e.WriteByte(',')
		} else {
			more = true
		}
		e.WriteByte('"')
		e.WriteString(field.Name) // direct write field name as json object key
		e.WriteByte2('"', ':')
		if !fv.assigned {
			if field.Repeated && field.Kind != metadata.MapKind {
				e.WriteByte2('[', ']')
			} else {
				e.WriteString(defaultValues[field.Kind])
			}
			continue
		}
		if field.Repeated {
			if field.Kind == metadata.MapKind {
				e.encodeMap(field.Message, fv)
			} else {
				e.encodeRepeatedValue(field, fv)
			}
		} else {
			e.encodeValue(field, &fv.pv)
		}
		if e.err != nil {
			return
		}
	}
	if !msg.Options.Flat {
		e.WriteByte('}')
	}
}

func (e *Encoder) decodeEntry(pb *proto.Buffer, out *[2]fieldValue) {
	for {
		key, err := pb.DecodeVarint()
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				e.err = err
			}
			return
		}
		tag, wire := int(key>>3), int(key&7)
		pv := e.consume(pb, wire)
		if e.err != nil {
			return
		}
		switch tag {
		case 1:
			// key(tag=1) must be a string
			if wire != proto.WireBytes {
				e.err = fmt.Errorf("invalid key wire type: %d", wire)
				return
			}
			out[0] = fieldValue{
				assigned: true,
				pv:       pv,
			}
		case 2:
			out[1] = fieldValue{
				assigned: true,
				pv:       pv,
			}
		default:
			e.err = fmt.Errorf("unexpected tag of entry: %d", tag)
			return
		}
	}
}

func (e *Encoder) encodeMap(msg *metadata.Message, fv *fieldValue) {
	// https://developers.google.cn/protocol-buffers/docs/proto#backwards-compatibility
	// assert filed[0].tag = 1 && filed[1].tag == 2
	// assert filed[0].kind == StringKind
	// assert !filed[1].option.repeated
	valueType := msg.Fields[1]
	e.WriteByte('{')
	i := 0
	more := false
	pv := &fv.pv
	for {
		pb := newProtoBuffer(pv.b)
		var entry [2]fieldValue
		e.decodeEntry(pb, &entry)
		putProtoBuffer(pb)
		if e.err != nil {
			return
		}
		if more {
			e.WriteByte(',')
		} else {
			more = true
		}
		if entry[0].assigned {
			e.WriteSafeString(entry[0].pv.b)
		} else {
			e.WriteString(`""`)
		}
		e.WriteByte(':')
		if entry[1].assigned {
			e.encodeValue(valueType, &entry[1].pv)
		} else {
			e.WriteString(defaultValues[valueType.Kind])
		}
		if e.err != nil {
			return
		}
		if i >= len(fv.more) {
			break
		}
		pv = &fv.more[i]
		i++
	}
	e.WriteByte('}')
}

func (e *Encoder) EncodeMessage(msg *metadata.Message, data []byte) {
	if len(msg.Fields) == 0 {
		if !msg.Options.Flat {
			e.WriteString("{}")
		}
		return
	}
	// https://developers.google.cn/protocol-buffers/docs/encoding#structure
	var (
		fixedValue [32]fieldValue
		values     []fieldValue
	)
	if len(msg.Fields) <= len(fixedValue) {
		values = fixedValue[:]
	} else {
		values = make([]fieldValue, len(msg.Fields))
	}
	pb := newProtoBuffer(data)
	for {
		key, err := pb.DecodeVarint()
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				e.err = err
			}
			break
		}
		tag, wire := int(key>>3), int(key&7)
		idx := msg.TagIndex(tag)
		if idx == -1 {
			// ignore
			e.consume(pb, wire)
			if e.err != nil {
				break
			}
			continue
		}
		field := msg.Fields[idx]
		pv := e.consume(pb, wire)
		if wireTypeOfKind[field.Kind] != wire && (!field.Repeated || wire != proto.WireBytes) {
			e.err = fmt.Errorf("expect wire type %d, got %d", wireTypeOfKind[field.Kind], wire)
			break
		}
		v := &values[idx]
		if !v.assigned || !field.Repeated {
			// always overwrite
			*v = fieldValue{
				assigned: true,
				pv:       pv,
			}
		} else {
			// repeated message/map(repeated MapEntry)
			v.more = append(v.more, pv)
		}
	}
	putProtoBuffer(pb)
	if e.err != nil {
		return
	}
	e.emitMessage(msg, values)
}

func NewEncoder(buf []byte) *Encoder {
	return &Encoder{buf: buf}
}
