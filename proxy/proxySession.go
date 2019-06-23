package main

import (
	"github.com/goph/emperror"
	"github.com/op/go-logging"
)

type SessionType int

const (
	Undefined SessionType = 0
	Client    SessionType = 1
	Command   SessionType = 2
)

type ProxySession struct {
	log         *logging.Logger
	instance    string
	proxy       *Proxy
	sessionType SessionType
	service     *ProxyServiceServer
}

func NewProxySession(instance string, proxy *Proxy, log *logging.Logger) *ProxySession {
	ps := &ProxySession{log:log, instance:instance, proxy:proxy, sessionType:Undefined}
	ps.service = NewProxyServiceServer(ps, log)
	proxy.AddSession(ps, instance)
	return ps
}

func (ps *ProxySession) GetService() *ProxyServiceServer {
	return ps.service
}

func (ps *ProxySession) GetInstance() string{
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

func (ps *ProxySession) GetProxy() *Proxy{
	return ps.proxy
}
