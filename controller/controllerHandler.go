package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/je4/bremote/api"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type ControllerServiceServer struct {
	controller *Controller
	log        *logging.Logger
}

func NewControllerServiceServer(controller *Controller, log *logging.Logger) ControllerServiceServer {
	css := ControllerServiceServer{controller: controller, log: log}
	return css
}

func (css ControllerServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/Ping( %v )", traceId, sourceInstance, targetInstance, param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (css ControllerServiceServer) NewClient(ctx context.Context, param *pb.NewClientParam) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	client := param.GetClient()
	stat := param.GetStatus()
	css.log.Infof("[%v] %v -> %v/NewClient( %v, %v )", traceId, sourceInstance, targetInstance, client, stat)

	if stat != common.ClientStatus_Empty {
		// todo: handle effective status of client
		return &empty.Empty{}, nil
	}

	// new client should start the browser...
	go func() {

		time.Sleep(time.Millisecond * 300)

		cw := api.NewClientWrapper(css.controller.GetInstance(), css.controller.GetSessionPtr())

		opts := map[string]interface{}{
			"headless":              false,
			"start-fullscreen":      true,
			"disable-notifications": true,
			"disable-infobars":      true,
			"disable-gpu":           false,
		}

		traceId = uniqid.New(uniqid.Params{"traceid_", false})
		css.log.Infof("[%v] starting browser of %v", traceId, client)
		if err := cw.StartBrowser(traceId, client, &opts); err != nil {
			css.log.Errorf("[%v] error starting client browser on %v: %v", traceId, client, err)
		}
	}()

	return &empty.Empty{}, nil
}

func (css ControllerServiceServer) RemoveClient(ctx context.Context, param *pb.String) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	client := param.GetValue()
	css.log.Infof("[%v] %v -> %v/RemoveClient( %v )", traceId, sourceInstance, targetInstance, client)

	return &empty.Empty{}, nil
}
