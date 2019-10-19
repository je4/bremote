package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/mintance/go-uniqid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"net"
)

type ControllerWrapper struct {
	instanceName            string
	session                 **yamux.Session
	controllerServiceClient *ControllerServiceClient
	conn                    *grpc.ClientConn
}

func NewControllerWrapper(instanceName string, session **yamux.Session) *ControllerWrapper {
	cw := &ControllerWrapper{instanceName: instanceName, session: session, controllerServiceClient: nil}
	return cw
}

func (pw *ControllerWrapper) connect() (err error) {
	if *pw.session == nil {
		pw.controllerServiceClient = nil
		return errors.New(fmt.Sprintf("session closed"))
	}

	// it's a singleton
	if pw.controllerServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	// singleton!!!
	doDial := pw.conn == nil
	if pw.conn != nil {
		if pw.conn.GetState() == connectivity.TransientFailure {
			pw.conn.Close()
			doDial = true
		}
		if pw.conn.GetState() == connectivity.Shutdown {
			doDial = true
		}
	}
	if doDial {
		pw.conn, err = grpc.Dial(":7777", grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				if *pw.session == nil {
					return nil, errors.New(fmt.Sprintf("session %s closed", s))
				}
				return (*pw.session).Open()
			}),
		)
		if err != nil {
			return errors.New("cannot dial grpc connection to :7777")
		}
	}
	c := NewControllerServiceClient(pw.conn)
	pw.controllerServiceClient = &c
	return nil
}

func (cw *ControllerWrapper) Ping(traceId string, targetInstance string, param string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceid", traceId)
	pingResult, err := (*cw.controllerServiceClient).Ping(ctx, &String{Value: param})
	if err != nil {
		return "", emperror.Wrapf(err, "error calling %s::Ping(%s)", targetInstance, param)
	}
	return pingResult.GetValue(), nil
}

func (cw *ControllerWrapper) NewClient(traceId string, targetInstance string, client string, clientStatus string, clientHttpAddr string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"sourceInstance", cw.instanceName,
		"targetInstance", targetInstance,
		"traceid", traceId)
	_, err := (*cw.controllerServiceClient).NewClient(ctx, &NewClientParam{Client: client, Status: clientStatus, HttpAddr: clientHttpAddr})
	if err != nil {
		return emperror.Wrapf(err, "error calling %x::NewClient(%s)", targetInstance, client)
	}
	return nil
}
