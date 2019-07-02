package main

import (
	"flag"
	"github.com/je4/bremote/common"
	"log"
	"os"
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
			Logfile: *logFile,
			Loglevel: *logLevel,
			InstanceName: *instanceName,
			CertPEM: *certPem,
			KeyPEM: *keyPem,
			CaPEM: *caPem,
			TLSAddr:*addr,
		}
	}

	if config.InstanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		config.InstanceName = "proxy-"+h
	}

	// create logger instance
	log, lf := common.CreateLogger(config.InstanceName, config.Logfile, config.Loglevel)
	defer lf.Close()

	proxy, err := NewProxy(config, log)
	if err != nil {
		log.Fatalf("error creating proxy instance: %+v", err)
	}

	err = proxy.ListenServe()
	if err != nil {
		log.Fatalf("error listening: %+v", err)
	}

}