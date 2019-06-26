package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/je4/grpc-proxy/proxy"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
	"info_age.net/bremote/common"
	"net"
	"sync"
)


type ProxySession struct {
	log         *logging.Logger
	instance    string
	proxy       *Proxy
	sessionType common.SessionType
	service     *ProxyServiceServer
	grpcServer  *grpc.Server
	session     *yamux.Session
}

func NewProxySession(instance string, session *yamux.Session, proxy *Proxy, log *logging.Logger) *ProxySession {
	ps := &ProxySession{instance: instance, session: session, proxy: proxy, log: log, sessionType: common.SessionType_Undefined}
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

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// Make sure we never forward internal services.
		//if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
		//	return nil, status.Errorf(codes.Unimplemented, "Unknown method")
		//}
		traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
		if err != nil {
			ps.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
			return nil, nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
		}

		// check for session
		sess, err := ps.proxy.GetSession(targetInstance)
		if err != nil {
			ps.log.Errorf("[%v] instance not found in call to %v::%v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)
			return nil, nil, status.Errorf(codes.Unavailable, "[%v] instance not found in call to %v::%v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)
		}

		// make sure, that we transfer the metadata to the target client
		ctx = metadata.AppendToOutgoingContext(ctx, "sourceInstance", sourceInstance, "targetInstance", targetInstance, "traceId", traceId )

		conn, err := grpc.DialContext(ctx, ":7777",
			grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				if sess.session == nil {
					return nil, errors.New(fmt.Sprintf("[%v] session %s closed", traceId, s))
				}
				return sess.session.Open()
			}),
			grpc.WithCodec(proxy.Codec()),
//			grpc.WithDefaultCallOptions(grpc.ForceCodec(proxy.Codec())),
		)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "[%v] error dialing %v on session %v for %v", traceId, ":7777", sess.GetInstance(), fullMethodName)
		}

		ps.log.Debugf("[%v] directing %v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)

		return ctx, conn, nil
	}

	// create a gRPC server object
	ps.grpcServer = grpc.NewServer(grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	ps.service = NewProxyServiceServer(ps, ps.log)

	// attach the Ping service to the server
	pb.RegisterProxyServiceServer(ps.grpcServer, ps.service)

	// start the gRPC erver
	ps.proxy.log.Info("launching gRPC server over TLS connection...")
	if err := ps.grpcServer.Serve(ps.session); err != nil {
		ps.proxy.log.Errorf("failed to serve: %v", err)
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

func (ps *ProxySession) GetSessionPtr() **yamux.Session{
	return &ps.session
}

func (ps *ProxySession) SetInstance(newinstance string) error {
	if err := ps.proxy.RenameSession(ps.instance, newinstance); err != nil {
		return emperror.Wrapf(err, "error renaming %v -> %v", ps.instance, newinstance)
	}
	ps.instance = newinstance
	return nil
}

func (ps *ProxySession) GetSessionType() common.SessionType {
	return ps.sessionType
}

func (ps *ProxySession) SetSessionType(sessionType common.SessionType) {
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
