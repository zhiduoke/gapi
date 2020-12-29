package pbjson

import (
	"gapi/metadata"
	"testing"
)

func newMessage(name string, fields []*metadata.Field) *metadata.Message {
	msg := &metadata.Message{
		Name:   name,
		Fields: fields,
	}
	msg.BakeTagIndex()
	return msg
}

func getMsgFooType() *metadata.Message {
	msg := newMessage("pbmsg.Foo", []*metadata.Field{
		{
			Tag:     1,
			Name:    "a",
			Kind:    metadata.StringKind,
			Options: metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:     2,
			Name:    "b",
			Kind:    metadata.BoolKind,
			Options: metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:     3,
			Name:    "c",
			Kind:    metadata.Int32Kind,
			Options: metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:  4,
			Name: "d",
			Kind: metadata.MessageKind,
			Message: newMessage("pbmsg.Foo.Embed", []*metadata.Field{
				{Tag: 1, Name: "a", Kind: metadata.Int32Kind, Options: metadata.FieldOptions{OmitEmpty: true}},
				{Tag: 2, Name: "b", Kind: metadata.StringKind, Options: metadata.FieldOptions{OmitEmpty: true}},
			}),
			Options: metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:      5,
			Name:     "e",
			Kind:     metadata.Int32Kind,
			Repeated: true,
			Options:  metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:      6,
			Name:     "f",
			Kind:     metadata.StringKind,
			Repeated: true,
			Options:  metadata.FieldOptions{OmitEmpty: true},
		}, {
			Tag:  7,
			Name: "g",
			Kind: metadata.MessageKind,
			Message: newMessage("pbmsg.Elem", []*metadata.Field{
				{Tag: 1, Name: "a", Kind: metadata.Int32Kind, Options: metadata.FieldOptions{OmitEmpty: true}},
				{Tag: 2, Name: "s", Kind: metadata.StringKind, Options: metadata.FieldOptions{OmitEmpty: true}},
			}),
			Repeated: true,
			Options:  metadata.FieldOptions{OmitEmpty: true},
		},
	})
	return msg
}

func runTransProtoToJson(s []byte, ty *metadata.Message) {
	enc := NewEncoder(nil)
	enc.EncodeMessageFast(ty, s)
	if enc.err != nil {
		panic(enc.err)
	}
}

var (
	pbCase0 = []byte{}
	pbCase1 = []byte{
		10, 1, 97, 16, 1, 24, 1, 34, 5, 8, 2, 18, 1, 98, 42, 3, 3, 4, 5, 50, 2, 102, 48, 50, 2, 102,
		49, 50, 2, 102, 50, 58, 6, 8, 6, 18, 2, 115, 48, 58, 6, 8, 7, 18, 2, 115, 49,
	}
)

func BenchmarkTransProtoToJsonCase0(b *testing.B) {
	ty := getMsgFooType()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTransProtoToJson(pbCase0, ty)
	}
}

func BenchmarkTransProtoToJsonCase1(b *testing.B) {
	ty := getMsgFooType()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTransProtoToJson(pbCase1, ty)
	}
}
