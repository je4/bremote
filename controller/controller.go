package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Controller struct {
	log           *logging.Logger
	instance      string
	addr          string
	httpStatic    string
	httpTemplates string
	caFile        string
	certFile      string
	keyFile       string
	conn          *tls.Conn
	session       *yamux.Session
	grpcServer    *grpc.Server
	httpServerInt    *http.Server
	end           chan bool
	roots         *x509.CertPool
	certificates  *[]tls.Certificate
	kvs           map[string]interface{} // key value store of controller
	templateCache gcache.Cache
}

func NewController(config Config, log *logging.Logger) *Controller {
	controller := &Controller{log: log,
		instance:      config.InstanceName,
		addr:          config.Proxy,
		httpStatic:    config.HttpStatic,
		httpTemplates: config.HttpTemplates,
		caFile:        config.CaPEM,
		certFile:      config.CertPEM,
		keyFile:       config.KeyPEM,
		end:           make(chan bool, 1),
		kvs:           make(map[string]interface{}),
		templateCache: gcache.New(100).LRU().Build(),
	}
	return controller
}

func (controller *Controller) GetInstance() string {
	return controller.instance
}

func (controller *Controller) GetSessionPtr() **yamux.Session {
	return &controller.session
}

func (controller *Controller) SetVar(key string, value interface{}) {
	controller.kvs[key] = value
}

func (controller *Controller) GetVar(key string) (interface{}, error) {
	ret, ok := controller.kvs[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("no value found for key %v", key))
	}
	return ret, nil
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

	if controller.certificates == nil {
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

			// starting 2 services
			wg.Add(1)

			go func() {
				if err := controller.ServeGRPC(); err != nil {
					controller.log.Errorf("error serving GRPC: %v", err)
					controller.Close()
				}
				wg.Done()
			}()
			if false {
				go func() {
					time.Sleep(time.Second * 1)
					if err := controller.ServeHTTPInt(); err != nil {
						controller.log.Errorf("error serving HTTP: %v", err)
						controller.Close()
					}
					wg.Done()
				}()
			}

			go func() {
				time.Sleep(time.Second * 2)
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

func (controller *Controller) ServeHTTPInt() error {
	httpservmux := http.NewServeMux()
	// static files only from /static
	fs := http.FileServer(http.Dir(controller.httpStatic))
	httpservmux.Handle("/static/", http.StripPrefix("/static/", fs))
	httpservmux.HandleFunc("/template/", func(w http.ResponseWriter, r *http.Request) {
		// get rid of /template and prefix with httpFolder
		file := filepath.Join(controller.httpTemplates, strings.TrimPrefix("/template", filepath.Clean(r.URL.Path)))

		// get source instance
		source := r.Header.Get("sourceinstance")

		key := fmt.Sprintf("%s-%s", source, filepath.Base(file))
		data, err := controller.GetVar(key)
		if err != nil {
			controller.log.Errorf("cannot execute template without data: %v", err)
			http.Error(w, fmt.Sprintf("cannot execute template without data: %v", err), http.StatusNotFound)
			return
		}
		var tmpl *template.Template
		h, err := controller.templateCache.Get(key)
		if err != nil {
			tmpl, err = template.ParseFiles(file)
			if err != nil {
				controller.log.Errorf("error in template %v: %v", file, err)
				http.Error(w, fmt.Sprintf("error in template %v: %v", file, err), http.StatusUnprocessableEntity)
				return
			}
			controller.templateCache.Set(key, tmpl)
		} else {
			tmpl = h.(*template.Template)
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			controller.log.Errorf("error in data for template %v: %v", file, err)
			http.Error(w, fmt.Sprintf("error in data for template %v: %v", file, err), http.StatusUnprocessableEntity)
			return
		}
	})

	controller.httpServerInt = &http.Server{Addr: "localhost:80", Handler: httpservmux}

	controller.log.Info("launching HTTP server over TLS connection...")
	// starting http server
	if err := controller.httpServerInt.Serve(controller.session); err != nil {
		controller.httpServerInt = nil
		return emperror.Wrapf(err, "failed to serve")
	}

	controller.httpServerInt = nil

	return nil
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
	if controller.httpServerInt != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		controller.httpServerInt.Shutdown(ctx)
		controller.httpServerInt = nil
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
