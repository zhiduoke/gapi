package pbjson

import (
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/tests/msgs"
)

func TestEncode(t *testing.T) {
	msgmd := &metadata.Message{
		Name: "testmsg",
		Fields: []*metadata.Field{
			{
				Tag:      1,
				Name:     "a",
				Kind:     metadata.Int32Kind,
				Repeated: false,
			},
			{
				Tag:      2,
				Name:     "b",
				Kind:     metadata.Int32Kind,
				Repeated: false,
				Options: metadata.FieldOptions{
					OmitEmpty: false,
				},
			},
			{
				Tag:      3,
				Name:     "c",
				Kind:     metadata.StringKind,
				Repeated: false,
				Options: metadata.FieldOptions{
					OmitEmpty: true,
				},
			},
		},
	}
	msgmd.BakeTagIndex()
	msg := msgs.SimpleMessage1{
		A: 1,
		B: 2,
		C: "",
	}
	b1, _ := json.Marshal(&msg)
	t.Log(string(b1))
	b2, _ := proto.Marshal(&msg)
	e := NewEncoder(nil)
	e.EncodeMessage(msgmd, b2)
	if e.Error() != nil {
		t.Fatal(e.Error())
	}
	t.Log(string(e.Bytes()))
}
