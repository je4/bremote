package common

import (
	"context"
	"errors"
	"strings"

	//	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"google.golang.org/grpc/metadata"
	"os"
)

var _logformat = logging.MustStringFormatter(
	`%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
)

func SingleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func CreateLogger(module string, logfile string, loglevel string) (log *logging.Logger, lf *os.File) {
	log = logging.MustGetLogger(module)
	var err error
	if logfile != "" {
		lf, err = os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Cannot open logfile %v: %v", logfile, err)
		}
		//defer lf.Close()

	} else {
		lf = os.Stderr
	}
	backend := logging.NewLogBackend(lf, "", 0)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.GetLevel(loglevel), "")

	logging.SetFormatter(_logformat)
	logging.SetBackend(backendLeveled)

	return
}

func RpcContextMetadata2(ctx context.Context) (traceId string, sourceInstance string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	//md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", "", errors.New("no metadata in context")
	}

	// check for sourceInstance Metadata
	si, exists := md["sourceinstance"]
	if !exists {
		return "", "", errors.New("no sourceinstance in context")
	}
	sourceInstance = si[0]

	// check for targetInstance Metadata
	tr, exists := md["traceid"]
	if !exists {
		return "", "", errors.New("no traceid in context")
	}
	traceId = tr[0]

	return
}
func RpcContextMetadata(ctx context.Context) (traceId string, sourceInstance string, targetInstance string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	//md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", "", "", errors.New("no metadata in context")
	}

	// check for targetInstance Metadata
	ti, exists := md["targetinstance"]
	if !exists {
		return "", "", "", errors.New("no targetinstance in context")
	}
	targetInstance = ti[0]

	// check for sourceInstance Metadata
	si, exists := md["sourceinstance"]
	if !exists {
		return "", "", "", errors.New("no sourceinstance in context")
	}
	sourceInstance = si[0]

	// check for targetInstance Metadata
	tr, exists := md["traceid"]
	if !exists {
		return "", "", "", errors.New("no traceid in context")
	}
	traceId = tr[0]

	return
}