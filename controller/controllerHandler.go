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
	"net/url"
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
			"headless":                            false,
			"start-fullscreen":                    true,
			"disable-notifications":               true,
			"disable-infobars":                    true,
			"disable-gpu":                         false,
			"allow-insecure-localhost":            true,
			"enable-immersive-fullscreen-toolbar": true,
			"views-browser-windows":               false,
			"kiosk":                               true,
			"disable-session-crashed-bubble":      true,
			"incognito":                           true,
			"disable-features":                    "InfiniteSessionRestore",
			//"no-first-run":                        true,
			"enable-fullscreen-toolbar-reveal": false,
			"useAutomationExtension":           false,
			"enable-automation":                false,
		}

		traceId = uniqid.New(uniqid.Params{"traceid_", false})
		css.log.Infof("[%v] starting browser of %v", traceId, client)
		if err := cw.StartBrowser(traceId, client, &opts); err != nil {
			css.log.Errorf("[%v] error starting client browser on %v: %v", traceId, client, err)
		}

		// check for autonavigation
		data, err := css.controller.GetVar(client, "autonav")
		if err == nil {
			autonav := data.(map[string]interface{})
			// check for url
			urlstring, ok := autonav["url"]
			if !ok {
				return
			}
			u, err := url.Parse(urlstring.(string))
			if err != nil {
				css.log.Errorf("cannot parse url %v: %v", urlstring.(string), err)
				return
			}
			nextStatus := autonav["nextstatus"].(string)
			css.log.Infof("%v::NewClient(%v) - autonav: %v -> %v", client, u.String(), nextStatus)

			cw := pb.NewClientWrapper(css.controller.instance, css.controller.GetSessionPtr())
			traceId := uniqid.New(uniqid.Params{"traceid_", false})
			err = cw.Navigate(traceId, client, u, nextStatus)
			if err != nil {
				css.log.Errorf("cannot navigate to %v: %v", u.String(), err)
				return
			}

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

func (css ControllerServiceServer) WebsocketMessage(ctx context.Context, req *pb.Bytes) (*empty.Empty, error) {
	traceId, sourceInstance, targetGroup, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "WebSocketMessage()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "WebSocketMessage()", err))
	}

	css.log.Infof("[%v] %v -> /ws() -> %v", traceId, sourceInstance, targetGroup)

	// todo: send to local webservice of target group

	return &empty.Empty{}, nil
}
