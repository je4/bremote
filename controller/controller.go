package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/goph/emperror"
	"github.com/gorilla/mux"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Controller struct {
	log                *logging.Logger
	instance           string
	addr               string
	httpsAddr          string
	caFile             string
	certFile           string
	keyFile            string
	httpsCertFile      string
	httpsKeyFile       string
	httpStatic         string
	httpTemplates      string
	httpTemplateCache  bool
	conn               *tls.Conn
	session            *yamux.Session
	grpcServer         *grpc.Server
	httpServerInt      *http.Server
	httpServerExt      *http.Server
	cmuxServer         *cmux.CMux
	end                chan bool
	roots              *x509.CertPool
	certificates       *[]tls.Certificate
	kvs                map[string]interface{} // key value store of controller
	templateCache      gcache.Cache
	templateDelimLeft  string
	templateDelimRight string
	templatesInternal  map[string]string
}

func NewController(config Config, log *logging.Logger) *Controller {
	controller := &Controller{log: log,
		instance:           config.InstanceName,
		addr:               config.Proxy,
		httpsCertFile:      config.HttpsCertPEM,
		httpsKeyFile:       config.HttpsKeyPEM,
		httpsAddr:          config.HttpsAddr,
		httpStatic:         config.HttpStatic,
		httpTemplates:      config.Templates.Folder,
		httpTemplateCache:  config.Templates.Cache,
		caFile:             config.CaPEM,
		certFile:           config.CertPEM,
		keyFile:            config.KeyPEM,
		end:                make(chan bool, 1),
		kvs:                make(map[string]interface{}),
		templateCache:      gcache.New(100).LRU().Build(),
		templateDelimLeft:  config.Templates.DelimLeft,
		templateDelimRight: config.Templates.DelimRight,
		templatesInternal:  make(map[string]string),
	}
	for _, internal := range config.Templates.Internal {
		controller.templatesInternal[internal.Name] = internal.File
	}
	return controller
}

func (controller *Controller) GetInstance() string {
	return controller.instance
}

func (controller *Controller) GetSessionPtr() **yamux.Session {
	return &controller.session
}

func (controller *Controller) SetVar(client string, key string, value interface{}) error {
	pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
	traceId := uniqid.New(uniqid.Params{"traceid_", false})

	json, err := json.Marshal(value)
	if err != nil {
		return emperror.Wrapf(err, "cannot encode data: %v", err)
	}

	err = pw.KVStoreSetValue(client, key, string(json), traceId)
	if err != nil {
		return emperror.Wrapf(err, "cannot get value for %s", key)
	}
	return nil
}

func (controller *Controller) DeleteVar(client string, key string) error {
	pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	err := pw.KVStoreDeleteValue(client, key, traceId)
	if err != nil {
		return emperror.Wrapf(err, "cannot get value for %s", key)
	}
	return nil
}

func (controller *Controller) GetVar(client string, key string) (interface{}, error) {

	pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	value, err := pw.KVStoreGetValue(client, key, traceId)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot get value for %s", key)
	}
	var data interface{}
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot decode data: %v", value)
	}

	return data, nil
}

func (controller *Controller) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	if controller.roots == nil {
		controller.roots = x509.NewCertPool()
		// todo: get rid of this...in a safe way


		rootCert, err := ioutil.ReadFile(controller.httpsCertFile)
		if err != nil {
			controller.log.Panicf("error reading root certificate %v: %v", controller.httpsCertFile, err)
		}
		ok := controller.roots.AppendCertsFromPEM(rootCert)
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

			// we want to create different services for HTTP and GRPC (HTTP/2)

			// create a new muxer on yamux listener
			cs := cmux.New(controller.session)
			controller.cmuxServer = &cs

			// first get http1
			httpl := (*controller.cmuxServer).Match(cmux.HTTP1Fast())
			// the rest should be grpc
			grpcl := (*controller.cmuxServer).Match(cmux.Any())

			// starting 2 services
			wg.Add(4)

			go func() {
				if err := controller.ServeGRPC(grpcl); err != nil {
					controller.log.Errorf("error serving GRPC: %v", err)
					controller.Close()
				}
				wg.Done()
			}()
			go func() {
				if err := controller.ServeHTTPInt(httpl); err != nil {
					controller.log.Errorf("error serving HTTP: %v", err)
					controller.Close()
				}
				wg.Done()
			}()

			go func() {
				if err := controller.ServeHTTPExt(); err != nil {
					controller.log.Errorf("error serving external HTTP: %v", err)
					controller.Close()
				}
				wg.Done()
			}()

			go func() {
				if err := controller.ServeCmux(); err != nil {
					controller.log.Errorf("error serving cmux: %v", err)
					controller.Close()
				}
				wg.Done()
			}()

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
	if err := pw.Init(traceId, controller.instance, common.SessionType_Controller, common.ClientStatus_Empty, controller.httpsAddr); err != nil {
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

func (controller *Controller) ServeCmux() error {
	if err := (*controller.cmuxServer).Serve(); err != nil {
		controller.cmuxServer = nil
		return emperror.Wrap(err, "cmux closed")
	}
	controller.cmuxServer = nil
	return nil
}

func (controller *Controller) ServeHTTPExt() error {
	r := mux.NewRouter()
	controller.addRestRoutes(r)

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Do stuff here
			controller.log.Debugf(r.RequestURI)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	})

	controller.httpServerExt = &http.Server{Addr: controller.httpsAddr, Handler: r}

	controller.log.Infof("launching external HTTPS on %s", controller.httpsAddr)
	err = controller.httpServerExt.ListenAndServeTLS(controller.httpsCertFile, controller.httpsKeyFile)
	if err != nil {
		controller.httpServerExt = nil
		return emperror.Wrapf(err, "failed to serve")
	}
	controller.httpServerExt = nil
	return nil
}

func (controller *Controller) templateHandler() func(w http.ResponseWriter, r *http.Request) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		// get rid of /template and prefix with httpFolder
		str := filepath.Clean(r.URL.Path)
		str = strings.TrimPrefix(str, string(os.PathSeparator)+controller.GetInstance())
		str = strings.TrimPrefix(str, string(os.PathSeparator)+"templates")
		file := filepath.Join(controller.httpTemplates, str)

		// get source instance
		source := r.Header.Get("X-Source-Instance")

		v, err := controller.GetVar(source, filepath.Base(file))
		if err != nil {
			controller.log.Errorf("cannot execute template without data: %v", err)
			http.Error(w, fmt.Sprintf("cannot execute template without data: %v", err), http.StatusNotFound)
			return
		}
		data := v.(map[string]interface{})

		var tmpl *template.Template
		key := fmt.Sprintf("%s-%s", source, filepath.Base(file))
		h, err := controller.templateCache.Get(key)
		if err != nil {
			tpldata, err := ioutil.ReadFile(file)
			if err != nil {
				controller.log.Errorf("error reading %v: %v", file, err)
				http.Error(w, fmt.Sprintf("error reading %v: %v", file, err), http.StatusNotFound)
				return
			}
			tmpl, err = template.New(file).Delims(controller.templateDelimLeft, controller.templateDelimRight).Parse(string(tpldata)) //ParseFiles(file)
			if err != nil {
				controller.log.Errorf("error in template %v: %v", file, err)
				http.Error(w, fmt.Sprintf("error in template %v: %v", file, err), http.StatusUnprocessableEntity)
				return
			}
			// add to cache only if enabled
			if controller.httpTemplateCache {
				controller.templateCache.Set(key, tmpl)
			}
		} else {
			tmpl = h.(*template.Template)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			controller.log.Errorf("error in data for template %v: %v", file, err)
			http.Error(w, fmt.Sprintf("error in data for template %v: %v", file, err), http.StatusUnprocessableEntity)
			return
		}
	}
	return hf
}

func (controller *Controller) ServeHTTPInt(listener net.Listener) error {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir(controller.httpStatic))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	//pattern := fmt.Sprintf("/%s/static/", controller.GetInstance())
	//r.Handle(pattern, http.StripPrefix(pattern, fs))

	controller.addRestRoutes(r)

	r.Use(controller.RestLogger())

	//	pattern = fmt.Sprintf("/%s/templates/", controller.GetInstance())
	r.PathPrefix("/templates/").HandlerFunc(controller.templateHandler())

	_ = func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			controller.log.Infof("[%s] %s %s %s\n", controller.GetInstance(), r.RemoteAddr, r.Method, r.URL)
			handler.ServeHTTP(w, r)
		})
	}

	r.Use(controller.RestLogger())

	controller.httpServerInt = &http.Server{Addr: "localhost:80", Handler: r}

	controller.log.Info("launching HTTP server over TLS connection...")
	// starting http server
	if err := controller.httpServerInt.Serve(listener); err != nil {
		controller.httpServerInt = nil
		return emperror.Wrapf(err, "failed to serve")
	}

	controller.httpServerInt = nil

	return nil
}

func (controller *Controller) ServeGRPC(listener net.Listener) error {
	// create a server instance
	server := NewControllerServiceServer(controller, controller.log)

	// create a gRPC server object
	controller.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterControllerServiceServer(controller.grpcServer, &server)

	// start the gRPC erver
	controller.log.Info("launching gRPC server over TLS connection...")
	if err := controller.grpcServer.Serve(listener); err != nil {
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

	if controller.httpServerExt != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		controller.httpServerExt.Shutdown(ctx)
		controller.httpServerExt = nil
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
