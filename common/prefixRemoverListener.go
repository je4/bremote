package common

import (
	"io"
	"net"
)

type prefixRemoveListener struct {
	net.Listener
	prefixLength int64
}

func NewPrefixRemoverListener( prefixLength int64, listener net.Listener) net.Listener {
	return &prefixRemoveListener{
		Listener:     listener,
		prefixLength: prefixLength,
	}
}

func (prl *prefixRemoveListener) Accept() (net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)
	for attempt := 0; ; attempt++ {
		conn, err = prl.Listener.Accept()
		if err == nil  {
			break
		}
		if t, ok := err.(interface{ Temporary() bool }); !ok || !t.Temporary() {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	buf := make([]byte, prl.prefixLength)
	io.ReadAtLeast(conn, buf, int(prl.prefixLength))
	return conn, nil
}