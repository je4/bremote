package main

import (
	"flag"
	"github.com/je4/bremote/common"
	"github.com/prologic/bitcask"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("cfg", "", "config file location")
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
	certPem := flag.String("cert", "", "tls certificate file in PEM format")
	keyPem := flag.String("key", "", "tls key file in PEM format")
	caPem := flag.String("ca", "", "tls ca file in PEM format")
	addr := flag.String("listen", "localhost:7777", "interface:port to listen")

	flag.Parse()
	var config Config
	if *configFile != "" {
		config = LoadConfig(*configFile)
	} else {
		config = Config{
			Logfile:      *logFile,
			Loglevel:     *logLevel,
			InstanceName: *instanceName,
			CertPEM:      *certPem,
			KeyPEM:       *keyPem,
			CaPEM:        *caPem,
			TLSAddr:      *addr,
		}
	}

	if config.InstanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		config.InstanceName = "proxy-" + h
	}

	// create logger instance
	log, lf := common.CreateLogger(config.InstanceName, config.Logfile, config.Loglevel)
	defer lf.Close()

	rtStat := common.NewRuntimeStats(config.RuntimeInterval.Duration, log )
	if config.RuntimeInterval.Duration > 0 {
		go rtStat.Run()
	}

	db, err := bitcask.Open(config.KVDBFile, bitcask.WithSync(true), bitcask.WithMaxValueSize(int(1 << 31)))
	if err != nil {
		log.Panicf("Error opening key value store \"%s\": %v", config.KVDBFile, err)
	}
	defer func() {
		log.Info("closing key value store")
		db.Close()
	}()

	proxy, err := NewProxy(config, db, log)
	if err != nil {
		log.Fatalf("error creating proxy instance: %+v", err)
	}

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		log.Infof("shutdown requested")
		proxy.Close()
		if config.RuntimeInterval.Duration > 0 {
			rtStat.Shutdown()
		}
	}()

	err = proxy.ListenServe()
	if err != nil {
		log.Fatalf("error listening: %+v", err)
	}
	log.Info("proxy ended")

}
