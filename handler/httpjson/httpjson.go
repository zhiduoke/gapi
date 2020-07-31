package httpjson

import (
	"github.com/zhiduoke/gapi"
	"github.com/zhiduoke/gapi/metadata"
)

type Handler struct{}

func (h *Handler) HandleRequest(call *metadata.Call, ctx *gapi.Context) ([]byte, error) {
	return h.handleInput(call.In, ctx)
}

func (h *Handler) WriteResponse(call *metadata.Call, ctx *gapi.Context, data []byte) error {
	return h.handleOutput(call.Out, data, ctx.Response())
}
