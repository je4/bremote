package main

import (
	"flag"
	"github.com/mintance/go-uniqid"
	"github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// static address to enable zero config distribution
const addr = `localhost:7777`

// static server certificate to enable zero config distribution
const rootPEM = `
-----BEGIN CERTIFICATE-----
MIID9jCCAt6gAwIBAgIJAOi0aUVvT/JiMA0GCSqGSIb3DQEBCwUAMIGPMQswCQYD
VQQGEwJERTEMMAoGA1UECAwDTlJXMQ4wDAYDVQQHDAVFYXJ0aDEXMBUGA1UECgwO
UmFuZG9tIENvbXBhbnkxCzAJBgNVBAsMAklUMRcwFQYDVQQDDA53d3cucmFuZG9t
LmNvbTEjMCEGCSqGSIb3DQEJARYUanVlcmdlbkBpbmZvLWFnZS5uZXQwHhcNMTkw
NjI2MTMxNDI5WhcNMjkwNjIzMTMxNDI5WjCBjzELMAkGA1UEBhMCREUxDDAKBgNV
BAgMA05SVzEOMAwGA1UEBwwFRWFydGgxFzAVBgNVBAoMDlJhbmRvbSBDb21wYW55
MQswCQYDVQQLDAJJVDEXMBUGA1UEAwwOd3d3LnJhbmRvbS5jb20xIzAhBgkqhkiG
9w0BCQEWFGp1ZXJnZW5AaW5mby1hZ2UubmV0MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAxzI1anTX5H082jcBziT65bBvv37WbbEPvtY+CfBMp1Wyy59T
sH8aIHryiv6zXhKP3Xw7H6ybFnS+ZMyMp66QfzoRkgRGK85Pz71iHSOagqIDRo0i
rM51QSUwc/0b/NiLTi+Lo/pf+BUmQyOnCum71Mw1JvI8qUVAQssu4tcK6wG8x+Ag
2Sxf38E2hqTzt599hvmrVcmRCf/yOW35Igjox/m+Fzq99BeCzRva6Qrl7aN2RaI+
1Pq0uvchPNSGQUgMMC2v5ZbVc9ruduxMav7jj87EPdMHzuB00RZLImHj1qEaSvK6
0zjkRU5JztvVefxGOlLWszlQc597/OVXm152RwIDAQABo1MwUTAdBgNVHQ4EFgQU
yj4IGbn7Iz2wGuITyJJjBxYy/pAwHwYDVR0jBBgwFoAUyj4IGbn7Iz2wGuITyJJj
BxYy/pAwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAfykYtqq+
XEiu5H9fyGfoczbuGDbtHkPU2/IYuaaJuHhT+S++tOlEqZEVgfRMibHdYryFZtx1
yPbWAspdLgTcooEHZEfEdP1YcbKXtpBn+fmxaWqN78ETl1j2E+e49Ykl1ztmSE+5
CNRmynE8E3RxCtUK+O2+gaChxZn4A/epnlO4JaMDPep6H+Ba/pcPKyIgesqZPv5S
d/uNLKFokMQMqVyV8hSWwE/D78oED5f/eoJ7UAEDh2jhLtZaodFN7nYnj9MOCqmE
48Se8WlIO64SzhXIcmhQixvFswxJm1Fru1JIM8rfDraSyPpKzwJdr2g80Il53pGt
/v72XYHCtlzjSQ==
-----END CERTIFICATE-----`

func main() {
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
	certPem := flag.String("cert", "", "tls client certificate file in PEM format")
	keyPem := flag.String("key", "", "tls client key file in PEM format")
	caPem := flag.String("ca", "", "tls root certificate file in PEM format")

	flag.Parse()
	if *instanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		instanceName = &h
	}

	// create logger instance
	log, lf := common.CreateLogger("controller-"+*instanceName, *logFile, *logLevel)
	defer lf.Close()

	controller := NewController(*instanceName, *caPem, *certPem, *keyPem, log)

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		log.Infof("shutdown requested")
		controller.Shutdown()
	}()

	go func() {

		time.Sleep(time.Second * 2)

		if controller.session == nil {
			log.Error("session connection not available")
			return
		}
		pw := api.NewProxyWrapper(*instanceName, &controller.session)

		traceId :=uniqid.New(uniqid.Params{"traceid_", false})
		clients, err := pw.GetClients(traceId, common.SessionType_Client)
		if err != nil {
			log.Errorf("cannot get clients: %v", err)
		}
		log.Infof("[%v] Clients: %v", traceId, clients)

		cw := api.NewClientWrapper(*instanceName, &controller.session)
		for _, client := range clients {
			traceId :=uniqid.New(uniqid.Params{"traceid_", false})
			ret, err := cw.Ping(traceId, client)
			if err != nil {
				log.Errorf("[%v] error pinging %v: %v", traceId, client, err)
				continue
			}
			log.Infof("[%v] ping result from %v: %v", traceId, client, ret)

			opts := map[string]interface{}{
				"headless":              false,
				"start-fullscreen":      true,
				"disable-notifications": true,
				"disable-infobars":      true,
				"disable-gpu":           false,
			}

			traceId =uniqid.New(uniqid.Params{"traceid_", false})
			log.Infof("[%v] starting down browser of %v", traceId, client)
			if err := cw.StartBrowser(traceId, client, &opts); err != nil {
				log.Errorf("[%v] error starting client browser on %v: %v", traceId, client, err)
			}

		}

		time.Sleep(time.Second * 15)
		if controller.session == nil {
			log.Error("session connection not available")
			return
		}
		traceId =uniqid.New(uniqid.Params{"traceid_", false})
		clients, err = pw.GetClients(traceId, common.SessionType_Client)
		if err != nil {
			log.Errorf("[%v] cannot get clients: %v", traceId, err)
		}
		log.Infof("Clients: %v", clients)
		for _, client := range clients {
			traceId =uniqid.New(uniqid.Params{"traceid_", false})
			log.Infof("[%v] shutting down browser of %v", traceId, client)
			if err := cw.ShutdownBrowser(traceId, client); err != nil {
				log.Errorf("error shutting down browser of %v: %v", client, err)
				continue
			}
		}
	}()

	if err := controller.Serve(); err != nil {
		log.Panicf("error serving %+v", err)
	}

}
