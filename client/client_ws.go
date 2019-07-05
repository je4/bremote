package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)


func (client *Client) wsEcho() func(w http.ResponseWriter, r *http.Request) {
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
