package common

type SessionType int

const (
	SessionType_All        SessionType = -1
	SessionType_Undefined  SessionType = 0
	SessionType_Client     SessionType = 1
	SessionType_Controller SessionType = 2
)
