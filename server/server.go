package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/nndi-oss/mussdproxy/mudcp"
	"github.com/valyala/fasthttp"
)

/*
* udcpServer
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
type udcpServer struct {
	sessionMu         sync.Mutex     // mutex for the buffer store
	sessionStore      *SessionBuffer // session buffer for buffering request data
	listener          *http.Server
	config            *Config
	app               *coreApplication
	activeRequestChan chan (struct{})
	requestTimeout    int // Seconds for the request to timeout

}

// The Core Application is an application that enables configuring the server,
// choosing applications and controlling the session. The core application
// is like a middleware that handles requests and then forwards them to the
// currently active application depending on the Client request.
type coreApplication struct {
	availableApplications []UdcpApplication // currently registered/active application
	activeApplication     *UdcpApplication  // currently registered/active application
}

// ListenAndServe starts the UDCP server on a specific address running a specified application
func (s *udcpServer) ListenAndServe(addr string, application UdcpApplication) error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env")
		return err
	}
	return fasthttp.ListenAndServe(addr, createHandler(application))
}
