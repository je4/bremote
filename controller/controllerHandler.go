package main

import (
	"context"
	"github.com/op/go-logging"
	pb "info_age.net/bremote/api"
)

type ControllerServiceServer struct {
	log *logging.Logger
}

func NewControllerServiceServer(log *logging.Logger) ControllerServiceServer {
	css := ControllerServiceServer{log: log}
	return css
}

func (css ControllerServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	css.log.Infof("Ping( %v )", param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}
