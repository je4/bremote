package main

import (
	"encoding/json"
	"flag"
	"github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version string = "0.2"
var servername = "RemoteScreenController/" + version

// static address to enable zero config distribution
//const addr = `localhost:7777`

func main() {
	configFile := flag.String("cfg", "", "config file location")

	flag.Parse()

	config := LoadConfig(*configFile)

	if config.InstanceName == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Panic("cannot get hostname")
		}
		config.InstanceName = "controller-" + h
	}

	// create logger instance
	log, lf := common.CreateLogger(config.InstanceName, config.Logfile, config.Loglevel)
	defer lf.Close()

	controller := NewController(config, servername, log)

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
		pw := api.NewProxyWrapper(config.InstanceName, &controller.session)

		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		clients, err := pw.GetClients(traceId, common.SessionType_Client, true)
		if err != nil {
			log.Errorf("cannot get clients: %v", err)
		}
		log.Infof("[%v] Clients: %v", traceId, clients)

		cw := api.NewClientWrapper(config.InstanceName, &controller.session)
		for _, client := range clients {

			data := map[string]interface{}{}
			err := json.Unmarshal([]byte(`{"title":"Your are Client01"}`), &data)
			if err != nil {
				panic("invalid metadata")
			}
			//controller.SetVar(client.InstanceName, "wstest.html", data)

			if client.Status == common.ClientStatus_Empty {
				opts := map[string]interface{}{
					"headless":                            false,
					"start-fullscreen":                    true,
					"disable-notifications":               true,
					"disable-infobars":                    true,
					"disable-gpu":                         false,
					"allow-insecure-localhost":            true,
					"enable-immersive-fullscreen-toolbar": true,
					"views-browser-windows":               false,
					"kiosk":                               true,
					"disable-session-crashed-bubble":      true,
					"incognito":                           true,
//					"enable-features":                     "PreloadMediaEngagementData,AutoplayIgnoreWebAudio,MediaEngagementBypassAutoplayPolicies",
					"disable-features":                    "InfiniteSessionRestore,TranslateUI,PreloadMediaEngagementData,AutoplayIgnoreWebAudio,MediaEngagementBypassAutoplayPolicies",
					//"no-first-run":                        true,
					"enable-fullscreen-toolbar-reveal": false,
					"useAutomationExtension":           false,
					"enable-automation":                false,
				}

				traceId = uniqid.New(uniqid.Params{"traceid_", false})
				log.Infof("[%v] starting browser of %v", traceId, client)
				if err := cw.StartBrowser(traceId, client.InstanceName, &opts); err != nil {
					log.Errorf("[%v] error starting client browser on %v: %v", traceId, client, err)
				}
			}
		}
	}()

	if err := controller.Serve(); err != nil {
		log.Panicf("error serving %+v", err)
	}

}
