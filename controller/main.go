package main

import (
	"flag"
	"info_age.net/bremote/api"
	"info_age.net/bremote/common"
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
MIIEZTCCA02gAwIBAgIJAIdMY33R/ifFMA0GCSqGSIb3DQEBCwUAMIHIMQswCQYD
VQQGEwJDSDEUMBIGA1UECAwLQmFzZWwtU3RhZHQxDjAMBgNVBAcMBUJhc2VsMTQw
MgYDVQQKDCtIb2Noc2NodWxlIGbDg8K8ciBHZXN0YWx0dW5nIHVuZCBLdW5zdCBG
SE5XMSIwIAYDVQQLDBlDZW50ZXIgZm9yIERpZ2l0YWwgTWF0dGVyMRQwEgYDVQQD
DAtFeGlmc2VydmljZTEjMCEGCSqGSIb3DQEJARYUanVlcmdlbi5lbmdlQGZobncu
Y2gwHhcNMTkwNDE1MDgyNzE4WhcNMjAwNDE0MDgyNzE4WjCByDELMAkGA1UEBhMC
Q0gxFDASBgNVBAgMC0Jhc2VsLVN0YWR0MQ4wDAYDVQQHDAVCYXNlbDE0MDIGA1UE
CgwrSG9jaHNjaHVsZSBmw4PCvHIgR2VzdGFsdHVuZyB1bmQgS3Vuc3QgRkhOVzEi
MCAGA1UECwwZQ2VudGVyIGZvciBEaWdpdGFsIE1hdHRlcjEUMBIGA1UEAwwLRXhp
ZnNlcnZpY2UxIzAhBgkqhkiG9w0BCQEWFGp1ZXJnZW4uZW5nZUBmaG53LmNoMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6wa14NQoGlACqlW2qe7EnTbf
lCn4KPX5duWnR+BE0BvYaw11gTvt6dCumwgc48QuYJzlsFY/4aQEAIGTktEpioxS
p8vlzDa/wVi+tTtWrMVwPoNRbu4tKTgNEPNJ0Wx7Jhb4MqRDOC0AtSUIhIoIq7tx
Eacvmw8ZLDFYxXmG920I3w6A7BNL+TYARYlUJI6hXSppt7vggMa0ZN7umXsrrp20
nG5qnXNSReTl0NmWakQLzG/jUgGOhFY485MXCobawIIlItYqGDq0UnwD2R8kuLXg
BiWpmCqI6yVroOAllsZjzo426tzmPt920rg5lWZL49Mg8tCDxAtA14z+CTnsJQID
AQABo1AwTjAdBgNVHQ4EFgQUX96Aw24Am6a+Iob4N/wsAc447RIwHwYDVR0jBBgw
FoAUX96Aw24Am6a+Iob4N/wsAc447RIwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0B
AQsFAAOCAQEAfSwBFUAUjDsnusZQ2hfzcu57D83m5YOyb+JaaZ3uktKG2PdJMfye
dy7IvSb+hvhxsvSWCJ52YIV8ITej7NGs8R80S134yVS+cm6i0/sYjKuKXHtgbdW7
kSPttWgM79Ji4/YO8kXdzzmKUuT3XszylFfLYgNsegdlQYY5LcvzUIAJFsb1kNJX
lDe29U43+k3V/NN3u9IJvdnVP0kYCm2dtNam1LEudlnRxzmcLTOE4bQZxQSK5iU5
VdCNHmQGa0aa9FmaTXNppiqUE0IwfJF55VqFbaGmzLFJ9DOVXsKyB38Yojmv6k1u
bXNoO/zdRfQFL0B7nTB/vJn93y9on3ORjA==
-----END CERTIFICATE-----`

func main() {
	logFile := flag.String("logfile", "", "log file location")
	logLevel := flag.String("loglevel", "DEBUG", "LOGLEVEL: CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG")
	instanceName := flag.String("instance", "", "instance name")
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

	controller := NewController(*instanceName, log)

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

		time.Sleep(time.Second*2)

		pw := api.NewProxyWrapper(*instanceName, &controller.session)

		clients, err := pw.GetClients(common.SessionType_Client)
		if err != nil {
			log.Errorf("cannot get clients: %v", err)
		}
		log.Infof("Clients: %v", clients)

		cw := api.NewClientWrapper(*instanceName, &controller.session)
		for _, client := range clients {
			ret, err := cw.Ping(client)
			if err != nil {
				log.Errorf("error pinging %v: %v", client, err)
				continue
			}
			log.Infof("ping result from %v: %v", client, ret)

			opts := map[string]interface{}{"headless":false}
			if err := cw.StartBrowser(client, &opts); err != nil {
				log.Errorf("error starting client browser on %v: %v", client, err)
			}
		}
	}()

	if err := controller.Serve(); err != nil {
		log.Panicf("error serving %+v", err)
	}

}
