package main

import (
	"context"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/je4/grpc-proxy/proxy"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
	"strings"
	"sync"
)

type SessionType int

const (
	SessionType_Undefined SessionType = 0
	SessionType_Client    SessionType = 1
	SessionType_Command   SessionType = 2
)

type ProxySession struct {
	log         *logging.Logger
	instance    string
	proxy       *Proxy
	sessionType SessionType
	service     *ProxyServiceServer
	grpcServer  *grpc.Server
	session     *yamux.Session
}

func NewProxySession(instance string, session *yamux.Session, proxy *Proxy, log *logging.Logger) *ProxySession {
	ps := &ProxySession{instance:instance, session:session, proxy:proxy, log:log, sessionType: SessionType_Undefined}
	return ps
}

func (ps *ProxySession) Serve() error {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if err := ps.ServeGRPC(); err != nil {
			ps.log.Errorf("error serving GRPC for instance %v: %v", ps.GetInstance(), err)
		}
		wg.Done()
	}()

	if err := ps.proxy.AddSession(ps, ps.GetInstance()); err != nil {
		return emperror.Wrapf(err, "cannot add session %v", ps.instance)
	}

	wg.Wait()
	return nil
}

func (ps *ProxySession) ServeGRPC() error {

	director := func(ctx context.Context, fullMethodName string) (*grpc.ClientConn, error) {
		// Make sure we never forward internal services.
		if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
			return nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		//md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			// Decide on which backend to dial
			if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
				// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
				return grpc.DialContext(ctx, "api-service.staging.svc.local", grpc.WithDefaultCallOptions(grpc.ForceCodec(proxy.Codec())))
			} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
				return grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithDefaultCallOptions(grpc.ForceCodec(proxy.Codec())))
			}
		}
		return nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}

	_ = director

	// create a gRPC server object
	ps.grpcServer = grpc.NewServer()

	ps.service = NewProxyServiceServer(ps, ps.log)

	// attach the Ping service to the server
	pb.RegisterProxyServiceServer(ps.grpcServer, ps.service)

	// start the gRPC erver
	ps.proxy.log.Info("launching gRPC server over TLS connection...")
	if err := ps.grpcServer.Serve(ps.session); err != nil {
		ps.proxy.log.Errorf("failed to serve: %v", err )
		return emperror.Wrapf(err, "failed to serve")
	}

	return nil
}

func (ps *ProxySession) GetService() *ProxyServiceServer {
	return ps.service
}

func (ps *ProxySession) GetInstance() string {
	return ps.instance
}

func (ps *ProxySession) SetInstance(newinstance string) error {
	if err := ps.proxy.RenameSession(ps.instance, newinstance); err != nil {
		return emperror.Wrapf(err, "error renaming %v -> %v", ps.instance, newinstance)
	}
	ps.instance = newinstance
	return nil
}

func (ps *ProxySession) GetSessionType() SessionType {
	return ps.sessionType
}

func (ps *ProxySession) SetSessionType(sessionType SessionType) {
	ps.sessionType = sessionType
	ps.log.Debugf("set session type of %v to %v", ps.instance, sessionType)
}

func (ps *ProxySession) GetProxy() *Proxy {
	return ps.proxy
}

func (ps *ProxySession) Close() error {
	_, err := ps.proxy.RemoveSession(ps.instance)
	if err != nil {
		return emperror.Wrapf(err, "cannot remove session %v", ps.instance)
	}
	ps.grpcServer.GracefulStop()
	return ps.session.Close()
}
