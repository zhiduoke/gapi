package service

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/zhiduoke/gapi/examples/demo/api"
)

type API struct {
}

func (s *API) Add(ctx context.Context, in *api.AddRequest) (*api.AddReply, error) {
	logrus.Infof("Add1: %s", in)
	return &api.AddReply{
		Sum: in.A + in.B,
	}, nil
}

func (s *API) Add2(ctx context.Context, in *api.Add2Request) (*api.Add2Reply, error) {
	logrus.Infof("Add2: %s", in)
	return &api.Add2Reply{
		Sum: in.A + in.B,
		Uid: in.UserId,
	}, nil
}

func (s *API) Sub(ctx context.Context, in *api.SubReq) (*api.SubResp, error) {
	logrus.Infof("Sub: %s", in)
	return &api.SubResp{
		Result: in.A - in.B,
	}, nil
}

func (s *API) Sub2(ctx context.Context, in *api.SubReq2) (*api.SubResp2, error) {
	logrus.Infof("Sub2: %s", in)
	return &api.SubResp2{
		Result: in.A - in.B,
	}, nil
}
