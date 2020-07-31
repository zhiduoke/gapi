package httpview

import (
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/zhiduoke/gapi"
	"github.com/zhiduoke/gapi/handler/httpview/viewproto"
	"github.com/zhiduoke/gapi/metadata"
)

type Handler struct {
}

func (*Handler) HandleRequest(call *metadata.Call, ctx *gapi.Context) ([]byte, error) {
	req := ctx.Request()
	var headers map[string]string
	if len(req.Header) > 0 {
		headers = make(map[string]string, len(req.Header))
		for k, v := range req.Header {
			if len(v) == 0 {
				headers[k] = ""
			} else {
				headers[k] = v[0]
			}
		}
	}
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&viewproto.HttpRequest{
		Method:  call.Name,
		Query:   req.URL.RawQuery,
		Headers: headers,
		Body:    body,
	})
}

func (*Handler) WriteResponse(_ *metadata.Call, ctx *gapi.Context, data []byte) error {
	var resp viewproto.HttpResponse
	err := proto.Unmarshal(data, &resp)
	if err != nil {
		return err
	}
	w := ctx.Response()
	if len(resp.Headers) > 0 {
		h := w.Header()
		for k, v := range resp.Headers {
			h.Set(k, v)
		}
	}
	w.WriteHeader(int(resp.Status))
	_, err = w.Write(resp.Body)
	return err
}
