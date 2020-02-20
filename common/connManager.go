package common

import (
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"net"
	"sync"
)

type connPair struct {
	conn1, conn2 net.Conn
}

func (cp *connPair) Close() error {
	mrb := emperror.NewMultiErrorBuilder()
	if err := cp.conn1.Close(); err != nil {
		mrb.Add(err)
	}
	if err := cp.conn2.Close(); err != nil {
		mrb.Add(err)
	}
	return mrb.ErrOrNil()
}

type connManager struct {
	sync.Mutex
	lastID int64
	conns  map[int64]connPair
	log    *logging.Logger
}

func NewConnManager(log *logging.Logger) *connManager {
	return &connManager{
		Mutex:  sync.Mutex{},
		lastID: 0,
		conns:  map[int64]connPair{},
		log:    log,
	}
}

func (cm *connManager) add(cp connPair) int64 {
	cm.Lock()
	defer cm.Unlock()

	cm.lastID++
	cm.conns[cm.lastID] = cp
	return cm.lastID
}

func (cm *connManager) closeAndRemove(id int64) error {
	cm.Lock()
	defer cm.Unlock()

	conn, ok := cm.conns[id]
	if !ok {
		return errors.New(fmt.Sprintf("connection %v not found", id))
	}
	delete(cm.conns, id)

	return conn.Close()
}

func (cm *connManager) closeAndRemoveAll() error {
	mrb := emperror.NewMultiErrorBuilder()
	for _, conn := range cm.conns {
		if err := conn.Close(); err != nil {
			mrb.Add(err)
		}
	}
	cm.conns = map[int64]connPair{}
	return mrb.ErrOrNil()
}
