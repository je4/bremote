package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	pb "github.com/je4/bremote/api"
	"github.com/mintance/go-uniqid"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func dummy(w http.ResponseWriter, r *http.Request) {
	return
}

func (controller *Controller) RestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Do stuff here
			controller.log.Infof(r.RequestURI)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}

func (controller *Controller) RestClientList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestClientList()")
		clients, err := controller.GetClients()
		if err != nil {
			controller.log.Errorf("cannot get clients: %v", err)
			//http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			http.Error(w, fmt.Sprintf("cannot get clients: %v", err), http.StatusInternalServerError)
			return
		}
		json, err := json.Marshal(clients)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestKVStoreList()")
		json, err := json.Marshal(controller.kvs)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		controller.log.Infof("RestKVStoreClientList(%v)", client)

		prefix := client + "-"
		result := make(map[string]interface{})
		for name, val := range controller.kvs {
			if strings.HasPrefix(name, prefix) {
				result[strings.TrimPrefix(name, prefix)] = val
			}
		}
		json, err := json.Marshal(result)
		if err != nil {
			controller.log.Errorf("cannot marshal reult: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientValue() func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, fmt.Sprintf("cannot marshal reult: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestKVStoreClientValuePut() func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, fmt.Sprintf("cannot decode data: %v", err), http.StatusInternalServerError)
			return
		}
		controller.SetVar(k, data)

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
	}
}

func (controller *Controller) RestKVStoreClientValueDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]
		key := vars["key"]
		controller.log.Infof("RestKVStoreClientValueDelete(%v, %v)", client, key)
		k := client + "-" + key

		controller.DeleteVar(k)

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
	}
}

func (controller *Controller) RestClientNavigate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		client := vars["client"]

		decoder := json.NewDecoder(r.Body)
		var data interface{}
		err := decoder.Decode(&data)
		if err != nil {
			controller.log.Errorf("cannot decode data: %v", err)
			http.Error(w, fmt.Sprintf("cannot decode data: %v", err), http.StatusInternalServerError)
			return
		}
		params := data.(map[string]interface{})
		u, err := url.Parse(params["url"].(string))
		if err != nil {
			controller.log.Errorf( "cannot parse url %v: %v", params["url"].(string), err)
			http.Error(w, fmt.Sprintf("cannot parse url %v: %v", params["url"].(string), err), http.StatusInternalServerError)
		}
		nextStatus := params["nextstatus"].(string)

		controller.log.Infof("%v::RestClientNavigate(%v, %v)", client, u.String(), nextStatus)

		cw := pb.NewClientWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		err = cw.Navigate(traceId, client, u, nextStatus)
		if err != nil {
			controller.log.Errorf( "cannot navigate to %v: %v", u.String(), err)
			http.Error(w, fmt.Sprintf("cannot navigate to %v: %v", u.String(), err), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
	}
}
