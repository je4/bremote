package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
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
	if err := pw.Init(traceId, controller.instance, common.SessionType_Controller, common.ClientStatus_Empty); err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (controller *Controller) GetClients() ([]common.ClientInfo, error) {
	pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	clients, err := pw.GetClients(traceId, common.SessionType_Client, true)
	if err != nil {
		return []common.ClientInfo{}, emperror.Wrap(err, "cannot get clients")
	}
	return clients, nil
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
