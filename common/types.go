package common

type SessionType int

const (
	SessionType_All        SessionType = -1
	SessionType_Undefined  SessionType = 0
	SessionType_Client     SessionType = 1
	SessionType_Controller SessionType = 2
	SessionType_Proxy      SessionType = 3
	SessionType_Actor      SessionType = 4
)

const (
	ClientStatus_Empty        string = ""
	ClientStatus_EmptyBrowser string = "emptyBrowser"
)

type ClientInfo struct {
	InstanceName string
	Status       string
	Type         SessionType
}
