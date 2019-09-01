package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/je4/grpc-proxy/proxy"
	"github.com/op/go-logging"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type ProxySession struct {
	log           *logging.Logger
	instance      string
	proxy         *Proxy
	groups        []string
	sessionType   common.SessionType
	service       *ProxyServiceServer
	grpcServer    *grpc.Server
	httpServerInt *http.Server
	cmuxServer    *cmux.CMux
	session       *yamux.Session
}

func NewProxySession(instance string, session *yamux.Session, groups []string, sessionType common.SessionType, proxy *Proxy, log *logging.Logger) *ProxySession {
	ps := &ProxySession{instance: instance,
		session: session,
		proxy: proxy,
		groups: groups,
		log: log,
		sessionType: sessionType,
	}
	return ps
}

func (ps *ProxySession) Serve() error {

	// we want to create different services for HTTP and GRPC (HTTP/2)

	// create a new muxer on yamux listener
	cs := cmux.New(ps.session)
	ps.cmuxServer = &cs

	// first get http1
	httpl := (*ps.cmuxServer).Match(cmux.HTTP1Fast())
	// the rest should be grpc
	grpcl := (*ps.cmuxServer).Match(cmux.Any())

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		if err := ps.ServeGRPC(grpcl); err != nil {
			ps.log.Errorf("error serving GRPC for instance %v: %v", ps.GetInstance(), err)
		}
		wg.Done()
	}()
	go func() {
		if err := ps.ServeHTTPInt(httpl); err != nil {
			ps.log.Errorf("error serving http for instance %v: %v", ps.GetInstance(), err)
		}
		wg.Done()
	}()

	go func() {
		if err := ps.ServeCmux(); err != nil {
			ps.log.Errorf("error serving for instance %v: %v", ps.GetInstance(), err)
		}
		wg.Done()
	}()

	if err := ps.proxy.AddSession(ps, ps.GetInstance()); err != nil {
		return emperror.Wrapf(err, "cannot add session %v", ps.instance)
	}

	wg.Wait()
	return nil
}

var upgrader = websocket.Upgrader{} // use default options

func wsEcho(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (ps *ProxySession) ProxyDirector() func(req *http.Request) {
	target, _ := url.Parse("http://localhost:80/")
	targetQuery := target.RawQuery
	r := regexp.MustCompile(`^/([^/]+)/`)

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		//req.URL.Host = target.Host
		matches := r.FindStringSubmatch(req.URL.Path)
		if len(matches) >= 2 {
			req.URL.Host = matches[1] + ":80"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+matches[1])
		} else {
			req.URL.Host = target.Host
		}

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
		req.Header.Set("X-Source-Instance", ps.GetInstance())
		ps.log.Debugf("%v -> %v", ps.GetInstance(), req.URL.String())
	}
	return director
}

func (ps *ProxySession) ServeHTTPInt(listener net.Listener) error {
	httpservmux := http.NewServeMux()

	// websocket...
	httpservmux.HandleFunc("/echo/", wsEcho)

	// the proxy
	// ignore error because of static url, which must be correct
	proxy := &httputil.ReverseProxy{Director: ps.ProxyDirector()}

	r2 := regexp.MustCompile(`^([^:]+):`)
	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			matches := r2.FindStringSubmatch(addr)
			if len(matches) < 2 {
				return nil, errors.New(fmt.Sprintf("invalid address %s", addr))
			}
			target := matches[1]
			sess, err := ps.proxy.GetSession(target)
			if err != nil {
				ps.log.Errorf("invalid target %s from address %s: %v", target, addr, err)
				return nil, emperror.Wrapf(err, "invalid target %s from address %s", target, addr)
			}
			return sess.session.Open()
		},
	}
	httpservmux.Handle("/", proxy)

	ps.httpServerInt = &http.Server{Addr: ":80", Handler: httpservmux}

	ps.log.Info("launching HTTP server over TLS connection...")
	// starting http server
	if err := ps.httpServerInt.Serve(listener); err != nil {
		ps.httpServerInt = nil
		return emperror.Wrapf(err, "failed to serve")
	}

	ps.httpServerInt = nil
	return nil
}

func (ps *ProxySession) ServeCmux() error {
	if err := (*ps.cmuxServer).Serve(); err != nil {
		ps.cmuxServer = nil
		return emperror.Wrap(err, "cmux closed")
	}
	ps.cmuxServer = nil
	return nil
}

func (ps *ProxySession) ServeGRPC(listener net.Listener) error {

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// Make sure we never forward internal services.
		//if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
		//	return nil, status.Errorf(codes.Unimplemented, "Unknown method")
		//}
		traceId, sourceInstance, targetInstance, err := common.RpcContextMetadata(ctx)
		if err != nil {
			ps.log.Errorf("invalid metadata in call to %v: %v", "Ping()", err)
			return nil, nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
		}

		// check for session
		sess, err := ps.proxy.GetSession(targetInstance)
		if err != nil {
			ps.log.Errorf("[%v] instance not found in call to %v::%v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)
			return nil, nil, status.Errorf(codes.Unavailable, "[%v] instance not found in call to %v::%v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)
		}

		// make sure, that we transfer the metadata to the target client
		ctx = metadata.AppendToOutgoingContext(ctx, "sourceInstance", sourceInstance, "targetInstance", targetInstance, "traceId", traceId)

		conn, err := grpc.DialContext(ctx, ":7777",
			grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				if sess.session == nil {
					return nil, errors.New(fmt.Sprintf("[%v] session %s closed", traceId, s))
				}
				return sess.session.Open()
			}),
			grpc.WithCodec(proxy.Codec()),
			//			grpc.WithDefaultCallOptions(grpc.ForceCodec(proxy.Codec())),
		)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "[%v] error dialing %v on session %v for %v", traceId, ":7777", sess.GetInstance(), fullMethodName)
		}

		ps.log.Debugf("[%v] directing %v -> %v%v", traceId, sourceInstance, targetInstance, fullMethodName)

		return ctx, conn, nil
	}

	// create a gRPC server object
	ps.grpcServer = grpc.NewServer(grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	ps.service = NewProxyServiceServer(ps, ps.log)

	// attach the Ping service to the server
	pb.RegisterProxyServiceServer(ps.grpcServer, ps.service)

	// start the gRPC erver
	ps.proxy.log.Info("launching gRPC server over TLS connection...")
	if err := ps.grpcServer.Serve(listener); err != nil {
		ps.proxy.log.Errorf("failed to serve: %v", err)
		return emperror.Wrapf(err, "failed to serve")
	}

	return nil
}

func (ps *ProxySession) GetService() *ProxyServiceServer {
	return ps.service
}

func (ps *ProxySession) GetInstance() string {
	return ps.instance
}

func (ps *ProxySession) GetSessionPtr() **yamux.Session {
	return &ps.session
}

func (ps *ProxySession) SetInstance(newinstance string) error {
	if err := ps.proxy.RenameSession(ps.instance, newinstance); err != nil {
		return emperror.Wrapf(err, "error renaming %v -> %v", ps.instance, newinstance)
	}
	ps.instance = newinstance
	return nil
}

func (ps *ProxySession) GetSessionType() common.SessionType {
	return ps.sessionType
}

func (ps *ProxySession) SetSessionType(sessionType common.SessionType) {
	ps.sessionType = sessionType
	ps.log.Debugf("set session type of %v to %v", ps.instance, sessionType)
}

func (ps *ProxySession) GetProxy() *Proxy {
	return ps.proxy
}

func (ps *ProxySession) Close() error {
	_, err := ps.proxy.RemoveSession(ps.instance)
	if err != nil {
		return emperror.Wrapf(err, "cannot remove session %v", ps.instance)
	}
	ps.grpcServer.GracefulStop()
	return ps.session.Close()
}
