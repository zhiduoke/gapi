package pdparser

import (
	"io/ioutil"
	"testing"
)

func TestParse(t *testing.T) {
	data, err := ioutil.ReadFile("../../examples/demo/api/http.pd")
	if err != nil {
		t.Fatal(err)
	}
	md, err := ParseSet(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range md.Routes {
		t.Logf("%+v", route)
		t.Logf(">> %+v", route.Call)
	}
}
