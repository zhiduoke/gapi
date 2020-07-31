package main

import (
	"net"

	"github.com/zhiduoke/gapi/examples/demo/api"
	"github.com/zhiduoke/gapi/examples/demo/service"
	"google.golang.org/grpc"
)

func main() {
	srv := grpc.NewServer()
	api.RegisterDemoAPIServer(srv, &service.API{})
	ln, err := net.Listen("tcp", ":19090")
	if err != nil {
		panic(err)
	}
	err = srv.Serve(ln)
	if err != nil {
		panic(err)
	}
}
