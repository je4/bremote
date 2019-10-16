package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

/**
Websocket Echo Connection
/echo/
*/
func (client *BrowserClient) wsEcho() func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{} // use default options

	wsecho := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
	return wsecho
}

/**
Websocket Group Connection
/ws/{group}/
*/
func (client *BrowserClient) websocketGroup() func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // upgrade all...
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			client.log.Errorf("cannot upgrade to websockets: %v", err)
			return
		}

		vars := mux.Vars(r)
		group := vars["group"]
		client.log.Debugf("websocketGroup(%v)", group)

		cws := &ClientWebsocket{
			client: client,
			conn:   conn,
			group:  group,
			send:   make(chan []byte),
		}
		client.SetGroupWebsocket(group, cws)
		go cws.writePump()
		go cws.readPump()
	}
}
