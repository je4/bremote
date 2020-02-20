package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/gorilla/mux"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/browser"
	"github.com/je4/bremote/common"
	"github.com/je4/ntp"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"github.com/sahmad98/go-ringbuffer"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type BrowserClient struct {
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
	httpProxy     string
	conn          *tls.Conn
	session       *yamux.Session
	grpcServer    *grpc.Server
	httpServerInt *http.Server
	httpServerExt *http.Server
	cmuxServer    *cmux.CMux
	end           chan bool
	browser       *browser.Browser
	status        string
	wsGroup       map[string]*ClientWebsocket
	browserLog    *ringbuffer.RingBuffer
	fw            *common.TCPForwarder
}

func NewClient(config Config, log *logging.Logger) *BrowserClient {
	client := &BrowserClient{log: log,
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
		httpProxy:     config.HttpProxy,
		end:           make(chan bool, 1),
		status:        common.ClientStatus_Empty,
		wsGroup:       make(map[string]*ClientWebsocket),
		browserLog:    ringbuffer.NewRingBuffer(100),
		fw:            nil,
	}

	return client
}

func (client *BrowserClient) writeBrowserLog(format string, a ...interface{}) {
	client.browserLog.Write(fmt.Sprintf(format, a...))
}

func (client *BrowserClient) getBrowserLog() []string {
	result := []string{}
	client.browserLog.Reader = client.browserLog.Writer
	for i := 0; i < client.browserLog.Size; i++ {
		elem := client.browserLog.Read()
		str, ok := elem.(string)
		if !ok {
			continue
		}
		result = append(result, str)
	}
	return result
}

func (client *BrowserClient) SetBrowser(browser *browser.Browser) error {
	if client.browser != nil {
		return errors.New("browser already exists")
	}
	client.browser = browser
	return nil
}

func (client *BrowserClient) SetGroupWebsocket(group string, ws *ClientWebsocket) {
	client.wsGroup[group] = ws
}

func (client *BrowserClient) DeleteGroupWebsocket(group string) {
	delete(client.wsGroup, group)
}

func (client *BrowserClient) GetGroupWebsocket(group string) (*ClientWebsocket, error) {
	ws, ok := client.wsGroup[group]
	if !ok {
		return nil, errors.New(fmt.Sprintf("no websocket connection for group %v", group))
	}
	return ws, nil

}

func (client *BrowserClient) SendGroupWebsocket(group string, message []byte) error {
	ws, err := client.GetGroupWebsocket(group)
	if err != nil {
		return emperror.Wrapf(err, "cannot send to group %v", group)
	}
	ws.send <- message
	return nil

}

func (client *BrowserClient) GetBrowser() (*browser.Browser, error) {
	if client.browser == nil {
		return nil, errors.New("browser not initialized")
	}
	return client.browser, nil
}

func (client *BrowserClient) SetStatus(status string) {
	client.status = status
}

func (client *BrowserClient) GetHTTPAddr() string {
	return client.httpsAddr
}

func (client *BrowserClient) GetStatus() string {
	if client.status != "" {
		if client.browser == nil {
			client.status = ""
		} else {
			if !client.browser.IsRunning() {
				client.status = ""
			}
		}
	}
	return client.status
}

func (client *BrowserClient) GetInstance() string {
	return client.instance
}

func (client *BrowserClient) GetSessionPtr() **yamux.Session {
	return &client.session
}

func (client *BrowserClient) ShutdownBrowser() error {
	if client.browser == nil {
		return errors.New("no browser available")
	}
	client.browser.Close()
	client.browser = nil
	return nil
}

func (client *BrowserClient) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
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
		// get instance name from tls certificate
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.Fatalf("server: x509.ParseCertificate: %s", err)
		}
		client.instance = x509Cert.Subject.CommonName
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

func (client *BrowserClient) ServeExternal() error {
	//			wg.Add(1)
	go func() {
		if err := client.ServeHTTPExt(); err != nil {
			client.log.Errorf("error serving external HTTP: %v", err)
			client.CloseInternal()
		}
		//				wg.Done()
	}()

	//			wg.Add(1)
	go func() {
		if err := client.ServeHTTPProxy(); err != nil {
			client.log.Errorf("error serving http proxy: %v", err)
			client.CloseInternal()
		}
		//				wg.Done()
	}()
	return nil
}

func (client *BrowserClient) ServeInternal() error {
	waitTime := time.Second

	for {
		var wg sync.WaitGroup

		if client == nil {
			return errors.New("client is nil")
		}

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

			//			wg.Add(1)
			go func() {
				if err := client.ServeGRPC(grpcl); err != nil {
					client.log.Errorf("error serving GRPC: %v", err)
					client.CloseInternal()
				}
				//				wg.Done()
			}()
			//			wg.Add(1)
			go func() {
				if err := client.ServeHTTPInt(httpl); err != nil {
					client.log.Errorf("error serving internal HTTP: %v", err)
					client.CloseInternal()
				}
				//				wg.Done()
			}()


			wg.Add(1)
			go func() {
				if err := client.ServeCmux(); err != nil {
					client.log.Errorf("error serving cmux: %v", err)
					client.CloseInternal()
				}
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Second * 2)
				if err := client.InitProxy(); err != nil {
					client.log.Errorf("cannot initialize proxy: %+v", err)
					//client.CloseExternal()
				}
			}()
		}
		wg.Wait()
		client.CloseInternal()
		client.log.Infof("sleeping %v seconds...", waitTime.Seconds())
		// wait 10 seconds or finish if needed
		select {
		case <-time.After(waitTime):
		case <-client.end:
			client.log.Info("shutting down")
			return nil
		}
		//time.Sleep(time.Second*3)
	}
	client.CloseInternal()
	return nil
}

func (client *BrowserClient) InitProxy() error {
	if client == nil {
		return errors.New("client is nil")
	}
	pw := pb.NewProxyWrapper(client.instance, &client.session)

	traceId := uniqid.New(uniqid.Params{"traceid_", false})
	if err := pw.Init(traceId, client.instance, common.SessionType_Client, client.GetStatus(), client.httpsAddr); err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (client *BrowserClient) ServeGRPC(listener net.Listener) error {
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

func (client *BrowserClient) getProxyDirector() func(req *http.Request) {

	target, _ := url.Parse("http://localhost:80/")
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		//		vars := mux.Vars(req)
		//		t := vars["target"]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = common.SingleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("X-Source-Instance", client.GetInstance())
		//		req.Header.Set("Access-Control-Allow-Origin", "*")
	}

	return director
}

func (client *BrowserClient) ntpHandlerFunc() func(data []byte) ([]byte, error) {
	return func(data []byte) ([]byte, error) {
		pw := pb.NewProxyWrapper(client.instance, client.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		ret, err := pw.NTPRaw(traceId, data)
		if err != nil {
			return nil, emperror.Wrapf(err, "get ntp data from proxy")
		}
		return ret, nil
	}
}

func (client *BrowserClient) ntp() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client.log.Info("ntp()")
		ret, err := ntp.Query(client.ntpHandlerFunc())
		if err != nil {
			client.log.Errorf("cannot query ntp service: %v", err)
			http.Error(w, fmt.Sprintf("cannot query ntp service: %v", err), http.StatusInternalServerError)
		}
		result := struct {
			Time        time.Time
			ClockOffset int64
		}{
			Time:        ret.Time.Local(),
			ClockOffset: int64(ret.ClockOffset / time.Millisecond),
		}
		jsonstr, err := json.Marshal(result)
		if err != nil {
			client.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(jsonstr))

	}
}

func (client *BrowserClient) browserClick() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client.log.Info("browserClick()")
		xstr := r.FormValue("x")
		ystr := r.FormValue("y")
		x, err := strconv.ParseInt(xstr, 10, 64)
		if err != nil {
			client.log.Errorf("cannot parse x %v: %v", xstr, err)
			http.Error(w, fmt.Sprintf("cannot parse x %v: %v", xstr, err), http.StatusBadRequest)
		}
		y, err := strconv.ParseInt(ystr, 10, 64)
		if err != nil {
			client.log.Errorf("cannot parse x %v: %v", ystr, err)
			http.Error(w, fmt.Sprintf("cannot parse y %v: %v", ystr, err), http.StatusBadRequest)
		}
		if err := client.browser.MouseClickXY(x, y); err != nil {
			client.log.Errorf("cannot click %v/%v: %v", x, y, err)
			http.Error(w, fmt.Sprintf("cannot click %v/%v: %v", x, y, err), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `"status":"ok"`)
	}
}

type MyTransport http.Transport

func (transport *MyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Make the request to the server.
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	resp.Header.Set("Access-Control-Allow-Origin", "*")
	return resp, nil
}

func (client *BrowserClient) ServeHTTPProxy() error {
	if client.httpProxy == "" {
		return nil
	}
	listener, err := net.Listen("tcp", client.httpProxy)
	if err != nil {
		return emperror.Wrapf(err, "cannot dial to %v", client.httpProxy)
	}

	// static session in client
	getTargetConn := func() (net.Conn, error) {
		if client.session == nil {
			return nil, errors.New("no active session")
		}
		return client.session.Open()
	}

	client.fw = common.NewTCPForwarder("[proxy]", 0, getTargetConn, client.log)
	client.log.Infof("launching external HTTP proxy on %s", client.httpProxy)
	if err := client.fw.Serve(listener); err != nil {
		return emperror.Wrapf(err, "error listening on %v", client.httpProxy)
	}
	return nil
}

func (client *BrowserClient) ServeHTTPExt() (err error) {
	r := mux.NewRouter()

	// static files only from /static
	fs := http.FileServer(http.Dir(client.httpStatic))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// the proxy
	// ignore error because of static url, which must be correct
	proxy := &httputil.ReverseProxy{Director: client.getProxyDirector()}
	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if client.session == nil {
				return nil, errors.New("no tls session available")
			}
			return client.session.Open()
		},
	}

	r.PathPrefix("/browser/click").Queries("x", "{x}", "y", "{y}").HandlerFunc(client.browserClick())
	r.Path("/ntp").HandlerFunc(client.ntp()).Methods("GET")
	// add the websocket echo client
	r.PathPrefix("/echo/").HandlerFunc(client.wsEcho())
	r.PathPrefix("/ws/{group}").HandlerFunc(client.websocketGroup())
	r.PathPrefix("/{target}/").Handler(proxy)

	r.PathPrefix("/").HandlerFunc(common.MakePreflightHandler(
		client.log,
	)).Methods("OPTIONS")

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client.log.Infof(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	client.httpServerExt = &http.Server{Addr: client.httpsAddr, Handler: r}

	client.log.Infof("launching external HTTPS on %s", client.httpsAddr)
	err = client.httpServerExt.ListenAndServeTLS(client.httpsCertFile, client.httpsKeyFile)
	if err != nil {
		client.httpServerExt = nil
		return emperror.Wrapf(err, "failed to serve")
	}
	client.httpServerExt = nil
	return nil
}

func (client *BrowserClient) screenshot(width int, height int, sigma float64) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client.log.Info("screenshot()")

		if client.browser == nil {
			client.log.Errorf("cannot create screenshot: no browser")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("cannot create screenshot: no browser")))
			return
		}
		buf, mime, err := client.browser.Screenshot(width, height, sigma)
		if err != nil {
			client.log.Errorf("cannot create screenshot: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("cannot create screenshot: %v", err)))
			return
		}
		w.Header().Add("Content-Type", mime)
		if _, err := w.Write(buf); err != nil {
			client.log.Errorf("cannot write image data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (client *BrowserClient) ServeHTTPInt(listener net.Listener) error {
	httpservmux := http.NewServeMux()
	// static files only from /static
	fs := http.FileServer(http.Dir(client.httpStatic))
	httpservmux.Handle("/static/", http.StripPrefix("/static/", fs))
	httpservmux.HandleFunc("/screenshot/full", client.screenshot(0, 0, 0))
	httpservmux.HandleFunc("/screenshot/medium", client.screenshot(800, 600, 0))
	httpservmux.HandleFunc("/screenshot/thumb", client.screenshot(240, 240, 1.5))

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

func (client *BrowserClient) ServeCmux() error {
	if err := (*client.cmuxServer).Serve(); err != nil {
		client.cmuxServer = nil
		return emperror.Wrap(err, "cmux closed")
	}
	client.cmuxServer = nil
	return nil
}

func (client *BrowserClient) _CloseServices() error {
	client.log.Infof("closing services")
	if client.grpcServer != nil {
		client.grpcServer.GracefulStop()
		client.grpcServer = nil
	}

	if client.httpServerInt != nil {
		ctx, err := context.WithTimeout(context.Background(), 2*time.Second)
		if err == nil {
			client.httpServerInt.Shutdown(ctx)
		}
		client.httpServerInt = nil
	}

	if client.httpServerExt != nil {
		ctx, err := context.WithTimeout(context.Background(), 2*time.Second)
		if err == nil {
			client.httpServerExt.Shutdown(ctx)
		}
		client.httpServerExt = nil
	}

	if client.fw != nil {
		client.fw.Shutdown()
		client.fw = nil
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
	client.log.Infof("services closed")
	return nil
}

func (client *BrowserClient) CloseInternal() error {
	client.log.Infof("closing internal services")
	if client.grpcServer != nil {
		client.grpcServer.GracefulStop()
		client.grpcServer = nil
	}

	if client.httpServerInt != nil {
		ctx, err := context.WithTimeout(context.Background(), 2*time.Second)
		if err == nil {
			client.httpServerInt.Shutdown(ctx)
		}
		client.httpServerInt = nil
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
	client.log.Infof("internal services closed")
	return nil
}

func (client *BrowserClient) Shutdown() error {
	err := client.CloseInternal()
	client.end <- true
	return err
}
