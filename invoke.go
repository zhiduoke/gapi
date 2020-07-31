package gapi

import (
	"github.com/zhiduoke/gapi/metadata"
	"google.golang.org/grpc"
)

type callCodec struct {
	call *metadata.Call
	h    CallHandler
}

func (c *callCodec) Marshal(v interface{}) ([]byte, error) {
	cc := v.(*Context)
	return c.h.HandleRequest(c.call, cc)
}

func (c *callCodec) Unmarshal(data []byte, v interface{}) error {
	cc := v.(*Context)
	return c.h.WriteResponse(c.call, cc, data)
}

func (c *callCodec) Name() string {
	return "call"
}

func defaultDial(server string) (*grpc.ClientConn, error) {
	return grpc.Dial(server, grpc.WithInsecure())
}
