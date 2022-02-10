package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/nndi-oss/ussdproxy/app/echo"
	"github.com/nndi-oss/ussdproxy/config"
	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/nndi-oss/ussdproxy/session/boltdb"
	"github.com/nndi-oss/ussdproxy/ussd"
	"github.com/nndi-oss/ussdproxy/ussd/africastalking"
	"github.com/valyala/fasthttp"
)

/*
* UssdProxyServer
- manages overall resources on the server side
- handles http requests
- converts http requests to udcp requests
- passes udcp requests to registered applications
- waits for app to process response
- responds via http after converting a udcp response
- handles applicatiion errors
- has registered ussd marshallers/unmarshallers (ussd.Writer, ussd.Reader)
- should support registering all supported handlers
- provides a UI for management/statistics?
-
* ## Properties
* - requestTimeout
* - charLengthLimit
* - encodingDetection?
*/
type UssdProxyServer struct {
	sessionMu sync.Mutex // mutex for the buffer store

	app               *coreApplication
	requestTimeout    int // Seconds for the request to timeout
	activeRequestChan chan (struct{})

	ussdReader ussd.UssdRequestReader
	ussdWriter ussd.UssdResponseWriter

	Session ussdproxy.Session // session buffer for buffering request data
	Config  *config.Config
}

// The Core Application is an application that enables configuring the server,
// choosing applications and controlling the session. The core application
// is like a middleware that handles requests and then forwards them to the
// currently active application depending on the Client request.
type coreApplication struct {
	availableApplications []*ussdproxy.UdcpApplication // currently registered/active application
	activeApplication     *ussdproxy.UdcpApplication   // currently registered/active application
}

func NewUssdProxyServer() *UssdProxyServer {
	at := &africastalking.AfricasTalkingUssdHandler{}

	return &UssdProxyServer{
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

type healthcheck struct{ status string }

func healthy() *healthcheck {
	return &healthcheck{
		status: "HEALTHY",
	}
}
func unhealthy() *healthcheck {
	return &healthcheck{
		status: "UNHEALTHY",
	}
}

func (s *UssdProxyServer) parseUssdRequest(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {
	return s.ussdReader.Read(ctx)
}

func (s *UssdProxyServer) ListenAndServe(addr string) {
	app := echo.NewEchoApplication()
	fasthttp.ListenAndServe(addr, s.FastHttpHandler(app))
}

func (s *UssdProxyServer) FastHttpHandler(app ussdproxy.UdcpApplication) fasthttp.RequestHandler {
	// TODO: review how sessions are handled at a global level
	app.UseSession(s.Session) // use the session from the server

	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if strings.HasPrefix(path, "/healthcheck") {
			b, err := json.Marshal(healthy())
			if err != nil {
				fmt.Errorf("Failed to marshal healthcheck. Error: %s", err)
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

		ussdAction := ussdproxy.UssdContinue

		ctx.SetContentType(s.ussdWriter.GetContentType())

		response, err := ussdproxy.ProcessUdcpRequest(request, app)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to process request, got %v", err))
			// TODO: should this be a protocol error?
			s.ussdWriter.WriteEnd(ussdproxy.NewProtocolErrorResponse(), ctx)
			return
		}

		if response == nil {
			fmt.Println(fmt.Errorf("failed to process request, got a nil response", err))
			// TODO: should this be a protocol error?
			s.ussdWriter.WriteEnd(ussdproxy.NewProtocolErrorResponse(), ctx)
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
	}
}
