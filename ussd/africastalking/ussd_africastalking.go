package ussd

import (
	"bytes"
	"io"
)

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

func dummyUdcpRequest(req *UssdRequest) (UdcpRequest, error) {
	return &udcpRequest{
		header: &UdcpHeader{
			Type:       DataLongPduType,
			Version:    UdcpVersion,
			MoreToSend: false,
		},
		data: []byte("dummy"),
		len:  5,
	}, nil
}

func ParseUssdRequest(ussdRequest *UssdRequest) (UdcpRequest, error) {
	// fmt.Printf("Parsing ussdRequest: %s\n", ussdRequest)
	ussdData := ussdRequest.Data
	typ := pduType(ussdData[0])
	len := int(ussdData[1])
	moreToSend := typ.HasMoreToSend()
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
