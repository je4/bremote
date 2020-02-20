package common

import (
	"github.com/op/go-logging"
	"net/http"
)

type HTTPProxyForwarder struct {
	log    *logging.Logger
	quit   chan bool
	exited chan bool
	//getConnection   func() (net.Conn, error)
	proxyHandler    http.Handler
	nonProxyHandler http.Handler
}

func NewHttpProxyForwarder(proxyHandler http.Handler, nonProxyHandler http.Handler, log *logging.Logger) *HTTPProxyForwarder {
	return &HTTPProxyForwarder{
		log:             log,
		quit:            make(chan bool),
		exited:          make(chan bool),
		proxyHandler:    proxyHandler,
		nonProxyHandler: nonProxyHandler,
	}
}


func (hpf *HTTPProxyForwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" || r.URL.IsAbs() {
		hpf.proxyHandler.ServeHTTP(w, r)
		/*
		if err := hpf.forwardTunnel(w, r); err != nil {
			hpf.log.Errorf("error forwarding to target node: %v", err)
		}
		*/
	} else {
		hpf.nonProxyHandler.ServeHTTP(w, r)
	}
}
