package server

import (
	"sync"

	"github.com/fasthttp/router"
	"github.com/nndi-oss/ussdproxy/app/echo"
	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/nndi-oss/ussdproxy/pkg/config"
	"github.com/nndi-oss/ussdproxy/pkg/session/boltdb"
	"github.com/nndi-oss/ussdproxy/pkg/telemetry"
	"github.com/nndi-oss/ussdproxy/pkg/ussd"
	"github.com/valyala/fasthttp"
)

const (
	DefaultServerRequestTimeout = 5_000
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
	telemetry      *telemetry.Telemetry

	ussdReader ussd.UssdRequestReader
	ussdWriter ussd.UssdResponseWriter

	Session ussdproxy.Session // session buffer for buffering request data
	Config  *config.UssdProxyConfig
}

func NewUssdProxyServer(userConfig *config.UssdProxyConfig, application ussdproxy.UdcpApplication) *UssdProxyServer {
	if userConfig == nil {
		panic("Invalid configuration for Server") // TODO: do better
	}

	ussdProvider := userConfig.GetProvider()

	return &UssdProxyServer{
		Config:         userConfig,
		app:            ussdproxy.NewMultiplexingApplication(echo.NewEchoApplication()),
		requestTimeout: userConfig.Server.RequestTimeout,
		ussdReader:     ussdProvider,
		ussdWriter:     ussdProvider,
		Session:        boltdb.GetOrCreateSession("test"),
		telemetry:      telemetry.New(userConfig.Telemetry.BindAddress()),
	}
}

func (s *UssdProxyServer) parseUssdRequest(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {
	return s.ussdReader.Read(ctx)
}

func (s *UssdProxyServer) ListenAndServe(addr string) error {
	r := router.New()

	r.GET("/healthz", s.healthcheckHandler)
	r.GET(s.Config.Ussd.CallbackURL, s.ussdCallbackHandler)
	r.POST(s.Config.Ussd.CallbackURL, s.ussdCallbackHandler)
	// TODO: Add telemetry stuff
	r.GET("/metrics", s.notImplementedHandler)
	// Admin routes, which need to be protected btw
	r.GET("/admin/apps", s.notImplementedHandler)
	r.GET("/admin/sessions", s.notImplementedHandler)
	r.GET("/admin/sessions/active", s.notImplementedHandler)
	r.GET("/admin/sessions/closed", s.notImplementedHandler)
	r.GET("/admin/settings/udcp", s.notImplementedHandler)
	r.GET("/admin/settings/apps", s.notImplementedHandler)

	return fasthttp.ListenAndServe(addr, r.Handler)
}
