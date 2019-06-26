package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
	"io/ioutil"
	"path/filepath"
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

func (css ClientServiceServer) StartBrowser(ctx context.Context, req *pb.BrowserInitFlags) (*empty.Empty, error) {
	css.log.Infof("StartBrowser()")

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

	browser, err := NewBrowser(execOptions, css.log)
	if err != nil {
		css.log.Errorf("initialize browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot initialize brower: %v", err))
	}
	defer browser.Close()

	// ensure that the browser process is started
	if err := browser.Run(); err != nil {
		css.log.Errorf("cannot run browser: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot run brower: %v", err))
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
