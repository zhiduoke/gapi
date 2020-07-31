package gapi

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/zhiduoke/gapi/metadata"
	"google.golang.org/grpc"
)

type routeHandler struct {
	s      *Server
	chain  []HandleFunc
	call   *metadata.Call
	ch     CallHandler
	client *grpc.ClientConn
}

func (h *routeHandler) invoke(ctx *Context) error {
	rpcctx := ctx.req.Context()
	call := h.call
	var cancel func()
	if call.Timeout != 0 {
		rpcctx, cancel = context.WithTimeout(rpcctx, call.Timeout)
	}
	err := h.client.Invoke(rpcctx, call.Name, ctx, ctx, grpc.ForceCodec(&callCodec{
		call: call,
		h:    h.ch,
	}))
	if cancel != nil {
		cancel()
	}
	return err
}

func (h *routeHandler) handle(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := h.s.ctxpool.Get().(*Context)
	ctx.reset(w, req, params, h.chain)
	err := ctx.Next()
	if err != nil {
		logrus.Errorf("handle route: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	h.s.ctxpool.Put(ctx)
}
