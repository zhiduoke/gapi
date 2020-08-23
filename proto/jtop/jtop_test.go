package jtop

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/jtop/testdata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"testing"
)

type protoreflecter interface {
	ProtoReflect() protoreflect.Message
}

func TestEncode(t *testing.T) {
	type TestCase struct {
		name string
		in   interface{}
		msg  *metadata.Message
	}

	cases := [...]TestCase{
		{
			name: "number",
			in:   &numberReq,
			msg:  testdata.TestMessages[".jtop.test.NumberReq"],
		},
		{
			name: "string",
			in:   &stringReq,
			msg:  testdata.TestMessages[".jtop.test.StringReq"],
		},
		{
			name: "bool",
			in:   &boolReq,
			msg:  testdata.TestMessages[".jtop.test.BoolReq"],
		},
		{
			name: "object",
			in:   &objectReq,
			msg:  testdata.TestMessages[".jtop.test.ObjectReq"],
		},
		{
			name: "array",
			in:   &arrayReq,
			msg:  testdata.TestMessages[".jtop.test.ArrayReq"],
		},
		{
			name: "map",
			in:   &mapReq1,
			msg:  testdata.TestMessages[".jtop.test.MapReq"],
		},
	}
	for _, c := range cases {
		t.Logf("test %s\n", c.name)
		jsonData, err := json.Marshal(c.in)
		if err != nil {
			t.Log(err)
			return
		}
		//t.Logf("json: %s\n", jsonData)
		r, err := Encode(c.msg, jsonData)
		if err != nil {
			t.Errorf("encode error: %s\n", err)
			return
		}
		r1, err := proto.Marshal(proto.MessageV1(c.in.(protoreflecter).ProtoReflect()))
		if err != nil {
			t.Errorf("proto marshal error: %s\n", err)
			return
		}
		if !reflect.DeepEqual(r, r1) {
			diffbytes(t, r, r1)
			t.Errorf("protobuf not equal\n")
			return
		}
		t.Logf("pass!\n")
	}
}

func Benchmark_JTOPEncode(b *testing.B) {
	jsonData, _ := json.Marshal(objectReq)
	msg := testdata.TestMessages[".jtop.test.ObjectReq"]
	for i := 0; i < b.N; i++ {
		Encode(msg, jsonData)
	}
}

func Benchmark_ProtoEncode(b *testing.B) {
	jsonData, _ := json.Marshal(objectReq)
	o := new(testdata.ObjectReq)
	for i := 0; i < b.N; i++ {
		json.Unmarshal(jsonData, o)
		msg := proto.MessageV1(o.ProtoReflect())
		proto.Marshal(msg)
	}
}

func diffbytes(t *testing.T, a, b []byte) {
	t.Log("pb_jtop:", a)
	t.Log("pb_prot:", b)
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Log(a[:i])
			t.Log(a[i:])
			t.Log(b[i:])
			break
		}
	}
}

var numberReq = testdata.NumberReq{
	I32:    32,
	I64:    -64,
	Ui32:   32,
	Ui64:   64,
	Si32:   -32,
	Si64:   -64,
	Float:  66.66,
	Double: -66.66,
	Fix32:  32,
	Fix64:  64,
	Sfix32: -32,
	Sfix64: -64,
}

var boolReq = testdata.BoolReq{
	A: true,
	B: false,
}

var stringReq = testdata.StringReq{
	Str:   "\u1234\\\u2222\t\n😪😪😪😪",
	Bae64: []byte("hello,i am ok"),
}

var objectReq1 = testdata.ObjectReq{
	Num:  &numberReq,
	Str:  &stringReq,
	Bool: &boolReq,
	Obj:  nil,
	A:    200,
	B:    false,
}

var objectReq = testdata.ObjectReq{
	Num:  &numberReq,
	Str:  &stringReq,
	Bool: &boolReq,
	Obj:  &objectReq1,
	A:    100,
	B:    true,
}

var arrayReq = testdata.ArrayReq{
	Nums:  []int32{1, 2, 3, 4, 5},
	Strs:  []string{"1", "22", "333", "4444", "55555"},
	Bools: []bool{true, false, false, true},
	Objs:  []*testdata.ObjectReq{&objectReq, &objectReq1, &objectReq1},
}

var mapReq = testdata.MapReq{
	Sms: map[string]string{"a": "1a", "b": "2b"},
	Smi: map[string]int32{"1": 1, "2": 2},
	//Bms: map[bool]string{true: "true", false: "false"},
	Smo: map[string]*testdata.ObjectReq{"obj0": &objectReq, "obj1": &objectReq1},
	//Imo: map[int32]*testdata.ObjectReq{0: &objectReq, 1: &objectReq1},
	Sma: map[string]*testdata.ArrayReq{"a": &arrayReq, "b": &arrayReq},
}

var mapReq1 = testdata.MapReq{
	Sms: map[string]string{"a": "1a"},
	Smi: map[string]int32{"1": 1},
	//Bms: map[bool]string{true: "true", false: "false"},
	Smo: map[string]*testdata.ObjectReq{"obj0": &objectReq},
	//Imo: map[int32]*testdata.ObjectReq{0: &objectReq, 1: &objectReq1},
	Sma: map[string]*testdata.ArrayReq{"a": &arrayReq},
}
