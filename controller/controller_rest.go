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

func (controller *Controller) addRestRoutes( r *mux.Router) {
	r.HandleFunc("/", dummy)
	r.HandleFunc("/groups", controller.RestGroupList()).Methods("GET")
	r.HandleFunc("/groups/{group}", controller.RestGroupGetMember()).Methods("GET")
	r.HandleFunc("/groups/{group}", controller.RestGroupAddInstance()).Methods("PUT")
	r.HandleFunc("/groups/{group}", controller.RestGroupDelete()).Methods("DELETE")
	r.HandleFunc("/kvstore", controller.RestKVStoreList())
	r.HandleFunc("/kvstore/{client}", controller.RestKVStoreClientList())
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValue()).Methods("GET")
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValuePut()).Methods("PUT")
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValueDelete()).Methods("DELETE")
	r.HandleFunc("/client", controller.RestClientList())
	r.HandleFunc("/client/{client}/navigate", controller.RestClientNavigate()).Methods("POST")
}

func (controller *Controller) RestGroupList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestGroupList()")

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		list, err := pw.GroupList(traceId)
		if err != nil {
			controller.log.Errorf( "cannot get proxy group list: %v", err)
			http.Error(w, fmt.Sprintf("cannot get proxy group list: %v", err), http.StatusInternalServerError)
		}


		json, err := json.Marshal(list)
		if err != nil {
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestGroupGetMember() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestGroupGetMember()")
		vars := mux.Vars(r)
		group := vars["group"]

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		list, err := pw.GroupGetMembers(traceId, group)
		if err != nil {
			controller.log.Errorf( "cannot get members of group %v: %v", group, err)
			http.Error(w, fmt.Sprintf("cannot get members of group %v: %v", group, err), http.StatusInternalServerError)
		}


		json, err := json.Marshal(list)
		if err != nil {
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(json))
	}
}

func (controller *Controller) RestGroupAddInstance() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestGroupAddInstance()")
		vars := mux.Vars(r)
		group := vars["group"]

		decoder := json.NewDecoder(r.Body)
		var data interface{}
		err := decoder.Decode(&data)
		if err != nil {
			controller.log.Errorf("cannot decode data: %v", err)
			http.Error(w, fmt.Sprintf("cannot decode data: %v", err), http.StatusInternalServerError)
			return
		}
		d2, ok := data.(map[string]interface{})
		if !ok {
			controller.log.Errorf("invalid data format (not map[string]interface{})")
			http.Error(w, fmt.Sprintf("invalid data format (not map[string]interface{})"), http.StatusInternalServerError)
			return
		}
		d3, ok := d2["instance"]
		if !ok {
			controller.log.Errorf("invalid data format - no instance")
			http.Error(w, fmt.Sprintf("invalid data format - no instance"), http.StatusInternalServerError)
			return
		}
		instance, ok := d3.(string)
		if !ok {
			controller.log.Errorf("invalid data format - instance not a string")
			http.Error(w, fmt.Sprintf("invalid data format - instance not a string"), http.StatusInternalServerError)
			return
		}

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		err = pw.GroupAddInstance(traceId, group, instance)
		if err != nil {
			controller.log.Errorf( "cannot add instance %v to group %v: %v", instance, group, err)
			http.Error(w, fmt.Sprintf("cannot add instance %v to group %v: %v", instance, group, err), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
	}
}

func (controller *Controller) RestGroupDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestGroupAddInstance()")
		vars := mux.Vars(r)
		group := vars["group"]

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		err := pw.GroupDelete(traceId, group)
		if err != nil {
			controller.log.Errorf( "cannot delete group %v: %v", group, err)
			http.Error(w, fmt.Sprintf("cannot delete group %v: %v", group, err), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `"status":"ok"`)
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
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
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
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
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
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
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
			controller.log.Errorf("cannot marshal result: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal result: %v", err), http.StatusInternalServerError)
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
		controller.log.Debugf("RestKVStoreClientValuePost(%v, %v)", client, key)
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
