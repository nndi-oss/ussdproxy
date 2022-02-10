package ussdproxy

import (
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

// Session provides read and write buffers for a specific Session
type Session interface {
	SessionID() string
	RecvBuffer() SessionBuffer
	SendBuffer() SessionBuffer
	IsOpen() bool
	Close()
	Reset()
	Commit()
}

// SessionBuffer is a read/write buffer
type SessionBuffer interface {
	Read() ([]byte, error)
	ReadAt(p []byte, offset int64) (int, error)
	Write(data []byte) error
	Set([]byte) error
	FillWith(SessionBuffer) error
	Purge()
	IsEmpty() bool
	Length() int
}
