package server

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

// ListenAndServe starts the UDCP server on a specific address running a specified application
func ListenAndServe(s *udcpServer, addr string, application UdcpApplication) error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env")
		return err
	}
	return s.ListenAndServe(addr, createHandler(application))
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

func createHandler(app UdcpApplication) func(ctx *fasthttp.RequestCtx) {
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

		if ctx.FormValue("text") == nil || ctx.FormValue("phoneNumber") == nil ||
			ctx.FormValue("sessionId") == nil {
			ctx.SetStatusCode(400)
			return
		}
		var request *UssdRequest
		var response UdcpResponse
		request = &UssdRequest{
			SessionID:   string(ctx.FormValue("sessionId")),
			PhoneNumber: string(ctx.FormValue("phoneNumber")),
			Data:        ctx.FormValue("text"), // Value is named text but we expect binary data here
			Channel:     string(ctx.FormValue("channel")),
		}
		response, err := ProcessUssdRequest(request, app)
		if err != nil {
			ctx.SetContentType("text/plain; charset=ascii")
			fmt.Print(ctx, UssdEnd)
			return
		}

		if response != nil {
			ussdAction := UssdContinue
			if response.IsErrorPdu() || response.IsReleaseDialoguePdu() {
				ussdAction = UssdEnd
			}
			//response.Write(ctx)
			ctx.WriteString(fmt.Sprintf("%s %s", ussdAction, response.ToString()))
		}
		ctx.SetContentType("text/plain; charset=ascii")
	}
}
