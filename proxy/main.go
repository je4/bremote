package main

import (
	"flag"
	"info_age.net/bremote/common"
	"log"
	"os"
)

// static address to enable zero config distribution
const addr = `localhost:7777`

func main() {
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
	certPem := flag.String("cert", "", "tls certificate file in PEM format")
	keyPem := flag.String("key", "", "tls key file in PEM format")
	caPem := flag.String("ca", "", "tls ca file in PEM format")

	flag.Parse()
	if *instanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		instanceName = &h
	}

	// create logger instance
	log, lf := common.CreateLogger("proxy-"+*instanceName, *logFile, *logLevel)
	defer lf.Close()

	proxy, err := NewProxy(*instanceName, *caPem, *certPem, *keyPem, log)
	if err != nil {
		log.Fatalf("error creating proxy instance: %+v", err)
	}

	err = proxy.ListenServe()
	if err != nil {
		log.Fatalf("error listening: %+v", err)
	}

}