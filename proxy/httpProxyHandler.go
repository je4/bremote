package main

import (
	"github.com/op/go-logging"
	"io"
	"net"
	"net/http"
	"sync"
)

type httpProxyForwardHandler struct {
	log           *logging.Logger
	getConnection func() (net.Conn, error)
	counter       int64
	sync.Mutex
}

func NewHTTPProxyForwardHandler(log *logging.Logger, getConnection func() (net.Conn, error)) *httpProxyForwardHandler {
	return &httpProxyForwardHandler{
		log:           log,
		getConnection: getConnection,
		counter:       0,
	}
}

func (hpfh *httpProxyForwardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hpfh.counter++
	counter := hpfh.counter
	hpfh.log.Infof("[%v] accepting tunnel from %v: %v %v", counter, r.RemoteAddr, r.Method, r.URL)
	//w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		hpfh.log.Errorf("[%v] hijacking not supported", counter)
		return
	}
	src, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		hpfh.log.Errorf("[%v] cannot hijack: %v", counter, err)
		return
	}
	defer src.Close()

	hpfh.log.Infof("[%v] connection successfully hijacked", counter)

	dest, err := hpfh.getConnection()
	if err != nil {
		hpfh.log.Errorf("[%v] cannot get connection: %v", counter, err)
		return
	}
	defer dest.Close()

	/*
	if err := r.Write(dest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		hpfh.log.Errorf("[%v] cannot write request to dest: %v", counter, err)
		return
	}
	 */
//	var wg sync.WaitGroup
//	wg.Add(2)

	_ = func(dst io.Writer, src io.Reader) {
		cnt, _ := io.Copy(dst, src)
		hpfh.log.Infof("[%v] copied %v bytes", counter, cnt)
		if conn, ok := dst.(interface {
			CloseWrite() error
		}); ok {
			conn.CloseWrite()
		}
		//wg.Done()
	}
	copyAndClose := func(dst, src *net.TCPConn) {
		cnt, err := io.Copy(dst, src)
		if err != nil {
			hpfh.log.Warningf("[%v] Error copying to client: %s", counter, err)
		}
		hpfh.log.Warningf("[%v] copied %v byte", counter, cnt)

		dst.CloseWrite()
		src.CloseRead()
	}

	src.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	srcTCP, srcOK := src.(*net.TCPConn)
	destTCP, destOK := dest.(*net.TCPConn)

	if srcOK && destOK {
		go copyAndClose(srcTCP, destTCP)
		go copyAndClose(destTCP, srcTCP)
	}
	//	go cp(src, dest)
	//	go cp(dest, src)

//	wg.Wait()
	//	bw.Flush()
	hpfh.log.Infof("[%v] connection from %v closed", counter, r.RemoteAddr)
}
