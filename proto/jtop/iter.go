package jtop

import (
	"fmt"
	"sync"
)

type TokenKind uint8

const (
	Invalid TokenKind = iota + 0
	Null
	False
	True
	Number
	String
	ObjectBegin
	ObjectEnd
	ArrayBegin
	ArrayEnd
	Comma
	Colon
)

var tokenMap = [...]string{
	Invalid:     "Invalid",
	Null:        "Null",
	False:       "False",
	True:        "True",
	Number:      "Number",
	String:      "String",
	ObjectBegin: "ObjectBegin",
	ObjectEnd:   "ObjectEnd",
	ArrayBegin:  "ArrayBegin",
	ArrayEnd:    "ArrayEnd",
	Comma:       "Comma",
	Colon:       "Colon",
}

type Token struct {
	Kind  TokenKind
	Value []byte
}

var tokenPool sync.Pool

func newToken(kind TokenKind, val []byte) *Token {
	v := tokenPool.Get()
	if v == nil {
		return &Token{Kind: kind, Value: val}
	}
	token := v.(*Token)
	token.Kind = kind
	token.Value = val
	return token
}

func (t *Token) String() string {
	return fmt.Sprintf("Kind: [%s],Value:[%s]", tokenMap[t.Kind], t.Value)
}

func (t *Token) PutBack() {
	tokenPool.Put(t)
}

type Iter struct {
	pos int
	buf []byte
}

type iterFunc func(*Iter) *Token

var iterMatch = [...]iterFunc{
	't': makeFixedIter(True, 4),
	'f': makeFixedIter(False, 5),
	'n': makeFixedIter(Null, 4),
	'[': makeFixed1Iter(ArrayBegin),
	']': makeFixed1Iter(ArrayEnd),
	'{': makeFixed1Iter(ObjectBegin),
	'}': makeFixed1Iter(ObjectEnd),
	',': makeFixed1Iter(Comma),
	':': makeFixed1Iter(Colon),
	'"': iterString,
	'-': iterNumber,
	'0': iterNumber,
	'1': iterNumber,
	'2': iterNumber,
	'3': iterNumber,
	'4': iterNumber,
	'5': iterNumber,
	'6': iterNumber,
	'7': iterNumber,
	'8': iterNumber,
	'9': iterNumber,
}

func NewIter(b []byte) *Iter {
	return &Iter{pos: 0, buf: b}
}

func (i *Iter) Next() bool {
	i.skipWhiteSpace()
	return !i.eof()
}

func (i *Iter) Consume() *Token {
	fn := iterMatch[i.buf[i.pos]]
	if fn != nil {
		return fn(i)
	}
	return newToken(Invalid, i.buf[i.pos:])
}

func (i *Iter) ConsumeKind() TokenKind {
	tk := i.Consume()
	kind := tk.Kind
	tk.PutBack()
	return kind
}

func (i *Iter) IsObject() bool {
	i.skipWhiteSpace()
	if i.eof() {
		return false
	}
	return i.buf[i.pos] == '{'
}

func (i *Iter) Bytes() []byte {
	return i.buf
}

func (i *Iter) SetBuf(b []byte) {
	i.buf = b
	i.pos = 0
}

func (i *Iter) skipWhiteSpace() {
	for ; i.pos < len(i.buf); i.pos++ {
		c := i.buf[i.pos]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		return
	}
}

// request enough bytes for token
func (i *Iter) request(need int) bool {
	return i.pos+need < len(i.buf)
}

func (i *Iter) eof() bool {
	return i.pos >= len(i.buf)
}

func makeFixedIter(kind TokenKind, size int) iterFunc {
	return func(i *Iter) *Token {
		if !i.request(size - 1) {
			return newToken(Invalid, i.buf[i.pos:])
		}
		begin := i.pos
		i.pos += size
		return newToken(kind, i.buf[begin:i.pos])
	}
}

func makeFixed1Iter(kind TokenKind) iterFunc {
	return func(i *Iter) *Token {
		begin := i.pos
		i.pos += 1
		return newToken(kind, i.buf[begin:i.pos])
	}
}

func iterNumber(i *Iter) *Token {
	begin := i.pos
	for i.pos < len(i.buf) {
		c := i.buf[i.pos]
		if c >= '0' && c <= '9' || c == '-' || c == '.' || c == 'e' || c == 'E' {
			i.pos++
			continue
		}
		break
	}
	return newToken(Number, i.buf[begin:i.pos])
}

func iterString(i *Iter) *Token {
	begin := i.pos
	i.pos++
	for i.pos < len(i.buf) {
		c := i.buf[i.pos]
		if c == '"' && i.buf[i.pos-1] != '\\' {
			i.pos++
			return newToken(String, i.buf[begin:i.pos])
		}
		i.pos++
	}
	return newToken(Invalid, i.buf[begin:])
}
