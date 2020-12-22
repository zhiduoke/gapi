package testdata

import (
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/pdparser"
	"io/ioutil"
	"log"
)

var TestMessages = map[string]*metadata.Message{}

func init() {
	data, err := ioutil.ReadFile("testdata/test.pd")
	if err != nil {
		log.Fatal(err)
	}
	md, err := pdparser.ParseSet(data)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range md.Routes {
		TestMessages[v.Call.In.Name] = v.Call.In
	}
}
