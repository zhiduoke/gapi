package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

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
	s.RegisterHandler("httpjson", &httpjson.Handler{})
	md, err := loadMetaddata()
	if err != nil {
		logrus.Fatalf("loadMetadata: %v", err)
	}
	err = s.UpdateRoute(md)
	if err != nil {
		logrus.Fatalf("UpdateRoute: %v", err)
	}
	err = http.ListenAndServe(":8080", s)
	if err != nil {
		logrus.Fatalf("serve: %v", err)
	}
}
