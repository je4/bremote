package main

import (
	"crypto/tls"
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"
	"sync"
)

type Proxy struct {
	log      *logging.Logger
	certFile string
	keyFile  string
	tlsCfg   *tls.Config
	sessions map[string]*ProxySession

	sync.RWMutex
}

/*
create a new Proxy instance
*/
func NewProxy(certFile string, keyFile string, log *logging.Logger) (*Proxy, error) {
	proxy := &Proxy{log: log,
		certFile: certFile,
		keyFile:  keyFile,
		sessions: make(map[string]*ProxySession)}
	if err := proxy.Init(); err != nil {
		return nil, emperror.Wrap(err, "cannot connect")
	}
	return proxy, nil
}

/*
initialize struct
*/
func (proxy *Proxy) Init() (err error) {
	// load tls files
	cert, err := tls.LoadX509KeyPair(proxy.certFile, proxy.keyFile)
	if err != nil {
		return err
	}
	// create tls configuration
	proxy.tlsCfg = &tls.Config{Certificates: []tls.Certificate{cert}}

	return
}

func (proxy *Proxy) AddSession(session *ProxySession, name string) error {
	proxy.log.Debugf("add session %v", name)
	proxy.Lock()
	defer proxy.Unlock()

	if _, ok := proxy.sessions[name]; ok {
		return errors.New("session already exists")
	}
	proxy.sessions[name] = session
	return nil
}

func (proxy *Proxy) GetSession(name string) (*ProxySession, error) {
	proxy.RLock()
	defer proxy.RUnlock()

	val, ok := proxy.sessions[name]
	if !ok {
		return nil, errors.New("session not found")
	}
	return val, nil
}

func (proxy *Proxy) CloseSession(name string) error {
	proxy.log.Debugf("close session %v", name)
	session, err := proxy.GetSession(name)
	if err != nil {
		return emperror.Wrapf(err, "error removing session %v", name)
	}
	session.Close()
	return nil
}

func (proxy *Proxy) RemoveSession(name string) (*ProxySession, error) {
	proxy.log.Debugf("remove session %v", name)
	proxy.Lock()
	defer proxy.Unlock()


	val, ok := proxy.sessions[name]
	if !ok {
		return nil, errors.New("session not found")
	}
	delete(proxy.sessions, name)
	return val, nil
}

func (proxy *Proxy) RenameSession(oldname string, newname string) error {
	proxy.log.Debugf("rename session %v -> %v", oldname, newname)
	proxy.Lock()
	defer proxy.Unlock()


	if _, ok := proxy.sessions[newname]; ok {
		return errors.New("session already exists")
	}

	val, ok := proxy.sessions[oldname]
	if !ok {
		return errors.New("session not found")
	}
	delete(proxy.sessions, oldname)
	proxy.sessions[newname] = val
	return nil
}

func (proxy *Proxy) ListenServe() (err error) {

	listener, err := tls.Listen("tcp", addr, proxy.tlsCfg)
	if err != nil {
		return emperror.Wrapf(err, "cannot start tcp listener on %v", addr)
	}
	defer listener.Close()

	for {
		proxy.log.Infof("waiting for incoming TLS connections on %v", addr)
		// Accept blocks until there is an incoming TCP connection
		incoming, err := listener.Accept()
		if err != nil {
			return emperror.Wrap(err, "couldn't accept incoming connection")
		}

		session, err := yamux.Client(incoming, nil)
		if err != nil {
			return emperror.Wrap(err, "couldn't create yamux")
		}

		proxy.log.Info("launching a gRPC server over incoming TCP connection")
		go proxy.Serve(session)
	}
	return
}

func (proxy *Proxy) Serve(session *yamux.Session) error {
	// create a server instance
	instancename := session.RemoteAddr().String()

	ps := NewProxySession(instancename, session, proxy, proxy.log)
	defer func() {
		if err := ps.Close(); err != nil {
			proxy.log.Errorf("error closing proxy session %v: %v", ps.GetInstance(), err)
		}
	}()

	return ps.Serve()
}

func (proxy *Proxy) Close() error {
	return nil
}
