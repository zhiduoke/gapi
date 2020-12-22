package jtop

import (
	"fmt"
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

func (t *Token) String() string {
	return fmt.Sprintf("Kind: [%s],Value:[%s]", tokenMap[t.Kind], t.Value)
}

type Iter struct {
	pos     int
	buf     []byte
	token   Token // only hold the last token
	topKind TokenKind
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
	return &Iter{pos: 0, buf: b, token: Token{Kind: Invalid}}
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
	i.token.Kind = Invalid
	i.token.Value = i.buf[i.pos:]
	return &i.token
}

func (i *Iter) ConsumeKind() TokenKind {
	return i.Consume().Kind
}

func (i *Iter) TopKind() TokenKind {
	if i.pos != 0 {
		return i.topKind
	}
	token := i.Consume()
	i.topKind = token.Kind
	i.pos = 0
	return token.Kind
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

func (i *Iter) setToken(kind TokenKind, val []byte) *Token {
	i.token.Kind = kind
	i.token.Value = val
	return &i.token
}

func makeFixedIter(kind TokenKind, size int) iterFunc {
	return func(i *Iter) *Token {
		if !i.request(size - 1) {
			return i.setToken(Invalid, i.buf[i.pos:])
		}
		begin := i.pos
		i.pos += size
		return i.setToken(kind, i.buf[begin:i.pos])
	}
}

func makeFixed1Iter(kind TokenKind) iterFunc {
	return func(i *Iter) *Token {
		begin := i.pos
		i.pos += 1
		return i.setToken(kind, i.buf[begin:i.pos])
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
	return i.setToken(Number, i.buf[begin:i.pos])
}

func iterString(i *Iter) *Token {
	begin := i.pos
	i.pos++
	for i.pos < len(i.buf) {
		c := i.buf[i.pos]
		if c == '"' && i.buf[i.pos-1] != '\\' {
			i.pos++
			return i.setToken(String, i.buf[begin:i.pos])
		}
		i.pos++
	}
	return i.setToken(Invalid, i.buf[begin:])
}
