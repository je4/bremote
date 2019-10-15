package main

import (
	"flag"
	"github.com/je4/bremote/common"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// static address to enable zero config distribution
//const addr = `localhost:7777`

func main() {
	configFile := flag.String("cfg", "", "config file location")
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
	certPem := flag.String("cert", "", "tls client certificate file in PEM format")
	keyPem := flag.String("key", "", "tls client key file in PEM format")
	caPem := flag.String("ca", "", "tls root certificate file in PEM format")
	addr := flag.String("proxy", "localhost:7777", "proxy addr:port")
	httpStatic := flag.String("httpstatic", "", "folder with static files")
	httpTemplates := flag.String("httptemplates", "", "folder with templates")
	httpsCertPem := flag.String("httpscertpem", "", "tls client certificate file in PEM format" )
	httpsKeyPem := flag.String("httpskeypem", "", "tls client key file in PEM format" )
	httpsAddr := flag.String("httpsaddr", "", "local listen addr for https addr:port" )

	flag.Parse()

	var config Config
	if *configFile != "" {
		config = LoadConfig(*configFile)
	} else {
		config = Config{
			Logfile: *logFile,
			Loglevel: *logLevel,
			InstanceName: *instanceName,
			Proxy: *addr,
			CertPEM: *certPem,
			KeyPEM: *keyPem,
			CaPEM: *caPem,
			HttpStatic: *httpStatic,
			HttpTemplates: *httpTemplates,
			HttpsCertPEM:*httpsCertPem,
			HttpsKeyPEM:*httpsKeyPem,
			HttpsAddr:*httpsAddr,
		}
	}

	if config.InstanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		config.InstanceName = "client-"+h
	}

	// create logger instance
	log, lf := common.CreateLogger(config.InstanceName, config.Logfile, config.Loglevel)
	defer lf.Close()

	client := NewClient(config, log)

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		log.Infof("shutdown requested")
		client.Shutdown()
	}()

	if err := client.Serve(); err != nil {
		log.Panicf("error serving %+v", err)
	}

}
