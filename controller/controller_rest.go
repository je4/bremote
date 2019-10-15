package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func dummy(w http.ResponseWriter, r *http.Request) {
	return
}

func (controller *Controller) RestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			controller.log.Infof(r.RequestURI)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}

func (controller *Controller) getProxyDirector() func(req *http.Request) {
	target, _ := url.Parse("http://localhost:80/")
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		//		vars := mux.Vars(req)
		//		t := vars["target"]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = common.SingleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("X-Source-Instance", controller.GetInstance())
	}

	return director
}

func (controller *Controller) addRestRoutes(r *mux.Router) {

	// the proxy
	// ignore error because of static url, which must be correct
	proxy := &httputil.ReverseProxy{Director: controller.getProxyDirector()}
	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if controller.session == nil {
				return nil, errors.New("no tls session available")
			}
			return controller.session.Open()
		},
	}

	r.HandleFunc("/", dummy)
	r.HandleFunc("/groups", controller.RestGroupList()).Methods("GET")
	r.HandleFunc("/groups/{group}", controller.RestGroupGetMember()).Methods("GET")
	r.HandleFunc("/groups/{group}", controller.RestGroupAddInstance()).Methods("PUT")
	r.HandleFunc("/groups/{group}", controller.RestGroupDelete()).Methods("DELETE")
	r.HandleFunc("/kvstore", controller.RestKVStoreList())
	r.HandleFunc("/kvstore/{client}", controller.RestKVStoreClientList())
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValue()).Methods("GET")
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValuePut()).Methods("PUT", "POST")
	r.HandleFunc("/kvstore/{client}/{key}", controller.RestKVStoreClientValueDelete()).Methods("DELETE")
	r.HandleFunc("/client", controller.RestClientList())
	r.HandleFunc("/client/{client}/navigate", controller.RestClientNavigate()).Methods("POST")
	r.PathPrefix("/{target}/").Handler(proxy)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			controller.log.Infof(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
}

func (controller *Controller) RestGroupList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		controller.log.Info("RestGroupList()")

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		list, err := pw.GroupList(traceId)
		if err != nil {
			controller.log.Errorf("cannot get proxy group list: %v", err)
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
			controller.log.Errorf("cannot get members of group %v: %v", group, err)
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
			controller.log.Errorf("cannot add instance %v to group %v: %v", instance, group, err)
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
			controller.log.Errorf("cannot delete group %v: %v", group, err)
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

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		value, err := pw.KVStoreList(traceId)
		if err != nil {
			controller.log.Errorf("cannot get value: %v", err)
			http.Error(w, fmt.Sprintf("cannot get value: %v", err), http.StatusInternalServerError)
		}

		result := map[string]interface{}{}
		for key, val := range (*value) {
			var d interface{}
			json.Unmarshal([]byte(val), &d)
			result[key] = d
		}


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

		pw := pb.NewProxyWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		value, err := pw.KVStoreClientList(client, traceId)
		if err != nil {
			controller.log.Errorf("cannot get value: %v", err)
			http.Error(w, fmt.Sprintf("cannot get value: %v", err), http.StatusInternalServerError)
		}

		result := map[string]interface{}{}
		for key, val := range (*value) {
			var d interface{}
			json.Unmarshal([]byte(val), &d)
			result[key] = d
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

		value, err := controller.GetVar(client, key)
		if err != nil {
			controller.log.Errorf("cannot get value: %v", err)
			http.Error(w, fmt.Sprintf("cannot get value: %v", err), http.StatusInternalServerError)
		}

		json, err := json.Marshal(value)
		if err != nil {
			controller.log.Errorf("cannot marshal value: %v", err)
			http.Error(w, fmt.Sprintf("cannot marshal value: %v", err), http.StatusInternalServerError)
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

		// encode/decode to check for valid json
		decoder := json.NewDecoder(r.Body)
		var data interface{}
		err := decoder.Decode(&data)
		if err != nil {
			controller.log.Errorf("cannot decode data: %v", err)
			http.Error(w, fmt.Sprintf("cannot decode data: %v", err), http.StatusInternalServerError)
			return
		}

		err = controller.SetVar(client, key, data )
		if err != nil {
			controller.log.Errorf("cannot set value: %v", err)
			http.Error(w, fmt.Sprintf("cannot set value: %v", err), http.StatusInternalServerError)
		}

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

		err := controller.DeleteVar(client, key)
		if err != nil {
			controller.log.Errorf("cannot delete value: %v", err)
			http.Error(w, fmt.Sprintf("cannot delete value: %v", err), http.StatusInternalServerError)
		}

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
			controller.log.Errorf("cannot parse url %v: %v", params["url"].(string), err)
			http.Error(w, fmt.Sprintf("cannot parse url %v: %v", params["url"].(string), err), http.StatusInternalServerError)
		}
		nextStatus := params["nextstatus"].(string)

		controller.log.Infof("%v::RestClientNavigate(%v, %v)", client, u.String(), nextStatus)

		cw := pb.NewClientWrapper(controller.instance, controller.GetSessionPtr())
		traceId := uniqid.New(uniqid.Params{"traceid_", false})
		err = cw.Navigate(traceId, client, u, nextStatus)
		if err != nil {
			controller.log.Errorf("cannot navigate to %v: %v", u.String(), err)
			http.Error(w, fmt.Sprintf("cannot navigate to %v: %v", u.String(), err), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
	}
}
