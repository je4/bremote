package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/browser"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Client struct {
	log           *logging.Logger
	instance      string
	addr          string
	httpStatic    string
	httpTemplates string
	caFile        string
	certFile      string
	keyFile       string
	httpsCertFile string
	httpsKeyFile  string
	httpsAddr     string
	conn          *tls.Conn
	session       *yamux.Session
	grpcServer    *grpc.Server
	httpServerInt *http.Server
	httpServerExt *http.Server
	cmuxServer    *cmux.CMux
	end           chan bool
	browser       *browser.Browser
	status        string
}

func NewClient(config Config, log *logging.Logger) *Client {
	client := &Client{log: log,
		instance:      config.InstanceName,
		addr:          config.Proxy,
		httpStatic:    config.HttpStatic,
		httpTemplates: config.HttpTemplates,
		caFile:        config.CaPEM,
		certFile:      config.CertPEM,
		keyFile:       config.KeyPEM,
		httpsCertFile: config.HttpsCertPEM,
		httpsKeyFile:  config.HttpsKeyPEM,
		httpsAddr:     config.HttpsAddr,
		end:           make(chan bool, 1),
		status:        common.ClientStatus_Empty,
	}

	return client
}

func (client *Client) SetBrowser(browser *browser.Browser) error {
	if client.browser != nil {
		return errors.New("browser already exists")
	}
	client.browser = browser
	return nil
}

func (client *Client) SetStatus(status string) {
	client.status = status
}

func (client *Client) GetStatus() string {
	return client.status
}

func (client *Client) GetInstance() string {
	return client.instance
}

func (client *Client) ShutdownBrowser() error {
	if client.browser == nil {
		return errors.New("no browser available")
	}
	client.browser.Close()
	client.browser = nil
	return nil
}

func (client *Client) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	if client.caFile != "" {
		bs, err := ioutil.ReadFile(client.caFile)
		if err != nil {
			client.log.Panicf("error reading %v: %v", client.caFile, err)
		}

		ok := roots.AppendCertsFromPEM(bs)
		if !ok {
			client.log.Panicf("failed to parse root certificate:\n%v", string(bs))
		}
	}

	certificates := []tls.Certificate{}
	if client.certFile != "" {
		cert, err := tls.LoadX509KeyPair(client.certFile, client.keyFile)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		certificates = append(certificates, cert)
	}

	client.log.Infof("trying to connect %v", client.addr)
	client.conn, err = tls.Dial("tcp", client.addr, &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		ServerName:         "localhost",
		Certificates:       certificates,
	})
	if err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", client.addr)
	}
	client.log.Info("connection established")

	client.session, err = yamux.Server(client.conn, yamux.DefaultConfig())
	if err != nil {
		return emperror.Wrap(err, "cannot setup yamux client")
	}
	client.log.Info("yamux session established")

	return
}

func (client *Client) Serve() error {
	waitTime := time.Second

	for {
		var wg sync.WaitGroup

		if err := client.Connect(); err != nil {
			client.log.Errorf("cannot connect client %v", err)
			waitTime += time.Second
			if waitTime > time.Second*10 {
				waitTime = time.Second * 10
			}
		} else {
			waitTime = time.Second

			// we want to create different services for HTTP and GRPC (HTTP/2)

			// create a new muxer on yamux listener
			cs := cmux.New(client.session)
			client.cmuxServer = &cs

			// first get http1
			httpl := (*client.cmuxServer).Match(cmux.HTTP1Fast())
			// the rest should be grpc
			grpcl := (*client.cmuxServer).Match(cmux.Any())


			wg.Add(4)

			go func() {
				if err := client.ServeGRPC(grpcl); err != nil {
					client.log.Errorf("error serving GRPC: %v", err)
					client.Close()
				}
				wg.Done()
			}()
				go func() {
					if err := client.ServeHTTPInt(httpl); err != nil {
						client.log.Errorf("error serving internal HTTP: %v", err)
						client.Close()
					}
					wg.Done()
				}()

			go func() {
				if err := client.ServeHTTPExt(); err != nil {
					client.log.Errorf("error serving external HTTP: %v", err)
					client.Close()
				}
				wg.Done()
			}()

			go func() {
				if err := client.ServeCmux(); err != nil {
					client.log.Errorf("error serving cmux: %v", err)
					client.Close()
				}
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Second * 2)
				if err := client.InitProxy(); err != nil {
					log.Panicf("cannot initialize proxy: %+v", err)
					client.CloseServices()
				}
			}()
		}
		wg.Wait()
		client.CloseServices()
		client.log.Infof("sleeping %v seconds...", waitTime.Seconds())
		// wait 10 seconds or finish if needed
		select {
		case <-time.After(waitTime):
		case <-client.end:
			client.log.Info("shutting down")
			return nil
		}
		//time.Sleep(time.Second*10)
	}
	client.Close()
	return nil
}

func (client *Client) InitProxy() error {

	pw := pb.NewProxyWrapper(client.instance, &client.session)

	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	if err := pw.Init(traceId, client.instance, common.SessionType_Client, client.GetStatus()); err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (client *Client) ServeGRPC(listener net.Listener) error {
	// create a server instance
	server := NewClientServiceServer(client, client.log)

	// create a gRPC server object
	client.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterClientServiceServer(client.grpcServer, &server)

	// start the gRPC erver
	client.log.Info("launching gRPC server over TLS connection...")
	if err := client.grpcServer.Serve(listener); err != nil {
		return emperror.Wrapf(err, "failed to serve")
	}
	return nil
}

func (client *Client) ServeHTTPExt() error {
	httpservmux := http.NewServeMux()

	// static files only from /static
	fs := http.FileServer(http.Dir(client.httpStatic))
	httpservmux.Handle("/static/", http.StripPrefix("/static/", fs))

	// the proxy
	// ignore error because of static url, which must be correct
	rpURL, _ := url.Parse("http://localhost:80/")
	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if client.session == nil {
				return nil, errors.New("no tls session available")
			}
			return client.session.Open()
		},
	}
	httpservmux.Handle("/", proxy)

	client.httpServerExt = &http.Server{Addr: client.httpsAddr, Handler: httpservmux}

	client.log.Infof("launching external HTTPS on %s", client.httpsAddr)
	err := client.httpServerExt.ListenAndServeTLS(client.httpsCertFile, client.httpsKeyFile)
	if err != nil {
		client.httpServerExt = nil
		return emperror.Wrapf(err, "failed to serve")
	}
	client.httpServerExt = nil
	return nil
}

func (client *Client) ServeHTTPInt(listener net.Listener) error {
	httpservmux := http.NewServeMux()
	// static files only from /static
	fs := http.FileServer(http.Dir(client.httpStatic))
	httpservmux.Handle("/static/", http.StripPrefix("/static/", fs))

	client.httpServerInt = &http.Server{Addr: ":80", Handler: httpservmux}

	client.log.Info("launching HTTP server over TLS connection...")
	// starting http server
	if err := client.httpServerInt.Serve(listener); err != nil {
		client.httpServerInt = nil
		return emperror.Wrapf(err, "failed to serve")
	}

	client.httpServerInt = nil
	return nil
}

func (client *Client) ServeCmux() error {
	if err := (*client.cmuxServer).Serve(); err != nil {
		client.cmuxServer = nil
		return emperror.Wrap(err, "cmux closed")
	}
	client.cmuxServer = nil
	return nil
}


func (client *Client) CloseServices() error {
	if client.grpcServer != nil {
		client.grpcServer.GracefulStop()
		client.grpcServer = nil
	}

	if client.httpServerInt != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		client.httpServerInt.Shutdown(ctx)
		client.httpServerInt = nil
	}

	if client.httpServerExt != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		client.httpServerExt.Shutdown(ctx)
		client.httpServerExt = nil
	}

	if client.cmuxServer != nil {
		client.cmuxServer = nil
	}

	if client.session != nil {
		client.session.Close()
		client.session = nil
	}

	if client.conn != nil {
		client.conn.Close()
		client.conn = nil
	}
	return nil
}

func (client *Client) Close() error {

	if err := client.CloseServices(); err != nil {
		return err
	}

	if client.browser != nil {
		client.browser.Close()
		client.browser = nil
	}

	return nil
}

func (client *Client) Shutdown() error {
	err := client.Close()
	client.end <- true
	return err
}
