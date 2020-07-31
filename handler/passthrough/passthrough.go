package passthrough

import (
	"io/ioutil"

	"github.com/zhiduoke/gapi"
	"github.com/zhiduoke/gapi/metadata"
)

type Handler struct {
}

func (h *Handler) HandleRequest(_ *metadata.Call, ctx *gapi.Context) ([]byte, error) {
	body, err := ioutil.ReadAll(ctx.Request().Body)
	ctx.Request().Body.Close()
	return body, err
}

func (h *Handler) WriteResponse(_ *metadata.Call, ctx *gapi.Context, data []byte) error {
	_, err := ctx.Response().Write(data)
	return err
}
