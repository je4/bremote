package main

import (
	"fmt"
	"net/http"
)

type noProxyHandler struct {}

func (nph *noProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("no proxy call: %v %v", r.Method, r.RequestURI)))
}
