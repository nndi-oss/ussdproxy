package flares

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/valyala/fasthttp"
)

// FlaresRequest struct represents a request coming in from the Network routed from a Flares services USSD API
//
// See: https://github.com/saulchelewani/ussd/blob/master/src/Http/Flares/FlaresRequest.php
type FlaresRequest struct {
	XMLName xml.Name `xml:"request"`
	Message string   `xml:"subscriberInput"`
	Session string   `xml:"sessionId"`
	Msisdn  string   `xml:"msisdn"`
}

// FlaresResponse XML struct for the response from a Flares services
//
// See: https://github.com/saulchelewani/ussd/blob/master/src/Http/Flares/FlaresResponse.php
type FlaresResponse struct {
	XMLName xml.Name `xml:"response"`
	Message string   `xml:"applicationResponse"`
	Session string   `xml:"sessionId"`
	Msisdn  string   `xml:"msisdn"`
}

func (t *FlaresResponse) GetText() string {
	return t.Message
}

// Bytes Converts the UssdRequest to a byte array with the data is separated by 0x0
func (u *FlaresRequest) Bytes() []byte {
	var buf bytes.Buffer
	if _, err := buf.Write([]byte(u.Session)); err != nil {
		return nil
	}
	if _, err := buf.Write([]byte(u.Msisdn)); err != nil {
		return nil
	}
	if _, err := buf.Write([]byte(u.Message)); err != nil {
		return nil
	}
	return buf.Bytes()
}

type FlaresUssdHandler struct{}

func (u *FlaresUssdHandler) Read(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {
	requestData := ctx.Request.Body()
	if requestData == nil {
		return nil, errors.New("flares: request was empty or nil")
	}

	var trRequest *FlaresRequest
	err := xml.Unmarshal(requestData, &trRequest)
	if err != nil {
		return nil, err
	}

	if trRequest.Msisdn == "" || trRequest.Session == "" {
		ctx.SetStatusCode(400)
		return nil, fmt.Errorf("invalid request, got body: %s", string(ctx.Request.Body()))
	}

	return parseUssdRequest(&UssdRequest{
		SessionID:   string(trRequest.Session),
		PhoneNumber: string(trRequest.Msisdn),
		Data:        []byte(trRequest.Message),
		Channel:     string(trRequest.Msisdn),
	})
}

func (u *FlaresUssdHandler) GetContentType() string {
	return "text/xml; charset=ascii"
}

func (u *FlaresUssdHandler) Write(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	trResponse := &FlaresResponse{
		Message: string(response.Data()),
		Msisdn:  "",
	}
	data, err := xml.Marshal(trResponse)
	if err != nil {
		return -1, err
	}
	return writer.Write(data)
}

func (u *FlaresUssdHandler) WriteEnd(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	trResponse := &FlaresResponse{
		Message: string(response.Data()),
		Msisdn:  "",
	}
	data, err := xml.Marshal(trResponse)
	if err != nil {
		return -1, err
	}
	return writer.Write(data)
}

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
	ussdData := ussdRequest.Data
	typ := ussdproxy.RequestPduType(string(ussdData[0:2]))
	if typ == ussdproxy.InvalidPduType {
		return nil, fmt.Errorf("%v got '%v'", ussdproxy.ErrInvalidHeader, typ)
	}
	//len := len(ussdData)
	moreToSend := typ.HasMoreToSend()
	// TODO: add a generic NewRequest func to account for the type
	return ussdproxy.NewDataRequest(ussdData[2:], moreToSend), nil
}
