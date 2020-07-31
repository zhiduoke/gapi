package pbjson

import (
	"encoding/base64"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
)

func (e *Encoder) encodeValueFast(field *metadata.Field, x uint64, buf []byte) {
	switch field.Kind {
	case metadata.StringKind:
		if field.Options.RawData {
			e.WriteByte('"')
			e.WriteBytes(buf)
			e.WriteByte('"')
		} else {
			e.WriteSafeString(buf)
		}
	case metadata.BytesKind:
		if field.Options.RawData {
			e.WriteBytes(buf)
		} else {
			n := base64.StdEncoding.EncodedLen(len(buf))
			// grow
			e.Reserve(n)
			m := len(e.buf)
			d := e.buf[m : m+n]
			base64.StdEncoding.Encode(d, buf)
			e.buf = e.buf[:m+n]
		}
	case metadata.MessageKind:
		e.EncodeMessageFast(field.Message, buf)
	default:
		writePrimary[field.Kind](e, x)
	}
}

func (e *Encoder) EncodeMessageFast(msg *metadata.Message, data []byte) {
	if len(msg.Fields) == 0 {
		if !msg.Options.Flat {
			e.WriteString("{}")
		}
		return
	}
	if !msg.Options.Flat {
		e.WriteByte('{')
	}
	curTag := 0
	var (
		curField     *metadata.Field
		fixedEmitted [32]bool
		emitted      []bool
	)
	if len(msg.Fields) <= len(fixedEmitted) {
		emitted = fixedEmitted[:]
	} else {
		emitted = make([]bool, len(msg.Fields))
	}
	closeChar := byte(0)
	more := false
	more1 := false
	pb := newProtoBuffer(data)
	//noinspection GoNilness
	for {
		key, err := pb.DecodeVarint()
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				e.err = err
			}
			break
		}
		var (
			x   uint64
			buf []byte
		)
		tag, wire := int(key>>3), int(key&7)
		switch wire {
		case proto.WireVarint:
			x, e.err = pb.DecodeVarint()
		case proto.WireFixed64:
			x, e.err = pb.DecodeFixed64()
		case proto.WireBytes:
			buf, e.err = pb.DecodeRawBytes(false)
		case proto.WireFixed32:
			x, e.err = pb.DecodeFixed32()
		case proto.WireStartGroup, proto.WireEndGroup:
			e.err = fmt.Errorf("unexpected wire type: %d", wire)
		default:
			panic("unreachable")
		}
		if e.err != nil {
			break
		}
		// reuse field info
		if tag != curTag {
			idx := msg.TagIndex(tag)
			if idx == -1 || emitted[idx] {
				// ignore
				continue
			}
			if closeChar != 0 {
				// close previous field
				e.WriteByte(closeChar)
				closeChar = 0
			}
			curTag = tag
			emitted[idx] = true
			curField = msg.Fields[idx]
			if more {
				e.WriteByte(',')
			} else {
				more = true
			}
			e.WriteByte('"')
			e.WriteString(curField.Name)
			e.WriteByte2('"', ':')
			if curField.Repeated {
				if curField.Kind == metadata.MapKind {
					e.WriteByte('{')
					closeChar = '}'
				} else {
					e.WriteByte('[')
					closeChar = ']'
				}
				more1 = false
			}
		} else if !curField.Repeated {
			continue
		}
		// emit value
		fwire := wireTypeOfKind[curField.Kind]
		if curField.Repeated {
			// array or map
			if wire == proto.WireBytes && fwire != proto.WireBytes {
				// packed
				more1 = e.emitPackedValue(curField.Kind, buf, more1)
				continue
			}
			if more1 {
				e.WriteByte(',')
			} else {
				more1 = true
			}
			if curField.Kind == metadata.MapKind {
				var entry [2]fieldValue
				pb := newProtoBuffer(buf)
				e.decodeEntry(pb, &entry)
				putProtoBuffer(pb)
				if e.err != nil {
					break
				}
				if entry[0].assigned {
					e.WriteSafeString(entry[0].pv.b)
				} else {
					e.WriteString(`""`)
				}
				e.WriteByte(':')
				valueType := curField.Message.Fields[1]
				if entry[1].assigned {
					e.encodeValueFast(valueType, entry[1].pv.x, entry[1].pv.b)
				} else {
					e.WriteString(defaultValues[valueType.Kind])
				}
				if e.err != nil {
					break
				}
				continue
			}
		}
		if fwire != wire {
			e.err = fmt.Errorf("expect wire type %d, got %d", fwire, wire)
			break
		}
		e.encodeValueFast(curField, x, buf)
		if e.err != nil {
			break
		}
	}
	putProtoBuffer(pb)
	if e.err != nil {
		return
	}
	if closeChar != 0 {
		e.WriteByte(closeChar)
	}
	for i, field := range msg.Fields {
		if emitted[i] || field.Options.OmitEmpty {
			continue
		}
		if more {
			e.WriteByte(',')
		} else {
			more = true
		}
		e.WriteByte('"')
		e.WriteString(field.Name)
		e.WriteByte2('"', ':')
		if field.Repeated && field.Kind != metadata.MapKind {
			e.WriteByte2('[', ']')
		} else {
			e.WriteString(defaultValues[field.Kind])
		}
	}
	if !msg.Options.Flat {
		e.WriteByte('}')
	}
}
