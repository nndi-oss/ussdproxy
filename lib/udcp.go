package ussdproxy

import (
	"io"
	"unicode"
	"unicode/utf8"
)

// UdcpHeader is the header information in the UdcpRequest
type UdcpHeader struct {
	Type       PduType
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

func isASCII(s []byte) bool {
	if len(s) == 1 {
		_, size := utf8.DecodeRune(s)
		return size < 2
	}

	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}

	return true
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
	return ProtocolVersion
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
			Type:       PduType(responseType),
			Version:    ProtocolVersion,
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
			Version:    ProtocolVersion,
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
			Version:    ProtocolVersion,
		},
		data: []byte(NoDataResponse),
		len:  len(NoDataResponse),
	}
}

// NewDataRequest returns a UdcpRequest
func NewDataRequest(data []byte, moreToSend bool) UdcpRequest {
	if !isASCII(data) {
		return NewErrorResponse(ErrorNotAsciiPduType)
	}

	typ := DataLongPduType
	if moreToSend {
		typ = DataPduWithMtsType
	}
	return &udcpRequest{
		header: &UdcpHeader{
			Type:       typ,
			Version:    ProtocolVersion,
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
			Version:    ProtocolVersion,
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
func NewReleaseDialogueResponse(reason PduType) UdcpResponse {
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       reason,
			Version:    ProtocolVersion,
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
func NewErrorResponse(errorCode PduType) UdcpResponse {
	// TODO: Where to put error code?
	return &udcpResponse{
		header: &UdcpHeader{
			Type:       errorCode,
			Version:    ProtocolVersion,
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
	return ProtocolVersion
}
