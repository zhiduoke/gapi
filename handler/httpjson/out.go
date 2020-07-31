package httpjson

import (
	"net/http"
	"sync"

	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/pbjson"
)

var encoderPool sync.Pool

func (h *Handler) handleOutput(msg *metadata.Message, data []byte, out http.ResponseWriter) error {
	var e *pbjson.Encoder
	if v := encoderPool.Get(); v != nil {
		e = v.(*pbjson.Encoder)
		e.Reset()
	} else {
		e = pbjson.NewEncoder(make([]byte, 0, len(data)))
	}
	e.EncodeMessage(msg, data)
	err := e.Error()
	if err != nil {
		encoderPool.Put(e)
		return err
	}
	out.Header().Set("Content-Type", "application/json")
	out.WriteHeader(http.StatusOK)
	_, err = out.Write(e.Bytes())
	encoderPool.Put(e)
	return err
}
