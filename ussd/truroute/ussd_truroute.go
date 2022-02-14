package truroute

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
	"github.com/valyala/fasthttp"
)

const (
	ussdInitialRequestCode   = 1
	ussdContinueResponseCode = 2
	ussdReleaseResponseCode  = 3
)

// TruRouteRequest struct represents a request coming in from the Network routed from a truroute services USSD API
//
// See: https://github.com/saulchelewani/ussd/blob/master/src/Http/TruRoute/TruRouteRequest.php
type TruRouteRequest struct {
	XMLName xml.Name `xml:"ussd"`
	Type    int      `xml:"type"`
	Message string   `xml:"msg"`
	Session string   `xml:"sessionid"`
	Msisdn  string   `xml:"msisdn"`
}

// TruRouteResponse XML struct for the response from a truroute services
//
// See: https://github.com/saulchelewani/ussd/blob/master/src/Http/TruRoute/TruRouteResponse.php
type TruRouteResponse struct {
	XMLName xml.Name                `xml:"ussd"`
	Type    int                     `xml:"type"`
	Message string                  `xml:"msg"`
	Premium TruRouteResponsePremium `xml:"premium"`
	Msisdn  string                  `xml:"msisdn"`
}

type TruRouteResponsePremium struct {
	Cost int    `xml:"cost"`
	Ref  string `xml:"ref"`
}

func (t *TruRouteResponse) isResponse() bool {
	return t.Type == ussdContinueResponseCode
}

func (t *TruRouteResponse) isRelease() bool {
	return t.Type == ussdReleaseResponseCode
}

func (t *TruRouteResponse) GetText() string {
	return t.Message
}

// Bytes Converts the UssdRequest to a byte array with the data is separated by 0x0
func (u *TruRouteRequest) Bytes() []byte {
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

type TrurouteUssdHandler struct{}

func (u *TrurouteUssdHandler) Read(ctx *fasthttp.RequestCtx) (ussdproxy.UdcpRequest, error) {
	requestData := ctx.Request.Body()
	if requestData == nil {
		return nil, errors.New("truroute: request was empty or nil")
	}

	var trRequest *TruRouteRequest
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

func (u *TrurouteUssdHandler) GetContentType() string {
	return "text/xml; charset=ascii"
}

func (u *TrurouteUssdHandler) Write(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	trResponse := &TruRouteResponse{
		Type:    ussdContinueResponseCode,
		Message: string(response.Data()),
		Premium: TruRouteResponsePremium{Cost: 0, Ref: ""},
		Msisdn:  "",
	}
	data, err := xml.Marshal(trResponse)
	if err != nil {
		return -1, err
	}
	return writer.Write(data)
}

func (u *TrurouteUssdHandler) WriteEnd(response ussdproxy.UdcpResponse, writer io.Writer) (int, error) {
	trResponse := &TruRouteResponse{
		Type:    ussdContinueResponseCode,
		Message: string(response.Data()),
		Premium: TruRouteResponsePremium{Cost: 0, Ref: ""},
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
