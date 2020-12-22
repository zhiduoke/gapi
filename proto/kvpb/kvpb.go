package kvpb

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
	annotation "github.com/zhiduoke/gapi/proto"
	"google.golang.org/protobuf/encoding/protowire"
	"math"
	"strconv"
	"sync"
	"unsafe"
)

type KV interface {
	GetForm(key string) (string, bool)
	GetContext(key string) (string, bool)
	GetHeader(key string) (string, bool)
	GetQuery(key string) (string, bool)
	GetParams(key string) (string, bool)
}

type kvGetter func(key string) (string, bool)

type Encoder struct {
	buf *proto.Buffer
	err error
}

var encoderPool sync.Pool

func newEncoder() *Encoder {
	v := encoderPool.Get()
	if v == nil {
		enc := new(Encoder)
		return enc
	}
	return v.(*Encoder)
}

func putEncoder(enc *Encoder) {
	encoderPool.Put(enc)
}

func Encode(msg *metadata.Message, kv KV) ([]byte, error) {
	enc := newEncoder()
	defer putEncoder(enc)
	enc.reset()
	enc.parseKV(msg, kv)
	if enc.err != nil {
		return nil, enc.err
	}
	buf := append([]byte(nil), enc.buf.Bytes()...)
	return buf, nil
}

func (e *Encoder) reset() {
	if e.buf == nil {
		e.buf = proto.NewBuffer(nil)
	}
	e.err = nil
	e.buf.Reset()
}

func (e *Encoder) parseKV(msg *metadata.Message, kv KV) {
	for _, field := range msg.Fields {
		var getters = [...]kvGetter{
			annotation.FIELD_BIND_FROM_DEFAULT: kv.GetForm,
			annotation.FIELD_BIND_FROM_CONTEXT: kv.GetContext,
			annotation.FIELD_BIND_FROM_QUERY:   kv.GetQuery,
			annotation.FIELD_BIND_FROM_HEADER:  kv.GetHeader,
			annotation.FIELD_BIND_FROM_PARAMS:  kv.GetParams,
		}
		fn := getters[annotation.FIELD_BIND(field.Options.Bind)]

		fv, ok := fn(field.Name)
		switch {
		case !ok && !field.Options.Validate:
			continue
		case !ok:
			e.err = fmt.Errorf("must provide param: %s", msg.Name)
			return
		case fieldCanSkip(field):
			continue
		case isNumeric(field.Kind):
			e.transNumber(fv, field)
		case field.Kind == metadata.StringKind ||
			field.Kind == metadata.BytesKind:
			e.transString(fv, field)
		default:
			e.err = fmt.Errorf("invalid kind: %d", field.Kind)
			return
		}
		if e.err != nil && !field.Options.Validate {
			e.err = nil
			continue
		}
		if e.err != nil {
			return
		}
	}
}

func (e *Encoder) transNumber(value string, field *metadata.Field) {
	wire := protowire.VarintType
	var (
		pv  uint64
		err error
	)
	switch field.Kind {
	case metadata.DoubleKind:
		wire = protowire.Fixed64Type
		var fv float64
		fv, err = strconv.ParseFloat(value, 64)
		pv = math.Float64bits(fv)
	case metadata.FloatKind:
		wire = protowire.Fixed32Type
		var fv float64
		fv, err = strconv.ParseFloat(value, 32)
		pv = uint64(math.Float32bits(float32(fv)))
	case metadata.Int32Kind:
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 32)
		pv = uint64(fv)
	case metadata.Int64Kind:
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 64)
		pv = uint64(fv)
	case metadata.Sint32Kind:
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 32)
		pv = protowire.EncodeZigZag(fv)
	case metadata.Sint64Kind:
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 64)
		pv = protowire.EncodeZigZag(fv)
	case metadata.Uint32Kind:
		pv, err = strconv.ParseUint(value, 10, 32)
	case metadata.Uint64Kind:
		pv, err = strconv.ParseUint(value, 10, 64)
	case metadata.Fixed32Kind:
		wire = protowire.Fixed32Type
		pv, err = strconv.ParseUint(value, 10, 32)
	case metadata.Sfixed32Kind:
		wire = protowire.Fixed32Type
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 32)
		pv = uint64(fv)
	case metadata.Fixed64Kind:
		wire = protowire.Fixed64Type
		pv, err = strconv.ParseUint(value, 10, 64)
	case metadata.Sfixed64Kind:
		wire = protowire.Fixed64Type
		var fv int64
		fv, err = strconv.ParseInt(value, 10, 64)
		pv = uint64(fv)
	case metadata.BoolKind:
		var v bool
		v, err = strconv.ParseBool(value)
		if v {
			pv = 1
			break
		}
	default:
		err = fmt.Errorf("invalid kind: %d", field.Kind)
		return
	}
	if err != nil {
		return
	}
	e.encodeKey(field.Tag, wire)
	e.encodeWire(wire, pv)
}

func (e *Encoder) transString(value string, field *metadata.Field) {
	e.encodeBytes(field.Tag, *(*[]byte)(unsafe.Pointer(&value)))
}

func (e *Encoder) encodeWire(wire protowire.Type, pv uint64) {
	switch wire {
	case protowire.Fixed32Type:
		e.buf.EncodeFixed32(pv)
	case protowire.Fixed64Type:
		e.buf.EncodeFixed64(pv)
	default:
		e.buf.EncodeVarint(pv)
	}
}

func (e *Encoder) encodeKey(tag int, wire protowire.Type) {
	k := protowire.EncodeTag(protowire.Number(tag), wire)
	e.buf.EncodeVarint(k)
}

func (e *Encoder) encodeBytes(tag int, v []byte) {
	k := protowire.EncodeTag(protowire.Number(tag), protowire.BytesType)
	e.buf.EncodeVarint(k)
	e.buf.EncodeRawBytes(v)
}
