package app

type ApplicationState uint8

// UdcpApplication is an application that can be executed by the UdcpServer
type UdcpApplication interface {
	ApplicationID() string
	Name() string
	Author() string
	Register()
	CurrentState(string) ApplicationState
	OnData(UdcpRequest, Session) (UdcpResponse, error)
	OnReceiveReady(UdcpRequest, Session) (UdcpResponse, error)
	OnError(UdcpRequest, Session) (UdcpResponse, error)
	OnReleaseDialogue(UdcpRequest, Session) (UdcpResponse, error)
}
