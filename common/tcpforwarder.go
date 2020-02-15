package common

import (
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"
	"io"
	"net"
	"sync"
)

type TCPForwarder struct {
	log     *logging.Logger
	quit    chan bool
	exited  chan bool
	session **yamux.Session
}

func NewTCPForwarder(session **yamux.Session, log *logging.Logger) *TCPForwarder {
	tf := &TCPForwarder{
		session: session,
		log:     log,
		quit:    make(chan bool),
		exited:  make(chan bool),
	}
	return tf
}

func (tf *TCPForwarder) Serve(listener net.Listener) error {
	var handlers sync.WaitGroup
	for {
		select {
		case <-tf.quit:
			tf.log.Info("Shutdown requested")
			listener.Close()
			handlers.Wait()
			close(tf.exited)
		default:
			conn, err := listener.Accept()
			if err != nil {
				tf.log.Errorf("failed to accept connection: %v", err)
				return emperror.Wrapf(err, "failed to accept connection")
			}
			defer conn.Close()
			handlers.Add(1)
			go func() {
				if err := tf.handleConnection(conn); err != nil {
					tf.log.Errorf("error handling connection: %v", err)
				}
				handlers.Done()
			}()
		}
	}
	return nil
}

func (tf *TCPForwarder) handleConnection(conn net.Conn) error {
	tf.log.Infof("accepted connection from %v", conn.RemoteAddr())

	if *tf.session == nil {
		return errors.New("session closed")
	}

	defer func() {
		tf.log.Infof("closing connection from %v", conn.RemoteAddr())
		conn.Close()
	}()
	target, err := (*tf.session).Open()
	if err != nil {
		return emperror.Wrapf(err, "cannot open session")
	}
	defer target.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(conn, target)
		wg.Done()
	}()
	go func() {
		io.Copy(target, conn)
		wg.Done()
	}()
	wg.Wait()
	return nil
}

func (tf *TCPForwarder) Shutdown() {
	close(tf.quit)
	<-tf.exited
	tf.log.Info("shutdown successful")
}
