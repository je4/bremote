package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
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
	// it's a singleton
	if cw.clientServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	conn, err := grpc.Dial(":7777", grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			if cw.session == nil {
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

func (cw *ClientWrapper) Ping( targetInstance string ) (string, error) {
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance )
	pingResult, err := (*cw.clientServiceClient).Ping(ctx, &String{Value:"ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging %v", targetInstance)
	}
	return pingResult.GetValue(), nil
}