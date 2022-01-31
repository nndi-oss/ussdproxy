package ussd

import (
	"github.com/nndi-oss/ussdproxy/udcp"
	"io"
)

// UssdResponseWriter writes a ussd response to the given io.Writer
// The ussd response is written out in the format the the connected
// system supports
type UssdResponseWriter interface {
	// Write writes a UdcpResponse to the given io.Writer
	Write(udcp.UdcpResponse, io.Writer) (int, error)
}

type UssdRequestReader interface {
	// Read reads a UdcpRequest from the given io.Reader
	Read(io.Reader) (udcp.UdcpRequest, error)
}
