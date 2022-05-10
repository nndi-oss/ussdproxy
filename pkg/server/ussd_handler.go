package server

import (
	"encoding/json"
	"fmt"
	"strings"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/valyala/fasthttp"
)

func (s *UssdProxyServer) ussdCallbackHandler(ctx *fasthttp.RequestCtx) {
	// TODO: review how sessions are handled at a global level
	s.sessionMu.Lock()
	s.app.UseSession(s.Session) // use the session from the server
	s.sessionMu.Unlock()
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

	request, err := s.parseUssdRequest(ctx)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to parse ussd request, got %v", err))
		// TODO: should this be a protocol error?
		s.ussdWriter.WriteEnd(ussdproxy.NewProtocolErrorResponse(), ctx)
		return
	}
	fmt.Println("Processing request ", request)
	// done := make(chan struct{})
	// requestDurationCtx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	// defer cancel()

	// select {
	// case <-done:
	ussdAction := ussdproxy.UssdContinue
	ctx.SetContentType(s.ussdWriter.GetContentType())
	response, err := ussdproxy.ProcessUdcpRequest(request, s.app)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to process request, got %v", err))
		// TODO: wrap the error according to the type
		s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ErrorNotAsciiPduType), ctx)
		// close(done)
		return
	}

	if response == nil {
		fmt.Println(fmt.Errorf("failed to process request, got %v response", err))
		// TODO: should this be a protocol error?
		s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ErrorPduType), ctx)
		// close(done)
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

	// close(done)

	// case <-requestDurationCtx.Done():
	// 	fmt.Println("timeout exceeded for ussd request", request)
	// 	s.ussdWriter.WriteEnd(ussdproxy.NewErrorResponse(ussdproxy.ReleaseDialogPduType), ctx)
	// }

}
