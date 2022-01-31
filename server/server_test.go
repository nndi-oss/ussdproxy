package server_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func removeBufferCacheFile(t *testing.T) {

	if err := os.Remove("./udcp.sessions"); err != nil {
		t.Error("Failed to remove buffer file")
	}
}

func makeUssdFormRequest(req UdcpRequest) url.Values {
	return map[string][]string{
		"sessionId":   {"testSession"},
		"phoneNumber": {"+265888123456"},
		"text":        {req.ToString()},
		"channel":     {"testChannel"},
	}
}

func decodePDU(responseData []byte) (UdcpResponse, error) {
	typ := responseData[0]
	ln := int(responseData[1])
	moreToSend := false
	if typ == DataPduWithMtsType {
		moreToSend = true
	}
	// skip the third byte since it's a separator
	data := make([]byte, ln)
	r := bytes.NewReader(responseData)
	r.Seek(3, 0)
	// Parse the header and data here
	if _, err := io.ReadAtLeast(r, data, ln); err != nil {
		return nil, InvalidHeaderError
	}
	return &udcpResponse{
		request: nil,
		header: &UdcpHeader{
			Type:       typ,
			Version:    UdcpVersion,
			MoreToSend: moreToSend,
		},
		data: data,
		len:  ln,
	}, nil
}

func makeReceiveReadyPdu() UdcpRequest {
	return &udcpRequest{
		header: &UdcpHeader{
			MoreToSend: false,
			Type:       ReceiveReadyPduType,
			Version:    UdcpVersion,
		},
		data: []byte("__NODATA__"),
		len:  len("__NODATA__"),
	}
}

func makeDataPdu(data []byte, moreToSend bool) UdcpRequest {
	typ := DataLongPduType
	if moreToSend {
		typ = DataPduWithMtsType
	}
	return &udcpRequest{
		header: &UdcpHeader{
			MoreToSend: moreToSend,
			Type:       uint8(typ),
			Version:    UdcpVersion,
		},
		data: data,
		len:  len(data),
	}
}

func TestReceiveReadies(t *testing.T) {
	removeBufferCacheFile(t)
	s := &fasthttp.Server{
		Handler: createHandler(NewTestEchoApplication()),
		Name:    "TestReceiveReadies",
	}

	serverCh := make(chan struct{})
	go func() {
		if err := s.ListenAndServe("localhost:8327"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		close(serverCh)
	}()

	clientCh := make(chan struct{})
	go func() {
		client := &http.Client{}
		ussdFormValues := makeUssdFormRequest(makeReceiveReadyPdu())
		httpRes, err := client.PostForm("http://localhost:8327", ussdFormValues)
		if err != nil {
			t.Errorf("Failed to connect to  %s", err)
			return
		}
		data := make([]byte, 140)
		httpRes.Body.Read(data)
		res, err := decodePDU(data)
		if !res.IsReceiveReadyPdu() {
			t.Fail()
		}

		close(clientCh)
	}()

	select {
	case <-clientCh:
	case <-time.After(time.Second):
		t.Fatalf("Client timeout")
	}
	select {
	case <-serverCh:
	case <-time.After(time.Second):
		close(serverCh)
		t.Log("Server timeout")
	}
}

func TestCanReceiveDataPdu(t *testing.T) {
	removeBufferCacheFile(t)
	s := &fasthttp.Server{
		Handler: createHandler(NewTestEchoApplication()),
		Name:    "TestReceiveReadies",
	}

	serverCh := make(chan struct{})
	go func() {
		if err := s.ListenAndServe("localhost:8327"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		close(serverCh)
	}()
	client := &http.Client{}
	ussdFormValues := makeUssdFormRequest(makeDataPdu([]byte("Hello UDCP World"), false))
	httpRes, err := client.PostForm("http://localhost:8327", ussdFormValues)
	if err != nil {
		t.Error("Failed to connect to  {}", err)
	}
	data := make([]byte, 127)
	httpRes.Body.Read(data)
	res, err := decodePDU(data)
	if !res.IsDataPdu() {
		t.Fail()
	}
	if string(res.Data()) != "Hello UDCP World" {
		t.Fatalf("Unexpected data. Expected: Hello UDCP World, Got: %s", string(res.Data()))
	}
	select {
	case <-serverCh:
	case <-time.After(time.Second):
		t.Log("Server timeout.")
	}
}

// TestEchoApplication provides a basic "echo" service
type TestEchoApplication struct {
	UdcpApplication

	echoBufferOffset int64
	state            ApplicationState
}

// NewTestEchoApplication creates a new TestEchoApplication
func NewTestEchoApplication() *TestEchoApplication {
	return &TestEchoApplication{
		state:            ReceiveReadyPduType,
		echoBufferOffset: 0,
	}
}

// Name the name of the application
func (app *TestEchoApplication) Name() string {
	return "Echo Application"
}

// ApplicationID the unique identifier for the application
func (app *TestEchoApplication) ApplicationID() string {
	return "echo"
}

// Author the author of the application
func (app *TestEchoApplication) Author() string {
	return "NNDI"
}

// Register the TestEchoApplication with the server
func (app *TestEchoApplication) Register() {
	// noop
}

// CurrentState the TestEchoApplication with the server
func (app *TestEchoApplication) CurrentState(sessionID string) ApplicationState {
	// noop
	return app.state
}

// OnError returns the request/response handler for the Echo Application
func (app *TestEchoApplication) OnError(request UdcpRequest, session Session) (UdcpResponse, error) {
	fmt.Printf("Received ErrorPdu, %s", request.Data())
	return NewErrorResponse(ErrorCodeProtoErrorMask), nil
}

// OnData returns the request/response handler for the Echo Application
func (app *TestEchoApplication) OnData(request UdcpRequest, session Session) (UdcpResponse, error) {
	if request.HasMoreToSend() {
		app.echoBufferOffset = 0
		fmt.Println("echo.OnData: Waiting for Client to send more data")
		return NewReceiveReadyResponse(), nil
	}
	_, err := session.RecvBuffer().Read()
	if err != nil {
		return NewErrorResponse(ErrorCodeProtoErrorMask), nil
	}
	// Since we want to echo stuff we fill the send buffer with recv buffer contents
	if err = session.SendBuffer().FillWith(session.RecvBuffer()); err != nil {
		fmt.Println("echo.OnData(): Failed to populate SendBuffer with data from RecvBuffer")
	}
	return app.echoRecvBuffer(request, session)
}

// OnReceiveReady returns data when a Client is waiting for server data
func (app *TestEchoApplication) OnReceiveReady(request UdcpRequest, session Session) (UdcpResponse, error) {
	if session.RecvBuffer() == nil || session.SendBuffer() == nil {
		return NewReceiveReadyResponse(), nil
	}
	return app.echoRecvBuffer(request, session)
}

// OnReleaseDialogue returns the request/response handler for the Echo Application
func (app *TestEchoApplication) OnReleaseDialogue(request UdcpRequest, session Session) (UdcpResponse, error) {
	return NewReleaseDialogueResponse(ReleaseCodeUserAbortMask), nil
}

func (app *TestEchoApplication) echoRecvBuffer(request UdcpRequest, session Session) (UdcpResponse, error) {
	buf := session.SendBuffer()
	if buf.IsEmpty() {
		return NewReceiveReadyResponse(), nil
	}
	var seekOffset int64
	var responseData []byte
	moreToSend := false
	dataLen := int64(buf.Length())
	if dataLen <= MaxDataLength {
		// small enough buffer with less than 127 bytes
		seekOffset = 0
		responseData = make([]byte, dataLen)
	} else {
		moreToSend = true
		// There's more than 127 bytes left in the buffer
		responseData = make([]byte, MaxDataLength)
		seekOffset = app.echoBufferOffset
	}

	if (dataLen - seekOffset) < MaxDataLength {
		moreToSend = false
		responseData = make([]byte, dataLen-seekOffset)
	}
	_, err := buf.ReadAt(responseData, seekOffset)
	if err != nil {
		log.Fatalf("Failed to read data from the SessionBuffer with (BufferSize:%d bytes Offset:%d seekOffset:%d). Got Error: %s \n", dataLen, app.echoBufferOffset, seekOffset, err)
		return NewErrorResponse(ErrorCodeProtoErrorMask), nil
	}
	if moreToSend && app.echoBufferOffset < dataLen {
		app.echoBufferOffset += MaxDataLength
	}
	return NewDataResponse(request, responseData, moreToSend), nil
}
