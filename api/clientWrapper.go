package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/mintance/go-uniqid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"reflect"
)

type ClientWrapper struct {
	instanceName string
	session **yamux.Session
	clientServiceClient *ClientServiceClient
}

func NewClientWrapper( instanceName string, session **yamux.Session) *ClientWrapper {
	cw := &ClientWrapper{instanceName:instanceName, session:session, clientServiceClient:nil}
	return cw
}

func (cw *ClientWrapper) connect() error {
	if *cw.session == nil {
		cw.clientServiceClient = nil
		return errors.New(fmt.Sprintf("session closed"))
	}

	// it's a singleton
	if cw.clientServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	conn, err := grpc.Dial(":7777", grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			if *cw.session == nil {
				return nil, errors.New(fmt.Sprintf("session %s closed", s))
			}
			return (*cw.session).Open()
		}),
	)
	if err != nil {
		return errors.New("cannot dial grpc connection to :7777")
	}
	c := NewClientServiceClient(conn)
	cw.clientServiceClient = &c
	return nil
}



func (cw *ClientWrapper) Ping( traceId string, targetInstance string ) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId )
	pingResult, err := (*cw.clientServiceClient).Ping(ctx, &String{Value:"ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging %v", targetInstance)
	}
	return pingResult.GetValue(), nil
}

func (cw *ClientWrapper) StartBrowser( traceId string, targetInstance string, execOptions *map[string]interface{} ) (error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId )

	flags := []*BrowserInitFlag{}
	for name, val := range *execOptions {
		bif := &BrowserInitFlag{}
		bif.Name = name
		switch val.(type) {
		case bool:
			bif.Value = &BrowserInitFlag_Bval{val.(bool)}
		case string:
			bif.Value = &BrowserInitFlag_Strval{val.(string)}
		default:
			return errors.New(fmt.Sprintf("invalid value type %v for %v=%v", reflect.TypeOf(val), name, val))
		}

		flags = append(flags, bif)
	}
	browserInitFlags := &BrowserInitFlags{Flags:flags}
	_, err := (*cw.clientServiceClient).StartBrowser(ctx, browserInitFlags)
	if err != nil {
		return emperror.Wrap(err, "error starting browser")
	}
	return nil
}

func (cw *ClientWrapper) ShutdownBrowser( traceId string, targetInstance string) (error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect to %v",  targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId )
	_, err := (*cw.clientServiceClient).ShutdownBrowser(ctx, &empty.Empty{})
	if err != nil {
		return emperror.Wrapf(err, "error shutting down browser of %v", targetInstance)
	}
	return nil
}

func (cw *ClientWrapper) GetStatus( traceId string, targetInstance string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect to %v",  targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId )
	ret, err := (*cw.clientServiceClient).GetStatus(ctx, &empty.Empty{})
	if err != nil {
		return "", emperror.Wrapf(err, "error getting status of %v", targetInstance)
	}
	return ret.GetValue(), nil
}

func (cw *ClientWrapper) SetStatus( traceId string, targetInstance string, stat string) (error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect to %v",  targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId )
	_, err := (*cw.clientServiceClient).SetStatus(ctx, &String{Value:stat})
	if err != nil {
		return emperror.Wrapf(err, "error setting status of %v", targetInstance)
	}
	return nil
}
