package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "info_age.net/bremote/api"
	"info_age.net/bremote/browser"
	"info_age.net/bremote/common"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type Client struct {
	log        *logging.Logger
	instance   string
	caFile     string
	certFile   string
	keyFile    string
	conn       *tls.Conn
	session    *yamux.Session
	grpcServer *grpc.Server
	end        chan bool
	browser    *browser.Browser
}

func NewClient(instance string, caFile string, certFile string, keyFile string, log *logging.Logger) *Client {
	client := &Client{log: log,
		instance: instance,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
		end:      make(chan bool, 1),
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

	client.log.Infof("trying to connect %v", addr)
	client.conn, err = tls.Dial("tcp", addr, &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		ServerName:         "localhost",
		Certificates:       certificates,
	})
	if err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", addr)
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

			wg.Add(1)

			go func() {
				if err := client.ServeGRPC(); err != nil {
					client.log.Errorf("error serving GRPC: %v", err)
				}
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Second * 1)
				if err := client.InitProxy(); err != nil {
					log.Panicf("cannot initialize proxy: %+v", err)
					client.Close()
				}
			}()
		}
		wg.Wait()
		client.Close()
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
	return nil
}

func (client *Client) InitProxy() error {

	pw := pb.NewProxyWrapper(client.instance, &client.session)

	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	if err := pw.Init(traceId, client.instance, common.SessionType_Client); err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (client *Client) ServeGRPC() error {
	// create a server instance
	server := NewClientServiceServer(client, client.log)

	// create a gRPC server object
	client.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterClientServiceServer(client.grpcServer, &server)

	// start the gRPC erver
	client.log.Info("launching gRPC server over TLS connection...")
	if err := client.grpcServer.Serve(client.session); err != nil {
		return emperror.Wrapf(err, "failed to serve")
	}
	return nil
}

func (client *Client) Close() error {
	if client.grpcServer != nil {
		client.grpcServer.GracefulStop()
		client.grpcServer = nil
	}

	if client.session != nil {
		client.session.Close()
		client.session = nil
	}

	if client.conn != nil {
		client.conn.Close()
		client.conn = nil
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
