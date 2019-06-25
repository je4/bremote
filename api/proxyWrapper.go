package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"info_age.net/bremote/common"
	"net"
)

type ProxyWrapper struct {
	instanceName string
	session **yamux.Session
	proxyServiceClient *ProxyServiceClient
}

func NewProxyWrapper( instanceName string, session **yamux.Session) *ProxyWrapper {
	cw := &ProxyWrapper{instanceName:instanceName, session:session, proxyServiceClient:nil}
	return cw
}

func (cw *ProxyWrapper) connect() error {
	// it's a singleton
	if cw.proxyServiceClient != nil {
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
	c := NewProxyServiceClient(conn)
	cw.proxyServiceClient = &c
	return nil
}

func (cw *ProxyWrapper) Ping( targetInstance string ) (string, error) {
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect to %v", targetInstance)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "targetInstance", targetInstance )
	pingResult, err := (*cw.proxyServiceClient).Ping(ctx, &String{Value:"ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging %v", targetInstance)
	}
	return pingResult.GetValue(), nil
}

func (cw *ProxyWrapper) GetClients( t common.SessionType ) ([]string, error) {
	if err := cw.connect(); err != nil {
		return []string{}, emperror.Wrapf(err, "cannot connect"	)
	}

	clients, err := (*cw.proxyServiceClient).GetClients(context.Background(), &empty.Empty{})
	if err != nil {
		return []string{}, emperror.Wrap(err, "cannot get clients")
	}
	ret := []string{}
	for _, c := range clients.GetClients() {
		// we only want we need
		if t != common.SessionType_All {
			if c.GetType() != ProxySessionType(t) {
				continue
			}
		}
		ret = append(ret, c.GetInstance())
	}
	return ret, nil
}
