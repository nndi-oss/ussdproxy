package ussdproxy

import "fmt"

type ApplicationState uint8

const (
	ApplicationInitializing ApplicationState = iota
	ApplicationInitialized
	ApplicationReady
	ApplicationStopped
	ApplicationShutdown
)

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
	GetOrCreateSession() Session
	UseSession(Session)
}

// MultiplexingApplication The Core Application is an application that enables configuring the server,
// choosing applications and controlling the session. The core application
// is like a middleware that handles requests and then forwards them to the
// currently active application depending on the Client request.
type MultiplexingApplication struct {
	currentApp            UdcpApplication
	availableApplications []UdcpApplication // currently registered/active application
}

func NewMultiplexingApplication(apps ...UdcpApplication) *MultiplexingApplication {
	if len(apps) < 1 {
		panic("NewMultiplexingApplication: invalid argument provided for 'apps'")
	}

	return &MultiplexingApplication{
		availableApplications: apps,
		currentApp:            apps[0],
	}
}

func (a *MultiplexingApplication) ApplicationID() string {
	return "udcp:core"
}

func (a *MultiplexingApplication) Name() string {
	return "udcp:core"
}

func (a *MultiplexingApplication) Author() string {
	return "NNDI"
}

func (a *MultiplexingApplication) Register() {

}

func (a *MultiplexingApplication) CurrentState(string) ApplicationState {
	return ApplicationReady
}

func (a *MultiplexingApplication) OnData(request UdcpRequest, session Session) (UdcpResponse, error) {
	// TODO: select the current application
	return ProcessUdcpRequest(request, a.currentApp)
}

func (a *MultiplexingApplication) OnReceiveReady(request UdcpRequest, session Session) (UdcpResponse, error) {
	return a.currentApp.OnReceiveReady(request, session)
}

func (a *MultiplexingApplication) OnError(request UdcpRequest, session Session) (UdcpResponse, error) {
	return a.currentApp.OnError(request, session)
}

func (a *MultiplexingApplication) OnReleaseDialogue(request UdcpRequest, session Session) (UdcpResponse, error) {
	return a.currentApp.OnReleaseDialogue(request, session)
}

func (a *MultiplexingApplication) GetOrCreateSession() Session {
	return a.currentApp.GetOrCreateSession()
}

func (a *MultiplexingApplication) UseSession(session Session) {
	a.currentApp.UseSession(session)
}

func ProcessUdcpRequest(udcpReq UdcpRequest, application UdcpApplication) (UdcpResponse, error) {
	/// Inorder to process a UssdRequest we must do the following things
	/// 0. Parse the UssdRequest to a UdcpRequest
	//// a. If the UdcpRequest is a DataPdu with MTS flag set return immediately otherwise continue
	/// 1. Check the session storage if the sessionId exists
	///   a. If the UdcpRequest is DataPdu with data then create the sessionId in the store
	///   b. Store the data from the UssdRequest.Data into the session store
	/// 2. If the session exists, load the current session buffer store
	/// 3. If the UdcpRequest is a a ReceiveReady request
	/// 4. Check if

	if udcpReq == nil {
		return NewErrorResponse(ErrorCodeProtoErrorMask), ErrUnknownParse
	}
	// TODO: Get session ID from the UDCP Request
	session := application.GetOrCreateSession()
	if session == nil {
		return NewErrorResponse(ErrorCodeProtoErrorMask), fmt.Errorf("session is nil or not configured")
	}
	// The UDCP provider has sent an error frame
	if udcpReq.IsErrorPdu() {
		fmt.Println("Received ErrorPdu. Initiating ReleaseDialogue")
		application.OnError(udcpReq, session)
		return NewReleaseDialogueResponse(ReleaseCodeUserAbortMask), nil
	}
	if udcpReq.IsDataPdu() {
		// fmt.Printf("Server: Received DataPDU session=%s moreToSend=%s\n", session.SessionID(), udcpReq.HasMoreToSend())
		session.RecvBuffer().Write(udcpReq.Data())
		// The UDCP provider has more data to send
		if !udcpReq.HasMoreToSend() {
			session.Commit()
		}
		response, err := application.OnData(udcpReq, session)
		if err != nil {
			return nil, err
		}
		if !udcpReq.HasMoreToSend() && !response.HasMoreToSend() {
			session.Reset()
		}
		return response, nil
	}

	// The UDCP provider (client) is waiting to receive data from us
	if udcpReq.IsReceiveReadyPdu() {
		fmt.Printf("Server: Received ReceiveReadyPdu session=%s\n", session.SessionID())
		response, err := application.OnReceiveReady(udcpReq, session)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Server: Done executing application.OnReceiveReady session=%s\n", session.SessionID())
		if !response.HasMoreToSend() {
			session.Reset()
		}
		return response, err
	}

	if udcpReq.IsReleaseDialoguePdu() {
		response, err := application.OnReleaseDialogue(udcpReq, session)
		if err != nil {
			return nil, err
		}
		// Server must also release if client has initiated a ReleaseDialogue
		if !response.IsReleaseDialoguePdu() {
			return NewErrorResponse(ErrorCodeProtoErrorMask), nil
		}
		return response, nil
	}
	// UDCP provider didn't specify the type of payload we're dealing with
	return NewErrorResponse(ErrorCodeProtoErrorMask), nil
}
