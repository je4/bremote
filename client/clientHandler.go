package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/browser"
	"github.com/je4/bremote/common"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"path/filepath"
)

type ClientServiceServer struct {
	client *BrowserClient
	log    *logging.Logger
}

func NewClientServiceServer(client *BrowserClient, log *logging.Logger) ClientServiceServer {
	css := ClientServiceServer{client: client, log: log}
	return css
}

func (css ClientServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/Ping( %v )", traceId, sourceInstance, targetInstance, param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (css ClientServiceServer) StartBrowser(ctx context.Context, req *pb.BrowserInitFlags) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/StartBrowser()", traceId, sourceInstance, targetInstance)

	if css.client.browser != nil {
		if css.client.browser.IsRunning() {
			//css.client.SetStatus(common.ClientStatus_EmptyBrowser)
			css.log.Infof("[%v] browser already running", traceId)
			return &empty.Empty{}, nil
		} else {
			if err := css.client.browser.Startup(); err != nil {
				css.log.Errorf("cannot startup browser: %v", err)
				return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot cannot browser: %v", err))
			}
			return &empty.Empty{}, nil
		}
	}
	// build option map
	execOptions := make(map[string]interface{})
	for _, opt := range req.Flags {
		oval := opt.GetValue()
		switch oval.(type) {
		case *pb.BrowserInitFlag_Strval:
			execOptions[opt.GetName()] = opt.GetStrval()
		case *pb.BrowserInitFlag_Bval:
			execOptions[opt.GetName()] = opt.GetBval()
		}
	}

	browser, err := browser.NewBrowser(execOptions, css.log, css.client.writeBrowserLog)
	if err != nil {
		css.log.Errorf("initialize browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot initialize browser: %v", err))
	}
	//defer browser.Close()
	if err := css.client.SetBrowser(browser); err != nil {
		css.log.Errorf("cannot set browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot set browser: %v", err))
	}

	// ensure that the browser process is started
	if err := browser.Run(); err != nil {
		css.log.Errorf("cannot run browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot run browser: %v", err))
	}

	path := filepath.Join(browser.TempDir, "DevToolsActivePort")
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		css.log.Errorf("error reading DevToolsActivePort: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("error reading DevToolsActivePort: %v", err))
	}
	//	lines := bytes.Split(bs, []byte("\n"))
	css.log.Debugf("DevToolsActivePort:\n%v", string(bs))

	css.client.SetStatus(common.ClientStatus_EmptyBrowser)
	return &empty.Empty{}, nil
}

func (css ClientServiceServer) Navigate(ctx context.Context, req *pb.NavigateParam) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "LoadPage()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/Navigate(%v, %v)", traceId, sourceInstance, targetInstance, req.GetUrl(), req.GetNextStatus())

	b, err := css.client.GetBrowser()
	if err != nil {
		css.log.Errorf("could not get browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("could not get browser: %v", err))
	}
	if !b.IsRunning() {
		if err := b.Startup(); err != nil {
			css.log.Errorf("could not start browser: %v", err)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("could not start browser: %v", err))
		}
	}

	tasks := chromedp.Tasks{
		chromedp.Navigate(req.GetUrl()),
//		browser.MouseClickXYAction(2,2),
	}
	err = b.Tasks(tasks)
	if err != nil {
		css.log.Errorf("could not navigate: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("could not navigate: %v", err))
	}
	css.client.SetStatus(req.GetNextStatus())
	return &empty.Empty{}, nil
}

func (css ClientServiceServer) ShutdownBrowser(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/ShutdownBrowser()", traceId, sourceInstance, targetInstance)
	css.client.ShutdownBrowser()
	css.client.SetStatus(common.ClientStatus_Empty)
	return &empty.Empty{}, nil
}

func (css ClientServiceServer) GetBrowserLog(ctx context.Context, param *empty.Empty) (*pb.BrowserLog, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v::GetStatus(): %v", css.client.GetInstance(), err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/GetBrowserLog()", traceId, sourceInstance, targetInstance)
	if !css.client.browser.IsRunning() {
		css.client.SetStatus(common.ClientStatus_Empty)
	}

	return &pb.BrowserLog{Entry: css.client.getBrowserLog()}, nil
}

func (css ClientServiceServer) GetStatus(ctx context.Context, param *empty.Empty) (*pb.String, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v::GetStatus(): %v", css.client.GetInstance(), err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/GetStatus()", traceId, sourceInstance, targetInstance)
	if css.client.browser == nil {
		css.client.SetStatus(common.ClientStatus_Empty)
	} else {
		if !css.client.browser.IsRunning() {
			css.client.SetStatus(common.ClientStatus_Empty)
		}
	}
	return &pb.String{Value: css.client.GetStatus()}, nil
}

func (css ClientServiceServer) GetHTTPSAddr(ctx context.Context, param *empty.Empty) (*pb.String, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v::GetHTTPSAddr(): %v", css.client.GetInstance(), err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/GetHTTPSAddr()", traceId, sourceInstance, targetInstance)
	return &pb.String{Value: css.client.GetHTTPAddr()}, nil
}

func (css ClientServiceServer) SetStatus(ctx context.Context, param *pb.String) (*empty.Empty, error) {
	stat := param.GetValue()
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v::SetStatus(%v): %v", css.client.GetInstance(), stat, err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	css.log.Infof("[%v] %v -> %v/SetStatus(%v)", traceId, sourceInstance, targetInstance, stat)
	css.client.SetStatus(stat)
	return &empty.Empty{}, nil
}

func (css ClientServiceServer) WebsocketMessage(ctx context.Context, req *pb.Bytes) (*empty.Empty, error) {
	traceId, sourceInstance, targetGroup, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "WebSocketMessage()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "WebSocketMessage()", err))
	}

	css.log.Infof("[%v] %v -> /ws() -> %v", traceId, sourceInstance, targetGroup)

	err = css.client.SendGroupWebsocket(targetGroup, req.GetValue())
	if err != nil {
		css.log.Errorf("cannot send message to %v: %v", targetGroup, err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot send message to %v: %v", targetGroup, err))
	}

	return &empty.Empty{}, nil
}

func (css ClientServiceServer) MouseClick(ctx context.Context, param *pb.ClickMessage) (*empty.Empty, error) {
	traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
	if err != nil {
		css.log.Errorf("invalid metadata in call to %v: %v", "MouseClick()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "MouseClick()", err))
	}

	css.log.Infof("[%v] %v -> %v/MouseClick()", traceId, sourceInstance, targetInstance)
	coord := param.GetCoord()
	if coord == nil {
		css.log.Errorf("only coordinates are supported in call to %v", "MouseClick()")
		return nil, status.Errorf(codes.Unimplemented, fmt.Sprintf("only coordinates are supported in call to %v", "MouseClick()"))
	}
	if err := css.client.browser.MouseClickXY(coord.GetX(), coord.GetY()); err != nil {
		css.log.Errorf("error in call to %v: %v", "MouseClick()", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("error in call to %v: %v", "MouseClick()", err))
	}

	return nil, status.Errorf(codes.Unimplemented, "method MouseClick not implemented")
}