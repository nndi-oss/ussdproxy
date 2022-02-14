package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nndi-oss/ussdproxy/app/echo"
	"github.com/nndi-oss/ussdproxy/config"
	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/nndi-oss/ussdproxy/session/boltdb"
	"github.com/nndi-oss/ussdproxy/ussd"
	"github.com/nndi-oss/ussdproxy/ussd/africastalking"
	"github.com/valyala/fasthttp"
)

// UssdProxyServer manages overall resources on the server side
//
// * handles ussd / http requests
//
// * converts ussd / http requests to udcp requests
//
// * passes udcp requests to registered applications
//
// * waits for app to process response
//
// * responds via http after converting a udcp response
//
// * handles application errors
//
// * has registered ussd marshallers/unmarshallers (ussd.Writer, ussd.Reader)
//
// * should support registering all supported handlers
//
// * provides a UI for management/statistics?
type UssdProxyServer struct {
	sessionMu sync.Mutex // mutex for the buffer store

	app            *ussdproxy.MultiplexingApplication
	requestTimeout int // Seconds for the request to timeout

	ussdReader ussd.UssdRequestReader
	ussdWriter ussd.UssdResponseWriter

	Session ussdproxy.Session // session buffer for buffering request data
	Config  *config.Config
}

func NewUssdProxyServer() *UssdProxyServer {
	at := &africastalking.AfricasTalkingUssdHandler{}

	return &UssdProxyServer{
		app:            ussdproxy.NewMultiplexingApplication(echo.NewEchoApplication()),
		requestTimeout: 5_000,
		ussdReader:     at,
		ussdWriter:     at,
		Session:        boltdb.GetOrCreateSession("test"),
	}
}

func ListenAndServe(addr string, app ussdproxy.UdcpApplication) error {
	s := NewUssdProxyServer()
	return fasthttp.ListenAndServe(addr, s.FastHttpHandler(app))
}

type healthcheck struct {
	Status string `json:"status"`
}

func healthy() *healthcheck {
	return &healthcheck{
		Status: "HEALTHY",
	}
}
func unhealthy() *healthcheck {
	return &healthcheck{
		Status: "UNHEALTHY",
	}
}

func (s *UssdProxyServer) parseUssdRequest(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {
	return s.ussdReader.Read(ctx)
}

func (s *UssdProxyServer) ListenAndServe(addr string) {
	app := ussdproxy.NewMultiplexingApplication(echo.NewEchoApplication())
	fasthttp.ListenAndServe(addr, s.FastHttpHandler(app))
}

func (s *UssdProxyServer) FastHttpHandler(app ussdproxy.UdcpApplication) fasthttp.RequestHandler {
	// TODO: review how sessions are handled at a global level
	s.sessionMu.Lock()
	app.UseSession(s.Session) // use the session from the server
	s.sessionMu.Unlock()

	return s.useFastHttpHandler
}

func (s *UssdProxyServer) useFastHttpHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if strings.HasPrefix(path, "/healthcheck") {
		b, err := json.Marshal(healthy())
		if err != nil {
			fmt.Println(fmt.Errorf("failed to marshal healthcheck. Error: %v", err))
			ctx.WriteString(unhealthy().Status)
			return
		}
		ctx.Write(b)
		ctx.SetContentType("application/json; charset=utf-8")
		return
	}
	method := string(ctx.Method())
	if strings.Compare(method, "POST") != 0 {
		ctx.SetStatusCode(405)
		return
	}

	fmt.Println("Received ussd request")
	request, err := s.parseUssdRequest(ctx)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to parse ussd request, got %v", err))
		// TODO: should this be a protocol error?
		s.ussdWriter.WriteEnd(ussdproxy.NewProtocolErrorResponse(), ctx)
		return
	}

	done := make(chan struct{})
	requestDurationCtx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*80))
	defer cancel()

	select {
	case <-done:
		ussdAction := ussdproxy.UssdContinue
		ctx.SetContentType(s.ussdWriter.GetContentType())
		response, err := ussdproxy.ProcessUdcpRequest(request, s.app)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to process request, got %v", err))
			// TODO: wrap the error according to the type
			s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ErrorNotAsciiPduType), ctx)
			close(done)
			return
		}

		if response == nil {
			fmt.Println(fmt.Errorf("failed to process request, got %v response", err))
			// TODO: should this be a protocol error?
			s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ErrorPduType), ctx)
			close(done)
			return
		}

		if response.IsErrorPdu() || response.IsReleaseDialoguePdu() {
			ussdAction = ussdproxy.UssdEnd
		}
		if ussdAction == ussdproxy.UssdEnd {
			s.ussdWriter.WriteEnd(response, ctx)
		} else {
			s.ussdWriter.Write(response, ctx)
		}

		close(done)

	case <-requestDurationCtx.Done():
		fmt.Println("timeout exceeded for ussd request", request)
		s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ReleaseDialogPduType), ctx)
	}

}
