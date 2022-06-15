package africastalking

import (
	"bytes"
	"fmt"
	"io"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/valyala/fasthttp"
)

type AfricasTalkingUssdHandler struct{}

func New() *AfricasTalkingUssdHandler {
	return &AfricasTalkingUssdHandler{}
}

func (u *AfricasTalkingUssdHandler) Read(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {

	requestData := ctx.FormValue("text")
	if string(requestData) == "" {
		requestData = []byte("R;__NODATA__") // if the request is empty, we default to a receive-ready
	}

	if string(ctx.FormValue("phoneNumber")) == "" || string(ctx.FormValue("sessionId")) == "" {
		ctx.SetStatusCode(400)
		return nil, fmt.Errorf("invalid request, got body: %s", string(ctx.Request.Body()))
	}

	return parseUssdRequest(&UssdRequest{
		SessionID:   string(ctx.FormValue("sessionId")),
		PhoneNumber: string(ctx.FormValue("phoneNumber")),
		Data:        requestData,
		Channel:     string(ctx.FormValue("channel")),
	})
}

func (u *AfricasTalkingUssdHandler) GetContentType() string {
	return "text/plain; charset=ascii"
}

func (u *AfricasTalkingUssdHandler) Write(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	writer.Write([]byte("CON "))
	return writer.Write(response.Data())
}

func (u *AfricasTalkingUssdHandler) WriteEnd(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	writer.Write([]byte("END "))
	return writer.Write(response.Data())
}

// UssdRequest struct represents a request coming in from the Network
// routed through AfricasTalking's USSD API
type UssdRequest struct {
	PhoneNumber string `json:"name=phoneNumber"`
	Data        []byte `json:"name=text"`
	SessionID   string `json:"name=sessionId"`
	Channel     string `json:"name=channel,omitempty"`
}

// Bytes Converts the UssdRequest to a byte array with the data is separated by 0x0
func (u *UssdRequest) Bytes() []byte {
	var buf bytes.Buffer
	if _, err := buf.Write([]byte(u.SessionID)); err != nil {
		return nil
	}
	if _, err := buf.Write([]byte(u.PhoneNumber)); err != nil {
		return nil
	}
	if _, err := buf.Write(u.Data); err != nil {
		return nil
	}
	return buf.Bytes()
}

func parseUssdRequest(ussdRequest *UssdRequest) (ussdproxy.UdcpRequest, error) {
	ussdData := string(ussdRequest.Data)
	typ := ussdproxy.RequestPduType(ussdData)
	if typ == ussdproxy.InvalidPduType {
		return nil, fmt.Errorf("%v got '%v'", ussdproxy.ErrInvalidHeader, typ)
	}
	//len := len(ussdData)
	moreToSend := typ.HasMoreToSend()
	// TODO: add a generic NewRequest func to account for the type
	return ussdproxy.NewDataRequest([]byte(ussdproxy.StripPdu(ussdData, typ)), moreToSend), nil
}
