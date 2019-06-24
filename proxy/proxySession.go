package main

import (
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "info_age.net/bremote/api"
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
