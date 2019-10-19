package common

import (
	"github.com/op/go-logging"
	"net/http"
)

// hack for CORS preflight within axios clients
func MakePreflightHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Preflight: %s", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin,X-Requested-With,Content-Type,Accept,Access-Control-Request-Method,Authorization")
		w.WriteHeader(http.StatusOK)
	}
}
