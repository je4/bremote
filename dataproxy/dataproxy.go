package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/elazarl/goproxy"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/op/go-logging"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type DataProxy struct {
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
	cmuxServer    *cmux.CMux
	end           chan bool
	whitelist     []string
}

func NewDataProxy(config Config, log *logging.Logger) *DataProxy {
	dp := &DataProxy{log: log,
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
		whitelist:     []string{"*"},
	}

	return dp
}

func (dp *DataProxy) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
	if dp.caFile != "" {
		bs, err := ioutil.ReadFile(dp.caFile)
		if err != nil {
			dp.log.Panicf("error reading %v: %v", dp.caFile, err)
		}

		ok := roots.AppendCertsFromPEM(bs)
		if !ok {
			dp.log.Panicf("failed to parse root certificate:\n%v", string(bs))
		}
	}

	certificates := []tls.Certificate{}
	if dp.certFile != "" {
		cert, err := tls.LoadX509KeyPair(dp.certFile, dp.keyFile)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		// get instance name from tls certificate
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.Fatalf("server: x509.ParseCertificate: %s", err)
		}
		dp.instance = x509Cert.Subject.CommonName
		certificates = append(certificates, cert)
	}

	dp.log.Infof("trying to connect %v", dp.addr)
	dp.conn, err = tls.Dial("tcp", dp.addr, &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		ServerName:         "localhost",
		Certificates:       certificates,
	})
	if err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", dp.addr)
	}
	dp.log.Info("connection established")

	dp.session, err = yamux.Server(dp.conn, yamux.DefaultConfig())
	if err != nil {
		return emperror.Wrap(err, "cannot setup yamux dp")
	}
	dp.log.Info("yamux session established")

	return
}

func (dp *DataProxy) SetWhitelist(whitelist []string)  {
	dp.whitelist = whitelist
}

func (dp *DataProxy) Serve() error {
	waitTime := time.Second

	for {
		var wg sync.WaitGroup

		if err := dp.Connect(); err != nil {
			dp.log.Errorf("cannot connect dp %v", err)
			waitTime += time.Second
			if waitTime > time.Second*10 {
				waitTime = time.Second * 10
			}
		} else {
			waitTime = time.Second

			// we want to create different services for HTTP and GRPC (HTTP/2)

			// create a new muxer on yamux listener
			cs := cmux.New(dp.session)
			dp.cmuxServer = &cs


			grpcl := (*dp.cmuxServer).MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
			//	httpl := ps.cmuxServer.Match(cmux.Any())
//			httpl := (*dp.cmuxServer).Match(cmux.HTTP1Fast())
//			http2l := (*dp.cmuxServer).Match(cmux.HTTP2())
			datal := (*dp.cmuxServer).Match(cmux.Any())

			wg.Add(3)

			go func() {
				if err := dp.ServeGRPC(grpcl); err != nil {
					dp.log.Errorf("error serving GRPC: %v", err)
					dp.Close()
				}
				wg.Done()
			}()
			/*
			go func() {
				if err := dp.ServeProxyInt(httpl); err != nil {
					dp.log.Errorf("error serving internal HTTP: %v", err)
					dp.Close()
				}
				wg.Done()
			}()

			go func() {
				if err := dp.ServeProxyInt(http2l); err != nil {
					dp.log.Errorf("error serving internal HTTP: %v", err)
					dp.Close()
				}
				wg.Done()
			}()
			*/
			go func() {
				if err := dp.ServeProxyInt(datal); err != nil {
					dp.log.Errorf("error serving internal HTTP: %v", err)
					dp.Close()
				}
				wg.Done()
			}()
			go func() {
				if err := dp.ServeCmux(); err != nil {
					dp.log.Errorf("error serving cmux: %v", err)
					dp.Close()
				}
				wg.Done()
			}()
		}
		wg.Wait()
		dp.CloseServices()
		dp.log.Infof("sleeping %v seconds...", waitTime.Seconds())
		// wait 10 seconds or finish if needed
		select {
		case <-time.After(waitTime):
		case <-dp.end:
			dp.log.Info("shutting down")
			return nil
		}
		//time.Sleep(time.Second*10)
	}
	dp.Close()
	return nil
}

func (dp *DataProxy) ServeCmux() error {
	if err := (*dp.cmuxServer).Serve(); err != nil {
		dp.cmuxServer = nil
		return emperror.Wrap(err, "cmux closed")
	}
	dp.cmuxServer = nil
	return nil
}


func (dp *DataProxy) ServeGRPC(listener net.Listener) error {
	// create a server instance
	server := NewDataProxyServiceServer(dp, dp.log)

	// create a gRPC server object
	dp.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterDataproxyServiceServer(dp.grpcServer, &server)

	// start the gRPC erver
	dp.log.Info("launching gRPC server over TLS connection...")
	if err := dp.grpcServer.Serve(listener); err != nil {
		return emperror.Wrapf(err, "failed to serve")
	}
	return nil
}

func (dp *DataProxy) ServeProxyInt(listener net.Listener) error {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	dp.log.Info("launching http proxy server over TLS connection...")
	if err := http.Serve(listener, proxy); err != nil {
		return emperror.Wrapf(err, "failed to serve http proxy")
	}
	return nil
}



func (dp *DataProxy) CloseServices() error {
	if dp.grpcServer != nil {
		dp.grpcServer.GracefulStop()
		dp.grpcServer = nil
	}

	if dp.cmuxServer != nil {
		dp.cmuxServer = nil
	}

	if dp.session != nil {
		dp.session.Close()
		dp.session = nil
	}

	if dp.conn != nil {
		dp.conn.Close()
		dp.conn = nil
	}
	return nil
}


func (dp *DataProxy) Close() error {

	if err := dp.CloseServices(); err != nil {
		return err
	}

	// we don't close the browser if connection is closed
	/*
		if client.browser != nil {
			client.browser.Close()
			client.browser = nil
		}
	*/

	return nil
}

func (dp *DataProxy) Shutdown() error {
	err := dp.Close()
	dp.end <- true
	return err
}
