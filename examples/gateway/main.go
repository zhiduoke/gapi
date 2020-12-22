package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiduoke/gapi"
	"github.com/zhiduoke/gapi/handler/httpjson"
	"github.com/zhiduoke/gapi/metadata"
	"github.com/zhiduoke/gapi/proto/pdparser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func jsonError(w http.ResponseWriter, code int, msg string) {
	type Errmsg struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	data, _ := json.Marshal(&Errmsg{
		Code:    code,
		Message: msg,
	})
	w.Write(data)
}

func loadMetaddata() (*metadata.Metadata, error) {
	data, err := ioutil.ReadFile("../demo/api/http.pd")
	if err != nil {
		return nil, err
	}
	return pdparser.ParseSet(data)
}

func main() {
	s := gapi.NewServer()
	s.Use(func(ctx *gapi.Context) error {
		err := ctx.Next()
		if err != nil {
			gerr := status.Convert(err)
			jsonError(ctx.Response(), int(gerr.Code()), gerr.Message())
		}
		return nil
	})
	s.RegisterMiddleware("auth", func(ctx *gapi.Context) error {
		uid := ctx.Request().FormValue("uid")
		if uid == "" {
			return status.Errorf(codes.Unauthenticated, "missing uid")
		}
		logrus.Infof("logged uid: %s", uid)
		ctx.Set("uid", uid)
		return ctx.Next()
	})
	s.RegisterMiddleware("mock_ctx", func(ctx *gapi.Context) error {
		ctx.Set("from_ctx", time.Now().String())
		return ctx.Next()
	})
	s.RegisterHandler("httpjson", &httpjson.Handler{})
	md, err := loadMetaddata()
	if err != nil {
		logrus.Fatalf("loadMetadata: %v", err)
	}
	logrus.Println("http server listen on :8080")
	for _, route := range md.Routes {
		logrus.Println(route.Method, route.Path)
	}
	err = s.UpdateRoute(md)
	if err != nil {
		logrus.Fatalf("UpdateRoute: %v", err)
	}
	err = http.ListenAndServe(":8080", s)
	if err != nil {
		logrus.Fatalf("serve: %v", err)
	}
	// a curl test
	// curl -X POST -H 'from_header: hello' -d 'from_form=xxx' http://localhost:8080/demo/request_bind/sssdsdsd\?from_query\=222ss
}
