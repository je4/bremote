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

type DataproxyWrapper struct {
	instanceName           string
	session                **yamux.Session
	dataProxyServiceClient *DataproxyServiceClient
	conn                   *grpc.ClientConn
}

func NewDataproxyWrapper(instanceName string, session **yamux.Session) *DataproxyWrapper {
	cw := &DataproxyWrapper{instanceName: instanceName,
		session:                session,
		dataProxyServiceClient: nil,
		conn:                   nil,
	}
	return cw
}

func (cw *DataproxyWrapper) connect() (err error) {
	if *cw.session == nil {
		cw.dataProxyServiceClient = nil
		return errors.New(fmt.Sprintf("session closed"))
	}

	// it's a singleton
	if cw.dataProxyServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	// singleton!!!
	doDial := cw.conn == nil
	if cw.conn != nil {
		if cw.conn.GetState() == connectivity.TransientFailure {
			cw.conn.Close()
			doDial = true
		}
		if cw.conn.GetState() == connectivity.Shutdown {
			doDial = true
		}
	}
	if doDial {
		cw.conn, err = grpc.Dial(":7777", grpc.WithInsecure(),
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
	}
	c := NewDataproxyServiceClient(cw.conn)
	cw.dataProxyServiceClient = &c
	return nil
}

func (cw *DataproxyWrapper) Ping(traceId string, targetInstance string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId)
	pingResult, err := (*cw.dataProxyServiceClient).Ping(ctx, &String{Value: "ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging %v", targetInstance)
	}
	return pingResult.GetValue(), nil
}

func (cw *DataproxyWrapper) SetWhitelist(traceId string, targetInstance string, whitelist []string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance, "traceId", traceId)
	param := &StringList{
		String_: whitelist,
	}

	_, err := (*cw.dataProxyServiceClient).SetWhitelist(ctx, param)
	if err != nil {
		return emperror.Wrap(err, "error navigating client")
	}
	return nil
}
