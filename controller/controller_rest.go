package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strings"
)

func dummy(w http.ResponseWriter, r *http.Request) {
	return
}


func (controller *Controller) RestClientList() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestClientList()")
		clients, err := controller.GetClients()
		if err != nil {
			controller.log.Errorf("cannot get clients: %v", err)
			//http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			http.Error(w, fmt.Sprintf("cannot get clients: %v", err), http.StatusInternalServerError )
			return
		}
		json, err := json.Marshal(clients)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError )
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreList() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestKVStoreList()")
		json, err := json.Marshal(controller.kvs)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError )
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientList() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		controller.log.Infof("RestKVStoreClientList(%v)", client)

		prefix := client+"-"
		result := make(map[string]interface{})
		for name, val := range controller.kvs {
			if strings.HasPrefix(name, prefix) {
				result[strings.TrimPrefix(name, prefix)] = val
			}
		}
		json, err := json.Marshal(result)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError )
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientValue() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		key := vars["key"]
		controller.log.Infof("RestKVStoreClientValue(%v, %v)", client, key)
		k := client + "-" + key
		// ignore error. empty interface is ok
		result, _ := controller.GetVar(k)
		json, err := json.Marshal(result)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError )
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientValuePost() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		key := vars["key"]
		controller.log.Infof("RestKVStoreClientValuePost(%v, %v)", client, key)
		k := client + "-" + key

		decoder := json.NewDecoder(r.Body)
		var data interface{}
		err := decoder.Decode(&data)
		if err != nil {
			controller.log.Errorf("cannot decode data: %v", err)
			http.Error(w, fmt.Sprintf("cannot decode data: %v", err), http.StatusInternalServerError )
			return
		}
		controller.SetVar(k, data)

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"`)
	}
}

func (controller *Controller) RestKVStoreClientValueDelete() (func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		key := vars["key"]
		controller.log.Infof("RestKVStoreClientValueDelete(%v, %v)", client, key)
		k := client + "-" + key

		controller.DeleteVar(k)

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"`)
	}
}

