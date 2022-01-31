package ussd

import (
	"bytes"
	"encoding/xml"
	"io"
)

const (
	TRUROUTE_INITIAL  = 1
	TRUROUTE_RESPONSE = 2
	TRUROUTE_RELEASE  = 3
)

// TruRouteRequest XML struct for the request from a truroute services
//
// See: https://github.com/saulchelewani/ussd/
type TruRouteRequest struct {
	XMLName xml.Name `xml:"ussd"`
	Type    int      `xml:"type"`
	Message string   `xml:"msg"`
	Session string   `xml:"sessionid"`
	Msisdn  string   `xml:"msisdn"`
}

// TruRouteResponse XML struct for the response from a truroute services
//
// See: https://github.com/saulchelewani/ussd/blob/master/src/UssdServiceProvider.php
type TruRouteResponse struct {
	XMLName xml.Name `xml:"ussd"`
	Type    int      `xml:"type"`
	Message string   `xml:"msg"`
	Premium struct {
		Cost int    `xml:"cost"`
		Ref  string `xml:"ref"`
	} `xml:"premium"`
	Msisdn string `xml:"msisdn"`
}

func (t *TruRouteResponse) isResponse() bool {
	return t.Type == TRUROUTE_RESPONSE
}

func (t *TruRouteResponse) isRelease() bool {
	return t.Type == TRUROUTE_RELEASE
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
	if _, err := buf.Write(u.Message); err != nil {
		return nil
	}
	return buf.Bytes()
}

func ParseTruRouteUssdRequest(ussdRequest *TruRouteRequest) (UdcpRequest, error) {
	ussdData := ussdRequest.Bytes()
	typ := ussdData[0]
	len := int(ussdData[1])
	moreToSend := false
	if typ == DataPduWithMtsType {
		moreToSend = true
	}
	// skip the third byte since it's a separator
	data := make([]byte, len)
	r := bytes.NewReader(ussdData)
	r.Seek(3, 0)
	// Parse the header and data here
	if _, err := io.ReadAtLeast(r, data, len); err != nil {
		return nil, InvalidHeaderError
	}
	return &udcpRequest{
		header: &UdcpHeader{
			Type:       typ,
			Version:    UdcpVersion,
			MoreToSend: moreToSend,
		},
		data: data,
		len:  len,
	}, nil
}
