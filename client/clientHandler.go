package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
	"info_age.net/bremote/browser"
	"info_age.net/bremote/common"
	"io/ioutil"
	"path/filepath"
)

type ClientServiceServer struct {
	client *Client
	log *logging.Logger
}

func NewClientServiceServer(client *Client, log *logging.Logger) ClientServiceServer {
	css := ClientServiceServer{client:client, log: log}
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

	// build option map
	execOptions := make( map[string]interface{})
	for _, opt := range req.Flags {
		oval := opt.GetValue()
		switch oval.(type) {
		case *pb.BrowserInitFlag_Strval:
			execOptions[opt.GetName()] = opt.GetStrval()
		case *pb.BrowserInitFlag_Bval:
			execOptions[opt.GetName()] = opt.GetBval()
		}
	}

	browser, err := browser.NewBrowser(execOptions, css.log)
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
	return &empty.Empty{}, nil
}