package main

import (
	"context"
	"github.com/op/go-logging"
	pb "info_age.net/bremote/api"
)

type ClientServiceServer struct {
	log *logging.Logger
}

func NewClientServiceServer(log *logging.Logger) ClientServiceServer {
	css := ClientServiceServer{log: log}
	return css
}

func (css ClientServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	css.log.Infof("Ping( %v )", param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}
