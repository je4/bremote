package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DataProxyServiceServer struct {
	dp  *DataProxy
	log *logging.Logger
}

func NewDataProxyServiceServer(client *DataProxy, log *logging.Logger) DataProxyServiceServer {
	css := DataProxyServiceServer{dp: client, log: log}
	return css
}

func (dpss DataProxyServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		dpss.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	dpss.log.Infof("[%v] %v -> %v/Ping( %v )", traceId, sourceInstance, targetInstance, param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (dpss DataProxyServiceServer) SetWhitelist(ctx context.Context, param *pb.StringList) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		dpss.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	dpss.log.Infof("[%v] %v -> %v/SetWhitelist( %v )", traceId, sourceInstance, targetInstance, param.GetString_())

	dpss.dp.SetWhitelist(param.GetString_())

	return &empty.Empty{}, nil
}
