package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type Controller struct {
	log        *logging.Logger
	instance   string
	addr string
	caFile     string
	certFile   string
	keyFile    string
	conn       *tls.Conn
	session    *yamux.Session
	grpcServer *grpc.Server
	end        chan bool
	roots *x509.CertPool
	certificates *[]tls.Certificate
}

func NewController(instance string, addr string, caFile string, certFile string, keyFile string, log *logging.Logger) *Controller {
	controller := &Controller{log: log,
		instance: instance,
		addr:addr,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
		end: make(chan bool, 1),
	}
	return controller
}

func (controller *Controller) GetInstance() (string) {
	return controller.instance
}

func (controller *Controller) GetSessionPtr() (**yamux.Session) {
	return &controller.session
}


func (controller *Controller) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	if controller.roots == nil {
		controller.roots = x509.NewCertPool()
		ok := controller.roots.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			panic("failed to parse root certificate")
		}
		if controller.caFile != "" {
			bs, err := ioutil.ReadFile(controller.caFile)
			if err != nil {
				controller.log.Panicf("error reading %v: %v", controller.caFile, err)
			}

			ok := controller.roots.AppendCertsFromPEM(bs)
			if !ok {
				controller.log.Panicf("failed to parse root certificate:\n%v", string(bs))
			}
		}
	}

	if controller.certificates ==  nil {
		certificates := []tls.Certificate{}
		if controller.certFile != "" {
			cert, err := tls.LoadX509KeyPair(controller.certFile, controller.keyFile)
			if err != nil {
				log.Fatalf("server: loadkeys: %s", err)
			}
			certificates = append(certificates, cert)
		}
		controller.certificates = &certificates
	}
	controller.log.Infof("trying to connect %v", controller.addr)
	controller.conn, err = tls.Dial("tcp", controller.addr, &tls.Config{
		RootCAs:            controller.roots,
		InsecureSkipVerify: true,
		ServerName:         "localhost",
		Certificates:       *controller.certificates,
	})
	if err != nil {
		controller.conn = nil
		return emperror.Wrapf(err, "cannot connect to %v", controller.addr)
	}
	controller.log.Info("connection established")

	controller.session, err = yamux.Server(controller.conn, yamux.DefaultConfig())
	if err != nil {
		controller.session = nil
		return emperror.Wrap(err, "cannot setup yamux controller")
	}
	controller.log.Info("yamux session established")

	return
}

func (controller *Controller) Serve() error {
	waitTime := time.Second

	for {
		var wg sync.WaitGroup

		if err := controller.Connect(); err != nil {
			controller.log.Errorf("cannot connect controller %v", err)
			waitTime += time.Second
			if waitTime > time.Second*10 {
				waitTime = time.Second * 10
			}
		} else {
			waitTime = time.Second

			wg.Add(1)

			go func() {
				if err := controller.ServeGRPC(); err != nil {
					controller.log.Errorf("error serving GRPC: %v", err)
					controller.Close()
				}
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Second * 1)
				if err := controller.InitProxy(); err != nil {
					controller.log.Errorf("cannot initialize proxy: %+v", err)
					controller.Close()
				}
			}()
		}
		wg.Wait()
		controller.Close()
		controller.log.Infof("sleeping %v seconds...", waitTime.Seconds())
		// wait 10 seconds or finish if needed
		select {
		case <-time.After(waitTime):
		case <-controller.end:
			controller.log.Info("shutting down")
			return nil
		}
		//time.Sleep(time.Second*10)
	}
	return nil
}

func (controller *Controller) InitProxy() error {

	pw := pb.NewProxyWrapper(controller.instance, &controller.session)

	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	if err := pw.Init(traceId, controller.instance, common.SessionType_Controller); err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (controller *Controller) GetClients() ([]string, error) {
	// gRPC dial over incoming net.Conn
	conn, err := grpc.Dial(":7777", grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			if controller.session == nil {
				return nil, errors.New(fmt.Sprintf("session %s closed", s))
			}
			return controller.session.Open()
		}),
	)
	if err != nil {
		return []string{}, errors.New("cannot dial grpc connection to :7777")
	}
	proxy := pb.NewProxyServiceClient(conn)
	clients, err := proxy.GetClients(context.Background(), &empty.Empty{})
	if err != nil {
		return []string{}, emperror.Wrap(err, "cannot get clients")
	}
	ret := []string{}
	for _, c := range clients.GetClients() {
		// we only want to get clients
		if c.GetType() != pb.ProxySessionType_Client {
			continue
		}
		ret = append(ret, c.GetInstance())
	}
	return ret, nil
}


func (controller *Controller) ServeGRPC() error {
	// create a server instance
	server := NewControllerServiceServer(controller, controller.log)

	// create a gRPC server object
	controller.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterControllerServiceServer(controller.grpcServer, &server)

	// start the gRPC erver
	controller.log.Info("launching gRPC server over TLS connection...")
	if err := controller.grpcServer.Serve(controller.session); err != nil {
		controller.grpcServer = nil
		return emperror.Wrapf(err, "failed to serve")
	}
	controller.grpcServer = nil
	return nil
}

func (controller *Controller) Close() error {
	if controller.grpcServer != nil {
		controller.grpcServer.GracefulStop()
		controller.grpcServer = nil
	}

	if controller.session != nil {
		controller.session.Close()
		controller.session = nil
	}

	if controller.conn != nil {
		controller.conn.Close()
		controller.conn = nil
	}

	return nil
}

func (controller *Controller) Shutdown() error {
	err := controller.Close()
	controller.end <- true
	return err
}
