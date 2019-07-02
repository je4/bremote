package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"
	"io/ioutil"
	"reflect"
	"sync"
)

/*
the proxy manages all client and controller sessions
*/
type Proxy struct {
	log      *logging.Logger
	instance string
	addr     string
	caFile   string
	certFile string
	keyFile  string
	tlsCfg   *tls.Config
	sessions map[string]*ProxySession

	sync.RWMutex
}

/*
create a new Proxy instance
*/
func NewProxy(config Config, log *logging.Logger) (*Proxy, error) {
	proxy := &Proxy{log: log,
		instance: config.InstanceName,
		addr:     config.TLSAddr,
		caFile:   config.CaPEM,
		certFile: config.CertPEM,
		keyFile:  config.KeyPEM,
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

	var certpool *x509.CertPool
	clientAuth := tls.NoClientCert

	// if we get a ca certificate, we do client auth...
	if proxy.caFile != "" {
		certpool = x509.NewCertPool()
		pem, err := ioutil.ReadFile(proxy.caFile)
		if err != nil {
			proxy.log.Fatalf("Failed to read client certificate authority: %v", err)
		}
		if !certpool.AppendCertsFromPEM(pem) {
			proxy.log.Fatalf("Can't parse client certificate authority")
		}
		clientAuth = tls.RequireAndVerifyClientCert
	}

	// create tls configuration
	proxy.tlsCfg = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   clientAuth,
		ClientCAs:    certpool,
	}

	return
}

/*
add a new session to the session list
*/
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

/*
get all sessions
*/
func (proxy *Proxy) GetSessions() map[string]*ProxySession {
	proxy.RLock()
	defer proxy.RUnlock()

	return proxy.sessions
}

/*
get session by name
*/
func (proxy *Proxy) GetSession(name string) (*ProxySession, error) {
	proxy.RLock()
	defer proxy.RUnlock()

	val, ok := proxy.sessions[name]
	if !ok {
		return nil, errors.New("session not found")
	}
	return val, nil
}

/*
remove session from session list and close the session
*/
func (proxy *Proxy) CloseSession(name string) error {
	proxy.log.Debugf("close session %v", name)
	session, err := proxy.GetSession(name)
	if err != nil {
		return emperror.Wrapf(err, "error removing session %v", name)
	}
	session.Close()
	return nil
}

/*
remove session from session list without closing it
*/
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

/*
rename session
needed if instance name is available after session creation
*/
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

	listener, err := tls.Listen("tcp", proxy.addr, proxy.tlsCfg)
	if err != nil {
		return emperror.Wrapf(err, "cannot start tcp listener on %v", proxy.addr)
	}
	defer listener.Close()

	for {
		proxy.log.Infof("waiting for incoming TLS connections on %v", proxy.addr)
		// Accept blocks until there is an incoming TCP connection
		incoming, err := listener.Accept()
		if err != nil {
			return emperror.Wrap(err, "couldn't accept incoming connection")
		}
		tlscon, ok := incoming.(*tls.Conn)
		if ok {
			proxy.log.Debug("conn: type assert to TLS succeedded")
			err := tlscon.Handshake()
			if err != nil {
				proxy.log.Errorf("handshake failed: %s", err)
			} else {
				proxy.log.Debug("conn: Handshake completed")
			}
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				publicKey, err := x509.MarshalPKIXPublicKey(v.PublicKey)
				if err != nil {
					proxy.log.Errorf("invalid public key %v: %v", reflect.TypeOf(v.PublicKey), v.PublicKey)
				}
				proxy.log.Debugf("client public key %v: %v", reflect.TypeOf(v.PublicKey), publicKey)
			}
		}
		session, err := yamux.Client(incoming, nil)
		if err != nil {
			return emperror.Wrap(err, "couldn't create yamux")
		}

		proxy.log.Info("launching a gRPC server over incoming TCP connection")
		go proxy.ServeSession(session)
	}
	return
}

func (proxy *Proxy) ServeSession(session *yamux.Session) error {
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
