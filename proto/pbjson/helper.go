package pbjson

import (
	"math"
	"strconv"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
)

var protoBufferPool sync.Pool

func newProtoBuffer(b []byte) *proto.Buffer {
	var buf *proto.Buffer
	if v := protoBufferPool.Get(); v != nil {
		buf = v.(*proto.Buffer)
	} else {
		buf = new(proto.Buffer)
	}
	buf.SetBuf(b)
	return buf
}

func putProtoBuffer(buf *proto.Buffer) {
	buf.SetBuf(nil)
	protoBufferPool.Put(buf)
}

var wireTypeOfKind = [metadata.MaxTypeKind]int{
	metadata.InvalidType:  -1,
	metadata.Int32Kind:    proto.WireVarint,
	metadata.Uint32Kind:   proto.WireVarint,
	metadata.Int64Kind:    proto.WireVarint,
	metadata.Uint64Kind:   proto.WireVarint,
	metadata.BoolKind:     proto.WireVarint,
	metadata.Sint32Kind:   proto.WireVarint,
	metadata.Sint64Kind:   proto.WireVarint,
	metadata.EnumKind:     proto.WireVarint,
	metadata.Fixed64Kind:  proto.WireFixed64,
	metadata.Sfixed64Kind: proto.WireFixed64,
	metadata.DoubleKind:   proto.WireFixed64,
	metadata.Fixed32Kind:  proto.WireFixed32,
	metadata.Sfixed32Kind: proto.WireFixed32,
	metadata.FloatKind:    proto.WireFixed32,
	metadata.StringKind:   proto.WireBytes,
	metadata.BytesKind:    proto.WireBytes,
	metadata.MessageKind:  proto.WireBytes,
	metadata.MapKind:      proto.WireBytes,
}

var writePrimary = [...]func(e *Encoder, x uint64){
	metadata.Int32Kind:    appendI64,
	metadata.Int64Kind:    appendI64,
	metadata.Sint32Kind:   appendS32,
	metadata.Sint64Kind:   appendS64,
	metadata.EnumKind:     appendI64,
	metadata.Sfixed32Kind: appendS32,
	metadata.Sfixed64Kind: appendS64,
	metadata.Uint32Kind:   appendU64,
	metadata.Uint64Kind:   appendU64,
	metadata.Fixed64Kind:  appendU64,
	metadata.Fixed32Kind:  appendU64,
	metadata.BoolKind:     appendBool,
	metadata.FloatKind:    appendF32,
	metadata.DoubleKind:   appendF64,
}

func appendS32(e *Encoder, x uint64) {
	x = uint64((uint32(x) >> 1) ^ uint32((int32(x&1)<<31)>>31))
	e.buf = strconv.AppendInt(e.buf, int64(x), 10)
}

func appendS64(e *Encoder, x uint64) {
	x = (x >> 1) ^ uint64((int64(x&1)<<63)>>63)
	e.buf = strconv.AppendInt(e.buf, int64(x), 10)
}

func appendI64(e *Encoder, x uint64) {
	e.buf = strconv.AppendInt(e.buf, int64(x), 10)
}

func appendU64(e *Encoder, x uint64) {
	e.buf = strconv.AppendUint(e.buf, x, 10)
}

func appendF32(e *Encoder, x uint64) {
	e.buf = strconv.AppendFloat(e.buf, float64(math.Float32frombits(uint32(x))), 'f', -1, 32)
}

func appendF64(e *Encoder, x uint64) {
	e.buf = strconv.AppendFloat(e.buf, math.Float64frombits(x), 'f', -1, 64)
}

func appendBool(e *Encoder, x uint64) {
	e.buf = strconv.AppendBool(e.buf, x != 0)
}

var defaultValues = [...]string{
	metadata.Int32Kind:    "0",
	metadata.Uint32Kind:   "0",
	metadata.Int64Kind:    "0",
	metadata.Uint64Kind:   "0",
	metadata.BoolKind:     "false",
	metadata.Sint32Kind:   "0",
	metadata.Sint64Kind:   "0",
	metadata.EnumKind:     "0",
	metadata.Fixed64Kind:  "0",
	metadata.Sfixed64Kind: "0",
	metadata.DoubleKind:   "0",
	metadata.Fixed32Kind:  "0",
	metadata.Sfixed32Kind: "0",
	metadata.FloatKind:    "0",
	metadata.StringKind:   `""`,
	metadata.BytesKind:    `""`,
	metadata.MessageKind:  "{}",
	metadata.MapKind:      "{}",
}
