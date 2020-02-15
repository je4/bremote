package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/je4/bremote/common"
	"github.com/je4/ntp"
	"github.com/op/go-logging"
	"github.com/prologic/bitcask"
	"github.com/soheilhy/cmux"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
)

/*
the proxy manages all client and controller sessions
*/
type Proxy struct {
	log      *logging.Logger
	db       *bitcask.Bitcask
	instance string
	addr     string
	caFile   string
	certFile string
	keyFile  string
	tlsCfg   *tls.Config
	sessions map[string]*ProxySession
	groups   *InstanceGroups
	webRoot  string
	typeMap  map[string]common.SessionType

	sync.RWMutex
	listener net.Listener
	ntpRaw   func(data []byte) ([]byte, error)
}

/*
create a new Proxy instance
*/
func NewProxy(config Config, db *bitcask.Bitcask, log *logging.Logger) (*Proxy, error) {
	proxy := &Proxy{log: log,
		instance: config.InstanceName,
		db:       db,
		addr:     config.TLSAddr,
		caFile:   config.CaPEM,
		certFile: config.CertPEM,
		keyFile:  config.KeyPEM,
		sessions: make(map[string]*ProxySession),
		groups:   NewInstanceGroups(),
		webRoot:  config.WebRoot,
		typeMap:  common.SessionTypeInt,
		ntpRaw:   ntp.MakeDefaultHandler(config.NTPHost, "", "", "", 0, 0),
	}
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

	proxy.listener, err = net.Listen("tcp", proxy.addr)
	if err != nil {
		return emperror.Wrapf(err, "cannot start tcp listener on %v", proxy.addr)
	}
	defer proxy.listener.Close()

	// da mux
	m := cmux.New(proxy.listener)

	httpL := m.Match(cmux.HTTP1Fast())
	tlsL := m.Match(cmux.Any())

	/**
	listener, err := tls.Listen("tcp", proxy.addr, proxy.tlsCfg)
	if err != nil {
		return emperror.Wrapf(err, "cannot start tcp listener on %v", proxy.addr)
	}
	defer listener.Close()
	*/

	go func() {
		http.Handle("/", http.FileServer(http.Dir(proxy.webRoot)))
		proxy.log.Infof("Serving folder %v on %v\n", proxy.webRoot, proxy.addr)
		if err := http.Serve(httpL, nil); err != nil {
			proxy.log.Errorf("couldnt serve http: %v", err)
		} else {
			proxy.log.Infof("http server shutdown: %v", err)
		}
	}()

	go func() {
		proxy.log.Infof("Serving tls on %v\n", proxy.addr)
		if err := proxy.Serve(tlsL); err != nil {
			proxy.log.Errorf("couldnt serve tls: %v", err)
		} else {
			proxy.log.Infof("tls server shutdown: %v", err)
		}
	}()
	m.Serve()
	return
}

func (proxy *Proxy) Serve(listener net.Listener) error {
	tlsListener := tls.NewListener(listener, proxy.tlsCfg)
	defer tlsListener.Close()

	for {
		proxy.log.Infof("waiting for incoming TLS connections on %v", proxy.addr)
		// Accept blocks until there is an incoming TCP connection
		incoming, err := tlsListener.Accept()
		if err != nil {
			proxy.log.Errorf("couldn't accept incoming connection: %v", err)
			if cmux.ErrListenerClosed == err {
				proxy.log.Errorf("tls listener closed")
				break
			}
			continue
		}
		tlscon, ok := incoming.(*tls.Conn)
		if !ok {
			proxy.log.Errorf("no tls connection: %v", err)
			continue
		}
		proxy.log.Debug("conn: type assert to TLS succeedded")
		if err := tlscon.Handshake(); err != nil {
			proxy.log.Errorf("handshake failed: %v", err)
			continue
		}

		groups := []string{}
		names := []string{}
		types := []common.SessionType{}
		state := tlscon.ConnectionState()
		for _, vs := range state.VerifiedChains {
			for _, v := range vs {
				// we are only interested in client certificates
				// ignore the rest of the chain...
				isClient := false
				for _, eku := range v.ExtKeyUsage {
					if eku == x509.ExtKeyUsageClientAuth {
						isClient = true
					}
				}
				if !isClient {
					continue
				}
				group := fmt.Sprintf("%s/%s",
					strings.Join(v.Subject.Organization, "/"),
					strings.Join(v.Subject.OrganizationalUnit, "/"),
				)
				if len(v.Subject.CommonName) == 0 {
					proxy.log.Errorf("no commonName in certificate: %v", v.Subject.String())
					continue
				}
				for _, loc := range v.Subject.Locality {
					t, ok := proxy.typeMap[loc]
					if ok {
						types = append(types, t)
					}
				}

				groups = append(groups, group)
				names = append(names, v.Subject.CommonName)
			}
		}
		if len(names) == 0 {
			proxy.log.Error("no commonName in certificates")
			continue
		}
		sessionType := common.SessionType_Undefined
		if len(types) > 0 {
			sessionType = types[0]
		}

		session, err := yamux.Client(incoming, nil)
		if err != nil {
			proxy.log.Errorf("couldn't create yamux: %v", err)
			continue
		}

		proxy.log.Info("launching a gRPC server over incoming TCP connection")
		go func() {
			proxy.ServeSession(session, groups, names[0], sessionType)
		}()
	}
	return nil
}

func (proxy *Proxy) ServeSession(session *yamux.Session, groups []string, instancename string, sessionType common.SessionType) error {
	// using master key means, that name has to be unique
	generic := instancename == "master"
	if generic {
		instancename = session.RemoteAddr().String()
	}

	ps := NewProxySession(instancename, session, groups, sessionType, generic, proxy, proxy.log)
	defer func() {
		if err := ps.Close(); err != nil {
			proxy.log.Errorf("error closing proxy session %v: %v", ps.GetInstance(), err)
		}
	}()

	return ps.Serve()
}

func (proxy *Proxy) setVar(key string, value string) error {
	proxy.db.Lock()
	defer proxy.db.Unlock()

	if err := proxy.db.Put([]byte(key), []byte(value)); err != nil {
		return emperror.Wrapf(err, "cannot write key %s", key)
	}
	//proxy.db.Sync()
	return nil
}

func (proxy *Proxy) deleteVar(key string) error {
	proxy.db.Lock()
	defer proxy.db.Unlock()

	if err := proxy.db.Delete([]byte(key)); err != nil {
		return emperror.Wrapf(err, "cannot delete key %s", key)
	}
	//proxy.db.Sync()
	return nil
}

func (proxy *Proxy) getVar(key string) (string, error) {
	proxy.db.RLock()
	defer proxy.db.Unlock()
	val, err := proxy.db.Get([]byte(key))
	if err != nil {
		return "", emperror.Wrapf(err, "no value found for key %v", key)
	}
	return string(val), nil
}

func (proxy *Proxy) getKeys(prefix string) ([]string, error) {
	proxy.db.RLock()
	defer proxy.db.Unlock()
	keys := []string{}
	err := proxy.db.Scan([]byte(prefix), func(key []byte) error {
		keys = append(keys, string(key))
		return nil
	})
	if err != nil {
		return []string{}, emperror.Wrapf(err, "error getting keys for prefix %v", prefix)
	}
	return keys, nil
}

func (proxy *Proxy) Close() error {
	proxy.listener.Close()
	return nil
}
