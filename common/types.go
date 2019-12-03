package common

type SessionType int

const (
	SessionType_All               SessionType = -1
	SessionType_Undefined         SessionType = 0
	SessionType_Client            SessionType = 1
	SessionType_Controller        SessionType = 2
	SessionType_Proxy             SessionType = 3
	SessionType_Actor             SessionType = 4
	SessionType_PassiveController SessionType = 5
)

var SessionTypeInt = map[string]SessionType{
	"all":               SessionType_All,
	"undefined":         SessionType_Undefined,
	"client":            SessionType_Client,
	"proxy":             SessionType_Proxy,
	"controller":        SessionType_Controller,
	"actor":             SessionType_Actor,
	"passivecontroller": SessionType_PassiveController,
}

var SessionTypeString = map[SessionType]string{
	SessionType_All:               "all",
	SessionType_Undefined:         "undefined",
	SessionType_Client:            "client",
	SessionType_Controller:        "controller",
	SessionType_Proxy:             "proxy",
	SessionType_Actor:             "actor",
	SessionType_PassiveController: "passivecontroller",
}

const (
	ClientStatus_Empty        string = ""
	ClientStatus_EmptyBrowser string = "emptyBrowser"
)

type ClientInfo struct {
	InstanceName string
	Status       string
	Type         SessionType
}
