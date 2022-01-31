package core

import (
	"fmt"
	"io"
)

// UdcpRequestHandler handler for a completed UdcpRequest
type UdcpRequestHandler func(UdcpRequest, Session) (UdcpResponse, error)

type UdcpData interface {
	Header() *UdcpHeader
	Data() []byte
	Len() int
	HasMoreToSend() bool
	IsDataPdu() bool
	IsReceiveReadyPdu() bool
	IsReleaseDialoguePdu() bool
	IsErrorPdu() bool
	Version() uint8
	ToString() string
}

// UdcpRequest represents a request from a USSD interaction (from the client)
type UdcpRequest interface {
	UdcpData
	//UssdRequest() UssdRequestInterface
}

// UdcpResponse is the server's response for a particular UdcpRequest
type UdcpResponse interface {
	UdcpData
	Request() UdcpRequest
	SetHeader(header *UdcpHeader) error
	SetData(data []byte) error
	Write(w io.Writer) error
}

// UssdRequestInterface struct represents a request coming in from the Network
type UssdRequestInterface interface {
	PhoneNumber() string
	Data() []byte
	RawText() string
	SessionID() string
	Channel() string
	Provider() string
}

// UdcpHeader is the header information in the UdcpRequest
type UdcpHeader struct {
	Type       pduType
	Version    uint8
	MoreToSend bool
}

type udcpRequest struct {
	// ussdRequest *UssdRequest
	header *UdcpHeader
	len    int
	data   []byte
}

type udcpResponse struct {
	header  *UdcpHeader
	request UdcpRequest
	len     int
	data    []byte
}

// ProcessUssdRequest is the main procedure to process ussd request into UdcpResponse
func ProcessUssdRequest(ussdRequest *UssdRequest, application UdcpApplication) (UdcpResponse, error) {
	udcpReq, err := ParseUssdRequest(ussdRequest)
	if err != nil {
		return nil, err
	}
	return ProcessUdcpRequest(udcpReq, application)
	// fmt.Printf("Finished parsing UdcpRequest: %s\n", udcpReq)
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
		return NewErrorResponse(ErrorCodeProtoErrorMask), UnknownParseError
	}
	session := GetOrCreateSession(udcpReq.SessionID)
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

// Implementation for UdcpRequest

func (req *udcpRequest) Header() *UdcpHeader {
	return req.header
}

func (req *udcpRequest) Data() []byte {
	return req.data
}
func (req *udcpRequest) Len() int {
	return req.len
}
func (req *udcpRequest) HasMoreToSend() bool {
	return req.Header().MoreToSend
}
func (req *udcpRequest) IsDataPdu() bool {
	return req.Header().Type == DataLongPduType || req.Header().Type == DataPduWithMtsType
}
func (req *udcpRequest) IsReceiveReadyPdu() bool {
	return req.Header().Type == ReceiveReadyPduType
}

func (req *udcpRequest) IsReleaseDialoguePdu() bool {
	typ := req.Header().Type
	return typ == ReleaseCodeUnknownMask ||
		typ == ReleaseCodeUserAbortMask ||
		typ == ReleaseCodeUssdTimeoutMask ||
		typ == ReleaseCodeIdleDialogMask
}

func (req *udcpRequest) IsErrorPdu() bool {
	return req.Header().Type == ErrorPduType
}

func (req *udcpRequest) Version() uint8 {
	return UdcpVersion
}

func (req *udcpRequest) String() string {
	return req.ToString()
}

func (req *udcpRequest) ToString() string {
	data := make([]byte, 0)
	data = append(data, []byte(req.header.Type.String())...)
	data = append(data, byte(len(req.data)))
	data = append(data, 0x00)
	data = append(data, req.data...)
	return string(data)
}

func (res *udcpResponse) ToString() string {
	data := make([]byte, 0)
	data = append(data, []byte(res.header.Type.String())...)
	data = append(data, byte(len(res.data)))
	data = append(data, 0x00)
	data = append(data, res.data...)
	return string(data)
}

// NewUdcpResponse returns a UdcpResponse
func NewUdcpResponse(request UdcpRequest, responseType uint8, moreToSend bool, data []byte) UdcpResponse {
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       pduType(responseType),
			Version:    UdcpVersion,
			MoreToSend: moreToSend,
		},
		request: request,
		data:    data,
		len:     len(data),
	}
}

// NewUdcpResponse returns a UdcpResponse
func NewDataResponse(request UdcpRequest, data []byte, moreToSend bool) UdcpResponse {
	typ := DataLongPduType
	if moreToSend {
		typ = DataPduWithMtsType
	}
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       typ,
			Version:    UdcpVersion,
			MoreToSend: moreToSend,
		},
		request: request,
		data:    data,
		len:     len(data),
	}
}

// NewReceiveReadyRequest returns a UdcpRequest with a ReceiveReady type
func NewReceiveReadyRequest() UdcpRequest {
	return &udcpRequest{
		header: &UdcpHeader{
			MoreToSend: false,
			Type:       ReceiveReadyPduType,
			Version:    UdcpVersion,
		},
		data: []byte(NoDataResponse),
		len:  len(NoDataResponse),
	}
}

// NewDataRequest returns a UdcpRequest
func NewDataRequest(data []byte, moreToSend bool) UdcpRequest {
	typ := DataLongPduType
	if moreToSend {
		typ = DataPduWithMtsType
	}
	return &udcpRequest{
		header: &UdcpHeader{
			Type:       typ,
			Version:    UdcpVersion,
			MoreToSend: moreToSend,
		},
		data: data,
		len:  len(data),
	}
}

// NewReceiveReadyResponse returns a UdcpResponse with a ReceiveReady type
func NewReceiveReadyResponse() UdcpResponse {
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       ReceiveReadyPduType,
			Version:    UdcpVersion,
			MoreToSend: true,
		},
		request: nil,
		data:    []byte(NoDataResponse),
		len:     0,
	}
}

func NewUserAbortReleaseDialogueResponse() UdcpResponse {
	return NewReleaseDialogueResponse(ReleaseCodeUserAbortMask)
}

// NewReleaseDialogueResponse returns a UdcpResponse with a ReceiveReady type
func NewReleaseDialogueResponse(reason pduType) UdcpResponse {
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       reason,
			Version:    UdcpVersion,
			MoreToSend: false,
		},
		request: nil,
		data:    []byte(NoDataResponse),
		len:     0,
	}
}

func NewProtocolErrorResponse() UdcpResponse {
	return NewErrorResponse(ErrorCodeProtoErrorMask)
}

// NewErrorResponse returns an error response
func NewErrorResponse(errorCode pduType) UdcpResponse {
	// TODO: Where to put error code?
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       errorCode,
			Version:    UdcpVersion,
			MoreToSend: false,
		},
		request: nil,
		data:    []byte("Unknown Error"),
		len:     0,
	}
}

func (res *udcpResponse) Write(w io.Writer) error {
	_, err := w.Write([]byte(res.ToString()))
	return err
}

func (res *udcpResponse) Request() UdcpRequest {
	return res.request
}

func (res *udcpResponse) SetHeader(header *UdcpHeader) error {
	res.header = header
	return nil
}

func (res *udcpResponse) SetData(data []byte) error {
	if len(data) > MaxUssdLength {
		res.header.MoreToSend = true
	}
	res.data = data
	return nil
}

func (res *udcpResponse) Header() *UdcpHeader {
	return res.header
}

func (res *udcpResponse) Data() []byte {
	return res.data
}
func (res *udcpResponse) Len() int {
	return res.len
}
func (res *udcpResponse) HasMoreToSend() bool {
	return res.Header().MoreToSend
}
func (res *udcpResponse) IsDataPdu() bool {
	return res.Header().Type == DataLongPduType || res.Header().Type == DataPduWithMtsType
}
func (res *udcpResponse) IsReceiveReadyPdu() bool {
	return res.Header().Type == ReceiveReadyPduType
}
func (res *udcpResponse) IsReleaseDialoguePdu() bool {
	typ := res.Header().Type
	return typ == ReleaseCodeUnknownMask ||
		typ == ReleaseCodeUserAbortMask ||
		typ == ReleaseCodeUssdTimeoutMask ||
		typ == ReleaseCodeIdleDialogMask
}

func (res *udcpResponse) IsErrorPdu() bool {
	return res.Header().Type == ErrorPduType
}

func (res *udcpResponse) Version() uint8 {
	return UdcpVersion
}

func parsePduType() int {
	return -1
}

func parseAsciiHeaderToType(header string) int {
	return -1
}
