package httpjson

import (
	"github.com/zhiduoke/gapi"
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/jtop"
	"github.com/zhiduoke/gapi/proto/kvpb"
	"io/ioutil"
)

func (h *Handler) handleInput(msg *metadata.Message, ctx *gapi.Context) ([]byte, error) {
	contentType := ctx.Request().Header.Get("Content-Type")
	var (
		pb  []byte
		err error
	)

	if contentType == "application/json" {
		// json
		var body []byte
		body, err = ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			return nil, err
		}
		pb, err = jtop.Encode(msg, body)
		if err != nil {
			return nil, err
		}
	}
	httppb, err := kvpb.Encode(msg, &httpKV{ctx: ctx})
	if err != nil {
		return nil, err
	}
	pb = append(pb, httppb...)
	return pb, nil
}

type httpKV struct {
	ctx *gapi.Context
}

func (h *httpKV) GetForm(key string) (string, bool) {
	v := h.ctx.Request().FormValue(key)
	return v, len(v) > 0
}

func (h *httpKV) GetContext(key string) (string, bool) {
	return h.ctx.Get(key)
}

func (h *httpKV) GetHeader(key string) (string, bool) {
	v := h.ctx.Request().Header.Get(key)
	return v, len(v) > 0
}

func (h *httpKV) GetQuery(key string) (string, bool) {
	v := h.ctx.Request().URL.Query().Get(key)
	return v, len(v) > 0
}

func (h *httpKV) GetParams(key string) (string, bool) {
	v := h.ctx.Params().ByName(key)
	return v, len(v) > 0
}
