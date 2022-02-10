package ussd

import (
	"io"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/valyala/fasthttp"
)

// UssdResponseWriter writes a ussd response to the given io.Writer
// The ussd response is written out in the format the the connected
// system supports
type UssdResponseWriter interface {
	GetContentType() string

	// Write writes a UdcpResponse to the given io.Writer
	Write(ussdproxy.UdcpResponse, io.Writer) (int, error)

	WriteEnd(ussdproxy.UdcpResponse, io.Writer) (int, error)
}

type UssdRequestReader interface {
	// Read reads a UdcpRequest from the given request context
	Read(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error)
	// TODO?: Read(io.Reader) (ussdproxy.UdcpRequest, error)
}
