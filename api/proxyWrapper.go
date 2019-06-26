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
	"info_age.net/bremote/common"
	"net"
)

type ProxyWrapper struct {
	instanceName       string
	session            **yamux.Session
	proxyServiceClient *ProxyServiceClient
}

func NewProxyWrapper(instanceName string, session **yamux.Session) *ProxyWrapper {
	cw := &ProxyWrapper{instanceName: instanceName, session: session, proxyServiceClient: nil}
	return cw
}

func (pw *ProxyWrapper) connect() error {
	if *pw.session == nil {
		pw.proxyServiceClient = nil
		return errors.New(fmt.Sprintf("session closed"))
	}

	// it's a singleton
	if pw.proxyServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	conn, err := grpc.Dial(":7777", grpc.WithInsecure(),
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
	c := NewProxyServiceClient(conn)
	pw.proxyServiceClient = &c
	return nil
}

func (cw *ProxyWrapper) Ping(traceId string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	pingResult, err := (*cw.proxyServiceClient).Ping(ctx, &String{Value: "ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging")
	}
	return pingResult.GetValue(), nil
}

func (cw *ProxyWrapper) Init(traceId string, instance string, sessionType common.SessionType) (error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).Init(ctx, &InitParam{Instance:&String{Value:instance},SessionType:ProxySessionType(sessionType)})
	if err != nil {
		return emperror.Wrapf(err, "error initializing instance")
	}
	return nil
}

func (cw *ProxyWrapper) GetClients(traceId string, t common.SessionType) ([]string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return []string{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	clients, err := (*cw.proxyServiceClient).GetClients(ctx, &empty.Empty{})
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
