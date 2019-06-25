package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
	"io/ioutil"
	"os"
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

	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		css.log.Error("cannot create tempdir")
		return nil, status.Errorf(codes.Internal, "cannot create tempdir")
	}
	defer os.RemoveAll(dir)

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:], chromedp.UserDataDir(dir),
		chromedp.Flag("start-fullscreen", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-infobars", true),
		)
	for _, opt := range req.Flags {
		oval := opt.GetValue()
		switch oval.(type) {
		case *pb.BrowserInitFlag_Strval:
			opts = append(opts, chromedp.Flag(opt.GetName(), opt.GetStrval()))
		case *pb.BrowserInitFlag_Bval:
			opts = append(opts, chromedp.Flag(opt.GetName(), opt.GetBval()))
		}
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(css.log.Infof))
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		css.log.Errorf("cannot start chrome: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot start chrome: %v", err))
	}

	path := filepath.Join(dir, "DevToolsActivePort")
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		css.log.Errorf("error reading DevToolsActivePort: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("error reading DevToolsActivePort: %v", err))
	}
//	lines := bytes.Split(bs, []byte("\n"))
	css.log.Debugf("DevToolsActivePort:\n%v", string(bs))

	return &empty.Empty{}, nil
}
