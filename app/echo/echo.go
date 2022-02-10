package echo

import (
	"fmt"
	"log"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
)

// EchoApplication provides a basic "echo" service
type EchoApplication struct {
	ussdproxy.UdcpApplication

	echoBufferOffset int64
	state            ussdproxy.ApplicationState
	session          ussdproxy.Session
}

// NewEchoApplication creates a new EchoApplication
func NewEchoApplication() *EchoApplication {
	return &EchoApplication{
		state:            ussdproxy.ApplicationReady,
		echoBufferOffset: 0,
	}
}

// Name the name of the application
func (app *EchoApplication) Name() string {
	return "Echo Application"
}

// ApplicationID the unique identifier for the application
func (app *EchoApplication) ApplicationID() string {
	return "echo"
}

// Author the author of the application
func (app *EchoApplication) Author() string {
	return "NNDI"
}

// Register the EchoApplication with the server
func (app *EchoApplication) Register() {
	// noop
}

// CurrentState the EchoApplication with the server
func (app *EchoApplication) CurrentState(sessionID string) ussdproxy.ApplicationState {
	// noop
	return app.state
}

func (app *EchoApplication) GetOrCreateSession() ussdproxy.Session {
	return app.session
}

func (app *EchoApplication) UseSession(session ussdproxy.Session) {
	app.session = session
}

// OnError returns the request/response handler for the Echo Application
func (app *EchoApplication) OnError(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	fmt.Printf("Received ErrorPdu, %s", request.Data())
	return ussdproxy.NewErrorResponse(ussdproxy.ErrorCodeProtoErrorMask), nil
}

// OnData returns the request/response handler for the Echo Application
func (app *EchoApplication) OnData(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	if request.HasMoreToSend() {
		app.echoBufferOffset = 0
		fmt.Println("echo.OnData: Waiting for Client to send more data")
		return ussdproxy.NewReceiveReadyResponse(), nil
	}
	_, err := session.RecvBuffer().Read()
	if err != nil {
		return ussdproxy.NewErrorResponse(ussdproxy.ErrorCodeProtoErrorMask), nil
	}
	// Handle the decoding of the data here
	//
	// This is the point at which you may send data to an external service
	// since at this point all the data the client intended to send is complete
	//
	// Since we want to echo stuff we fill the send buffer with recv buffer contents
	if err = session.SendBuffer().FillWith(session.RecvBuffer()); err != nil {
		fmt.Println("echo.OnData(): Failed to populate SendBuffer with data from RecvBuffer")
	}
	return app.echoRecvBuffer(request, session)
}

// OnReceiveReady returns data when a Client is waiting for server data
func (app *EchoApplication) OnReceiveReady(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	if session.RecvBuffer() == nil || session.SendBuffer() == nil {
		fmt.Println("echo.OnReceiveReady(): One of the buffers was empty")
		return ussdproxy.NewReceiveReadyResponse(), nil
	}
	return app.echoRecvBuffer(request, session)
}

// OnReleaseDialogue returns the request/response handler for the Echo Application
func (app *EchoApplication) OnReleaseDialogue(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	return ussdproxy.NewReleaseDialogueResponse(ussdproxy.ReleaseCodeUserAbortMask), nil
}

func (app *EchoApplication) echoRecvBuffer(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	buf := session.SendBuffer()
	if buf.IsEmpty() {
		fmt.Println("echo.OnReceiveReady(): Send buffer was empty")
		return ussdproxy.NewReceiveReadyResponse(), nil
	}
	var seekOffset int64
	var responseData []byte
	moreToSend := false
	dataLen := int64(buf.Length())
	if dataLen <= ussdproxy.MaxDataLength {
		// small enough buffer with less than 127 bytes
		seekOffset = 0
		responseData = make([]byte, dataLen)
	} else {
		moreToSend = true
		// There's more than 127 bytes left in the buffer
		responseData = make([]byte, ussdproxy.MaxDataLength)
		seekOffset = app.echoBufferOffset
		// If we have more data in the sendBuffer than we can send
		// we have to start keeping track of offsets of our position in the send buffer
		fmt.Printf("BufferSize:%d bytes Offset:%d seekOffset:%d\n", dataLen, app.echoBufferOffset, seekOffset)
	}

	if (dataLen - seekOffset) < ussdproxy.MaxDataLength {
		moreToSend = false
		responseData = make([]byte, dataLen-seekOffset)
	}
	_, err := buf.ReadAt(responseData, seekOffset)
	if err != nil {
		log.Fatalf("Failed to read data from the SessionBuffer with (BufferSize:%d bytes Offset:%d seekOffset:%d). Got Error: %s \n", dataLen, app.echoBufferOffset, seekOffset, err)
		return ussdproxy.NewErrorResponse(ussdproxy.ErrorCodeProtoErrorMask), nil
	}
	if moreToSend && app.echoBufferOffset < dataLen {
		app.echoBufferOffset += ussdproxy.MaxDataLength
	}
	fmt.Printf("Flushing out: %s \n", string(responseData))
	return ussdproxy.NewDataResponse(request, responseData, moreToSend), nil
}
