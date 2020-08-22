package jtop

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
	"google.golang.org/protobuf/encoding/protowire"
	"io"
	"math"
	"strconv"
	"sync"
	"unsafe"
)

type Encoder struct {
	buf       *proto.Buffer
	err       error
	iter      *Iter
	rootField *metadata.Field // cache
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

func Encode(msg *metadata.Message, data []byte) ([]byte, error) {
	enc := newEncoder()
	defer putEncoder(enc)
	iter := NewIter(data)
	if iter.TopKind() != ObjectBegin {
		return nil, errors.New("invalid json input: must be json object")
	}
	enc.reset(iter)
	enc.rootField.Message = msg
	enc.transValue(enc.rootField)
	if enc.err != nil {
		return nil, enc.err
	}
	buf := append([]byte(nil), enc.buf.Bytes()...)
	return buf, nil
}

func (e *Encoder) reset(iter *Iter) {
	e.iter = iter
	if e.rootField == nil {
		e.rootField = &metadata.Field{Name: "root_field", Kind: metadata.MessageKind}
	}
	e.rootField.Message = nil
	if e.buf == nil {
		e.buf = proto.NewBuffer(nil)
	}
	e.buf.Reset()
}

func (e *Encoder) setErrorMissMatch(jsonType string, pbKind metadata.TypeKind) {
	e.err = fmt.Errorf("type miss match: json type (%s) and protobuf kind (%d)", jsonType, pbKind)
}

func (e *Encoder) setErrorInvalidJsonToken(token *Token, err error) {
	if err == nil {
		err = errors.New("")
	}
	e.err = fmt.Errorf("invalid json input value: [%s] :%s", token.Value, err)
}

func (e *Encoder) setErrorInvalidJsonFormat(err error) {
	if err != nil {
		e.err = fmt.Errorf("invalid json format :%s", err)
		return
	}
	e.err = fmt.Errorf("invalid json format")
}

func (e *Encoder) encodeBytes(tag int, v []byte) {
	k := protowire.EncodeTag(protowire.Number(tag), protowire.BytesType)
	e.buf.EncodeVarint(k)
	e.buf.EncodeRawBytes(v)
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

func (e *Encoder) transBool(token *Token, field *metadata.Field) {
	if field.Kind != metadata.BoolKind {
		e.setErrorMissMatch("bool", field.Kind)
		return
	}
	wire, pv, ok := e.parseNumber(token, field)
	if !ok {
		return
	}
	e.encodeKey(field.Tag, wire)
	e.encodeWire(wire, pv)
}

func (e *Encoder) transNull(token *Token, field *metadata.Field) {
	if !fieldNullable(field) {
		e.setErrorMissMatch("null", field.Kind)
		return
	}
	v := token.Value
	if v[0] == 'n' && v[1] == 'u' && v[2] == 'l' && v[3] == 'l' {
		return
	}
	e.err = fmt.Errorf("invalid json input value: %s", token.Value)
}

func (e *Encoder) transNumber(token *Token, field *metadata.Field) {
	wire, pv, ok := e.parseNumber(token, field)
	if !ok {
		return
	}
	e.encodeKey(field.Tag, wire)
	e.encodeWire(wire, pv)
}

func (e *Encoder) transString(token *Token, field *metadata.Field) {
	var pv []byte
	switch field.Kind {
	case metadata.StringKind:
		s, ok := e.unquoteString(token.Value)
		if !ok {
			e.setErrorInvalidJsonToken(token, errors.New("invalid string format"))
			return
		}
		pv = s
	case metadata.BytesKind:
		maxLen := base64.StdEncoding.DecodedLen(len(token.Value) - 2)
		b := make([]byte, maxLen)
		l, err := base64.StdEncoding.Decode(b, token.Value[1:len(token.Value)-1])
		if err != nil {
			e.setErrorInvalidJsonToken(token, err)
			return
		}
		pv = b[:l]
	default:
		e.setErrorMissMatch("string", field.Kind)
		return
	}
	e.encodeBytes(field.Tag, pv)
}

func (e *Encoder) transObjectAsMap(_ *Token, field *metadata.Field) {
	msg := field.Message
	if len(msg.Fields) != 2 {
		e.err = errors.New("invalid metadata: map type error")
		return
	}

	var kvField = [...]*metadata.Field{msg.Fields[0], msg.Fields[1]}
	kvEnc := newEncoder()
	kvEnc.reset(e.iter)
	done := false

KvEncode:
	for {
		kvEnc.reset(e.iter)
		for i := 0; i < 2; i++ {
			field := kvField[i]
			kind, ok := kvEnc.transValue(field)
			if !ok {
				break KvEncode
			}
			if kind == ObjectEnd {
				done = true
				break KvEncode
			}
			if !isValueToken(kind) {
				i--
			}
		}
		e.encodeBytes(field.Tag, kvEnc.buf.Bytes())
	}

	if !done && kvEnc.err == nil {
		e.setErrorInvalidJsonFormat(io.ErrUnexpectedEOF)
	}

	if kvEnc.err != nil {
		e.err = kvEnc.err
	}

	putEncoder(kvEnc)
}

func (e *Encoder) transObject(token *Token, field *metadata.Field) {
	if field.Kind == metadata.MapKind {
		e.transObjectAsMap(token, field)
		return
	}
	if field.Kind != metadata.MessageKind {
		e.setErrorMissMatch("object", field.Kind)
		return
	}

	msg := field.Message
	objEnc := e
	root := e.rootField.Message != nil
	if !root {
		objEnc = newEncoder()
		objEnc.reset(e.iter)
	} else {
		e.rootField.Message = nil
	}

	done := false

	for objEnc.iter.Next() {
		tk := objEnc.iter.Consume()
		if tk.Kind == ObjectEnd {
			done = true
			break
		}
		if tk.Kind == Comma {
			continue
		}
		if tk.Kind != String {
			objEnc.setErrorInvalidJsonToken(tk, errors.New("unexpected key"))
			break
		}
		// read key
		key, ok := objEnc.unquoteString(tk.Value)
		if !ok {
			objEnc.setErrorInvalidJsonToken(tk, errors.New("invalid string format"))
			break
		}
		objEnc.ignoreToken()
		field := msg.GetField(string(key))
		if field != nil {
			objEnc.transValue(field)
			continue
		}
		objEnc.ignoreValueTokens()
	}
	if !done && objEnc.err == nil {
		objEnc.setErrorInvalidJsonFormat(io.ErrUnexpectedEOF)
	}

	if objEnc.err != nil {
		e.err = objEnc.err
		putEncoder(objEnc)
		return
	}

	if !root {
		e.encodeBytes(field.Tag, objEnc.buf.Bytes())
		putEncoder(objEnc)
	}
}

func (e *Encoder) packNumeric(_ *Token, field *metadata.Field) {
	packEnc := newEncoder()
	packEnc.reset(e.iter)
	for packEnc.iter.Next() {
		tk := packEnc.iter.Consume()
		if tk.Kind == ArrayEnd {
			break
		}
		if tk.Kind != Number && tk.Kind != True && tk.Kind != False {
			continue
		}
		wire, pv, ok := packEnc.parseNumber(tk, field)
		if !ok {
			return
		}
		packEnc.encodeWire(wire, pv)
	}
	e.encodeBytes(field.Tag, packEnc.buf.Bytes())
	putEncoder(packEnc)
}

func (e *Encoder) transArray(token *Token, field *metadata.Field) {
	if !field.Repeated {
		e.setErrorMissMatch("array", field.Kind)
		return
	}
	// https://developers.google.com/protocol-buffers/docs/encoding#packed
	// packed scalar numeric types
	if isNumeric(field.Kind) {
		e.packNumeric(token, field)
		return
	}

	for {
		kind, ok := e.transValue(field)
		if !ok || kind == ArrayEnd || kind == Invalid {
			break
		}
	}
}

func (e *Encoder) transValue(filed *metadata.Field) (TokenKind, bool) {
	if !e.iter.Next() || e.err != nil {
		return Invalid, false
	}
	token := e.iter.Consume()
	kind := token.Kind
	switch token.Kind {
	case Invalid:
		e.setErrorInvalidJsonToken(token, nil)
	case Null:
		e.transNull(token, filed)
	case True, False:
		e.transBool(token, filed)
	case Number:
		e.transNumber(token, filed)
	case String:
		e.transString(token, filed)
	case ObjectBegin:
		e.transObject(token, filed)
	case ArrayBegin:
		e.transArray(token, filed)
	case ObjectEnd, ArrayEnd, Comma, Colon:
	default:
		e.setErrorInvalidJsonToken(token, nil)
	}
	return kind, e.err == nil
}

func (e *Encoder) unquoteString(data []byte) ([]byte, bool) {
	// TODO:(thewinds)
	return data[1 : len(data)-1], true
}

func (e *Encoder) parseNumber(token *Token, field *metadata.Field) (wire protowire.Type, pv uint64, ok bool) {
	wire = protowire.VarintType
	sval := *(*string)(unsafe.Pointer(&token.Value))
	var err error
	switch field.Kind {
	case metadata.DoubleKind:
		wire = protowire.Fixed64Type
		var fv float64
		fv, err = strconv.ParseFloat(sval, 64)
		pv = math.Float64bits(fv)
	case metadata.FloatKind:
		wire = protowire.Fixed32Type
		var fv float64
		fv, err = strconv.ParseFloat(sval, 32)
		pv = uint64(math.Float32bits(float32(fv)))
	case metadata.Int32Kind:
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 32)
		pv = uint64(fv)
	case metadata.Int64Kind:
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 64)
		pv = uint64(fv)
	case metadata.Sint32Kind:
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 32)
		pv = protowire.EncodeZigZag(fv)
	case metadata.Sint64Kind:
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 64)
		pv = protowire.EncodeZigZag(fv)
	case metadata.Uint32Kind:
		pv, err = strconv.ParseUint(sval, 10, 32)
	case metadata.Uint64Kind:
		pv, err = strconv.ParseUint(sval, 10, 64)
	case metadata.Fixed32Kind:
		wire = protowire.Fixed32Type
		pv, err = strconv.ParseUint(sval, 10, 32)
	case metadata.Sfixed32Kind:
		wire = protowire.Fixed32Type
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 32)
		pv = uint64(fv)
	case metadata.Fixed64Kind:
		wire = protowire.Fixed64Type
		pv, err = strconv.ParseUint(sval, 10, 64)
	case metadata.Sfixed64Kind:
		wire = protowire.Fixed64Type
		var fv int64
		fv, err = strconv.ParseInt(sval, 10, 64)
		pv = uint64(fv)
	case metadata.BoolKind:
		v := token.Value
		if token.Kind == True && v[0] == 't' && v[1] == 'r' && v[2] == 'u' && v[3] == 'e' {
			pv = 1
			break
		}
		if token.Kind == False && v[0] == 'f' && v[1] == 'a' && v[2] == 'l' && v[3] == 's' && v[4] == 'e' {
			pv = 0
			break
		}
		err = fmt.Errorf("invalid value: %s", token.Value)
	default:
		e.setErrorMissMatch("number", field.Kind)
		err = e.err
		return
	}
	if err != nil {
		e.setErrorInvalidJsonToken(token, err)
		return
	}
	return wire, pv, true
}

func (e *Encoder) ignoreToken() {
	if e.iter.Next() {
		e.iter.ConsumeKind()
	}
}

func (e *Encoder) ignoreValueTokens() {
	if !e.iter.Next() {
		return
	}
	var hold, want TokenKind
	firstKind := e.iter.ConsumeKind()
	switch firstKind {
	case Null, False, True, Number, String:
		return
	case ObjectBegin:
		hold = ObjectBegin
		want = ObjectEnd
	case ArrayBegin:
		hold = ArrayBegin
		want = ArrayEnd
	default:
		return
	}
	diff := 1
	for diff != 0 && e.iter.Next() {
		kind := e.iter.ConsumeKind()
		if kind == hold {
			diff++
			continue
		}
		if kind == want {
			diff--
		}
	}
}
