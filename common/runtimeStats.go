package common

import (
	"github.com/op/go-logging"
	"runtime"
	"github.com/shirou/gopsutil/mem"
	"time"
)

type runtimeStats struct {
	interval time.Duration
	log *logging.Logger
	quit     chan bool
}

func NewRuntimeStats(interval time.Duration, log *logging.Logger) *runtimeStats {
	return &runtimeStats{
		interval: interval,
		log:      log,
		quit:     make(chan bool),
	}
}

func (rs *runtimeStats) Shutdown() {
	rs.quit <- true
}

func (rs *runtimeStats) Run() {
	for {
		select {
		case <-rs.quit:
			return
		case <-time.After(rs.interval):
			v, err := mem.VirtualMemory()
			if err != nil {
				rs.log.Errorf("cannot get vmem: %v", err)
				continue
			}
			routines := runtime.NumGoroutine()
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			rs.log.Debugf("Go Routines: %v, Memory Total: %v, Used:%v, UsedPercent:%f%%\n", routines, v.Total, v.Used, v.UsedPercent)
		}
	}
}
