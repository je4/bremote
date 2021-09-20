package main

import (
	"flag"
	"github.com/je4/bremote/v2/common"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// static address to enable zero config distribution
//const addr = `localhost:7777`

func main() {
	configFile := flag.String("cfg", "./dataproxy.toml", "config file location")
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
	certPem := flag.String("cert", "", "tls dp certificate file in PEM format")
	keyPem := flag.String("key", "", "tls dp key file in PEM format")
	caPem := flag.String("ca", "", "tls root certificate file in PEM format")
	addr := flag.String("proxy", "localhost:7777", "proxy addr:port")
	httpStatic := flag.String("httpstatic", "", "folder with static files")
	httpTemplates := flag.String("httptemplates", "", "folder with templates")
	httpsCertPem := flag.String("httpscertpem", "", "tls dp certificate file in PEM format")
	httpsKeyPem := flag.String("httpskeypem", "", "tls dp key file in PEM format")
	httpsAddr := flag.String("httpsaddr", "", "local listen addr for https addr:port")

	flag.Parse()

	var doLocal = false
	var exPath = ""
	if !common.FileExists(*configFile) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath = filepath.Dir(ex)
		if common.FileExists(filepath.Join(exPath, *configFile)) {
			doLocal = true
			*configFile = filepath.Join(exPath, *configFile)
			*certPem = filepath.Join(exPath, *certPem)
			*keyPem = filepath.Join(exPath, *keyPem)
			*caPem = filepath.Join(exPath, *caPem)
			*httpStatic = filepath.Join(exPath, *httpStatic)
			*httpTemplates = filepath.Join(exPath, *httpTemplates)
			*httpsCertPem = filepath.Join(exPath, *httpsCertPem)
			*httpsKeyPem = filepath.Join(exPath, *httpsKeyPem)
		}
	}

	var config Config
	if *configFile != "" {
		config = LoadConfig(*configFile)
	} else {
		config = Config{
			Logfile:       *logFile,
			Loglevel:      *logLevel,
			InstanceName:  *instanceName,
			Proxy:         *addr,
			CertPEM:       *certPem,
			KeyPEM:        *keyPem,
			CaPEM:         *caPem,
			HttpStatic:    *httpStatic,
			HttpTemplates: *httpTemplates,
			HttpsCertPEM:  *httpsCertPem,
			HttpsKeyPEM:   *httpsKeyPem,
			HttpsAddr:     *httpsAddr,
		}
	}

	if config.InstanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		config.InstanceName = "dp-" + h
	}
	if doLocal {
		config.KeyPEM = filepath.Join(exPath, config.KeyPEM)
		config.CaPEM = filepath.Join(exPath, config.CaPEM)
		config.CertPEM = filepath.Join(exPath, config.CertPEM)
		config.HttpsCertPEM = filepath.Join(exPath, config.HttpsCertPEM)
		config.HttpsKeyPEM = filepath.Join(exPath, config.HttpsKeyPEM)
		config.HttpTemplates = filepath.Join(exPath, config.HttpTemplates)
		config.HttpStatic = filepath.Join(exPath, config.HttpStatic)
	}

	// create logger instance
	log, lf := common.CreateLogger(config.InstanceName, config.Logfile, config.Loglevel)
	defer lf.Close()

	rtStat := common.NewRuntimeStats(config.RuntimeInterval.Duration, log)
	if config.RuntimeInterval.Duration > 0 {
		go rtStat.Run()
	}

	client := NewDataProxy(config, log)

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
		if config.RuntimeInterval.Duration > 0 {
			rtStat.Shutdown()
		}
	}()

	if err := client.Serve(); err != nil {
		log.Panicf("error serving %+v", err)
	}

}
