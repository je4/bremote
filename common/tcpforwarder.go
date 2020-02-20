package common

import (
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"io"
	"net"
	"sync"
	"time"
)

type IdleTimeoutConn struct {
	Conn net.Conn
}

func (self IdleTimeoutConn) Read(buf []byte) (int, error) {
	self.Conn.SetDeadline(time.Now().Add(5 * time.Second))
	return self.Conn.Read(buf)
}

func (self IdleTimeoutConn) Write(buf []byte) (int, error) {
	self.Conn.SetDeadline(time.Now().Add(5 * time.Second))
	return self.Conn.Write(buf)
}

type TCPForwarder struct {
	log    *logging.Logger
	quit   chan bool
	exited chan bool
	//session    **yamux.Session
	prefix        string
	prefixSkip    int64
	getTargetConn func() (net.Conn, error)
	connManager   *connManager
}

func NewTCPForwarder(prefix string, prefixSkip int64, getTargetConn func() (net.Conn, error), log *logging.Logger) *TCPForwarder {
	tf := &TCPForwarder{
		prefix:     prefix,
		prefixSkip: prefixSkip,
		//session:    session,
		getTargetConn: getTargetConn,
		log:           log,
		quit:          make(chan bool),
		connManager:   NewConnManager(log),
	}
	return tf
}

func (tf *TCPForwarder) Serve(listener net.Listener) error {
	//var handlers sync.WaitGroup
	for {
		type accepted struct {
			conn net.Conn
			err  error
		}
		c := make(chan accepted, 1)
		go func() {
			conn, err := listener.Accept()
			if err != nil {
				c <- accepted{nil, err}
				return
			}
			if tf.prefixSkip > 0 {
				buf := make([]byte, tf.prefixSkip)
				io.ReadAtLeast(conn, buf, int(tf.prefixSkip))
			}
			c <- accepted{conn, err}
		}()

		select {
		case <-tf.quit:
			tf.log.Info("Shutdown requested")
			listener.Close()
			tf.connManager.closeAndRemoveAll()
			//handlers.Wait()
			tf.quit <- true
			return nil
		case acc := <-c:
			if acc.err != nil {
				//tf.log.Errorf( "cannot accept connection: %v", acc.err)
				return emperror.Wrapf(acc.err, "cannot accept connection")
			}
			//defer acc.conn.CloseInternal()
			//handlers.Add(1)
			if err := tf.handleConnection(acc.conn); err != nil {
				tf.log.Errorf("error handling connection: %v", err)
			}
			//handlers.Done()
		}
	} // for
	return nil
}

func (tf *TCPForwarder) handleConnection(conn net.Conn) error {

	target, err := tf.getTargetConn()
	if err != nil {
		return emperror.Wrapf(err, "cannot get target session")
	}

	connID := tf.connManager.add(connPair{
		conn1: conn,
		conn2: target,
	})
	tf.log.Debugf("[%v] accepted connection from %v", connID, conn.RemoteAddr())

	/*
	tcpConn, connIsTCP := conn.(*net.TCPConn)
	if connIsTCP {
		if err := tcpConn.SetKeepAlive(false); err != nil {
			tf.log.Errorf("[%v] disabling keepalive failed: %v", connID, err)
		}
//		tcpConn.SetKeepAlivePeriod(time.Second * 3)
	}
	*/
	idleConn := IdleTimeoutConn{conn}
	idleTarget := IdleTimeoutConn{target}


	if tf.prefix != "" {
		target.Write([]byte(tf.prefix))
	}
	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			if _, err := io.Copy(idleConn, idleTarget); err != nil {
				tf.log.Debugf("[%v] error reading: %v", connID, err)
				tf.connManager.closeAndRemove(connID)
			}
			wg.Done()
		}()
		go func() {
			if _, err := io.Copy(idleTarget, idleConn); err != nil {
				tf.log.Debugf("[%v] error writing: %v", connID, err)
				tf.connManager.closeAndRemove(connID)
			}
			wg.Done()
		}()
		wg.Wait()
		tf.log.Debugf("[%v] connection done", connID)
		tf.connManager.closeAndRemove(connID)
	}()
	return nil
}

func (tf *TCPForwarder) Shutdown() {
	tf.log.Info("shutdown initiated")
	tf.quit <- true
	<-tf.quit
	tf.log.Info("shutdown successful")
}
