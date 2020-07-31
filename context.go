package gapi

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	req    *http.Request
	resp   http.ResponseWriter
	chain  []HandleFunc
	next   int
	values map[string]string
	params httprouter.Params
}

func (ctx *Context) Set(name string, value string) {
	if ctx.values == nil {
		ctx.values = map[string]string{}
	}
	ctx.values[name] = value
}

func (ctx *Context) Get(name string) (string, bool) {
	v, ok := ctx.values[name]
	return v, ok
}

func (ctx *Context) Request() *http.Request {
	return ctx.req
}

func (ctx *Context) Response() http.ResponseWriter {
	return ctx.resp
}

func (ctx *Context) SetResponse(r http.ResponseWriter) {
	ctx.resp = r
}

func (ctx *Context) Params() httprouter.Params {
	return ctx.params
}

func (ctx *Context) Next() error {
	if ctx.next < len(ctx.chain) {
		h := ctx.chain[ctx.next]
		ctx.next++
		return h(ctx)
	}
	return nil
}

func (ctx *Context) reset(w http.ResponseWriter, req *http.Request, params httprouter.Params, chain []HandleFunc) {
	ctx.req = req
	ctx.resp = w
	ctx.chain = chain
	ctx.next = 0
	ctx.values = nil
	ctx.params = params
}
