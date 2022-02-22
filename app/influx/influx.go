package influx

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	lineprotocol "github.com/influxdata/line-protocol"
	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
)

const (
	FormatPipeDelimited = iota
	FormatSENML
	FormatJSON
)

// InfluxDbApp provides an application that sends metrics to influxdb
type InfluxDbApp struct {
	ussdproxy.UdcpApplication

	currentConn net.Conn
	InputFormat uint8
	Addr        string
	Username    string
	Password    string
	Database    string
	session     ussdproxy.Session
}

func NewInfluxApp(addr, database, user, password string) *InfluxDbApp {
	conn1, err := net.Dial("tcp", addr) // test the conneciotn
	if err != nil {
		panic(fmt.Errorf("failed to connect to %s got: %v ", addr, err))
	}
	defer conn1.Close()

	return &InfluxDbApp{
		Database:    database,
		Username:    user,
		Password:    password,
		Addr:        addr,
		InputFormat: FormatPipeDelimited,
	}
}

func (app *InfluxDbApp) GetOrCreateSession() ussdproxy.Session {
	return app.session
}

func (app *InfluxDbApp) UseSession(session ussdproxy.Session) {
	app.session = session
}

func extractPipeDelimitedTagsAndFields(data []byte) (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	// e.g. temp:38.29|lat:-35.012|lng:12.12|tag_meter_no:ABCD|ts:145923830|tag_host:2981
	// for each string starting with tag use as tag,
	// use everything else as a field
	for _, entry := range strings.Split(string(data), "|") {
		keyValueItem := strings.Split(entry, ":")
		if len(keyValueItem) == 2 {
			// TODO(zikani03): fix where the __NODATA__ is coming from
			keyPart := strings.ReplaceAll(keyValueItem[0], "__NODATA__", "")
			valuePart := keyValueItem[1]

			if keyPart == "ts" {
				if timestampValue, err := strconv.ParseInt(valuePart, 10, 64); err == nil {
					fields["ts"] = timestampValue
				}
			}

			if intValue, err := strconv.ParseInt(valuePart, 10, 64); err == nil {
				fields[keyPart] = intValue
			} else if floatValue, err := strconv.ParseFloat(valuePart, 64); err == nil {
				fields[keyPart] = floatValue
			} else {
				// not an integer or float so we just assume it's a tag now
				tags[keyPart] = valuePart
			}
		}
	}

	return tags, fields
}

func (app *InfluxDbApp) onDataWriteToInflux(data []byte) error {
	var tags map[string]string
	var fields map[string]interface{}

	if app.InputFormat == FormatPipeDelimited {
		tags, fields = extractPipeDelimitedTagsAndFields(data)
		fields["ts"] = time.Now().UnixMilli()
		tags["app"] = "ussdproxy"
	}

	conn, err := net.Dial("tcp", app.Addr)
	if err != nil {
		log.Fatalf("failed to connect %v", err)
		return err
	}
	serializer := lineprotocol.NewEncoder(conn)
	serializer.SetMaxLineBytes(1024)
	serializer.SetFieldTypeSupport(lineprotocol.UintSupport)
	serializer.FailOnFieldErr(true)
	// serializer.SetPrecision(time.Millisecond)

	event, err := lineprotocol.New(app.Database, tags, fields, time.Now())
	if err != nil {
		return err
	}

	_, err = serializer.Encode(event)
	if err != nil {
		return fmt.Errorf("failed to send data got: %v", err)
	}
	err = conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection got %v", err)
	}
	return nil
}

// Name the name of the application
func (app *InfluxDbApp) Name() string {
	return "InfluxDB Forwarder"
}

// ApplicationID the unique identifier for the application
func (app *InfluxDbApp) ApplicationID() string {
	return "influxdb"
}

// Author the author of the application
func (app *InfluxDbApp) Author() string {
	return "NNDI"
}

// Register the InfluxDbApp with the server
func (app *InfluxDbApp) Register() {
	// noop
}

// OnError returns the request/response handler for the Echo Application
func (app *InfluxDbApp) OnError(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	fmt.Printf("Received ErrorPdu, %s", request.Data())
	return ussdproxy.NewProtocolErrorResponse(), nil
}

// OnData returns the request/response handler for the Echo Application
func (app *InfluxDbApp) OnData(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	// if request.HasMoreToSend() {
	// 	// Waiting for Client to send more data
	// 	return ussdproxy.NewReceiveReadyResponse(), nil
	// }
	// Handle the decoding of the data here
	//
	// This is the point at which you may send data to an external service
	// since at this point all the data the client intended to send is complete
	data, err := session.RecvBuffer().Read()
	if err != nil {
		return ussdproxy.NewProtocolErrorResponse(), nil
	}
	if err = app.onDataWriteToInflux(data); err != nil {
		return ussdproxy.NewProtocolErrorResponse(), nil
	}
	// We're ready to receive more data
	return ussdproxy.NewReceiveReadyResponse(), nil
}

// OnReceiveReady returns data when a Client is waiting for server data
func (app *InfluxDbApp) OnReceiveReady(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	return ussdproxy.NewReceiveReadyResponse(), nil
}

// OnReleaseDialogue returns the request/response handler for the Echo Application
func (app *InfluxDbApp) OnReleaseDialogue(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	return ussdproxy.NewUserAbortReleaseDialogueResponse(), nil
}
