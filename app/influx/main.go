package influx

import (
	"flag"
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	udcp "github.com/nndi-oss/ussdproxy"
)

const (
	INPUT_FORMAT_PIPE_DELIMITED = 0x1
	INPUT_FORMAT_SENML          = 0x2
	INPUT_FORMAT_JSON           = 0x3
)

var (
	addr              = flag.String("addr", ":8327", "TCP Address to listen to")
	influxApplication *InfluxDbApp
)

func main() {
	flag.Parse()
	// main server for udcp
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: "labs",
		Password: "labs123",
	})
	if err != nil {
		log.Fatal("Failed to initialize InfluxDB client", err)
		return
	}
	influxApplication = &InfluxDbApp{
		Client:      client,
		Database:    "udcp_influx",
		InputFormat: INPUT_FORMAT_PIPE_DELIMITED,
	}
	if err := udcp.ListenAndServe(*addr, influxApplication); err != nil {
		log.Fatalf("Failed to start Udcpudcp. Error %s", err)
	}
}

// InfluxDbApp provides an application that sends metrics to influxdb
type InfluxDbApp struct {
	udcp.UdcpApplication

	Client      influxdb.Client
	Database    string
	InputFormat uint8
}

func (app *InfluxDbApp) onDataWriteToInflux(data []byte) error {
	// Create a new point batch
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  app.Database,
		Precision: "s",
	})
	var tags map[string]string
	// TODO: Parse the tags and fields
	if app.InputFormat == INPUT_FORMAT_PIPE_DELIMITED {
		// e.g. temp:38.29|lat:-35.012|lng:12.12|tag_meter_no:ABCD|ts:145923830|tag_host:2981
		// for each string starting with tag use as tag,
		// use everything else as a field
	}
	tags = map[string]string{
		"app": "udcp:influxdb",
	}

	fields := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}

	pt, err := influxdb.NewPoint(
		app.Database,
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		log.Fatal("Failed to create point, won't add to batch", err)
	}
	bp.AddPoint(pt)
	if err := app.Client.Write(bp); err != nil {
		log.Fatal(err)
	}
	return err
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
func (app *InfluxDbApp) OnError(request udcp.UdcpRequest, session udcp.Session) (udcp.UdcpResponse, error) {
	fmt.Printf("Received ErrorPdu, %s", request.Data())
	return udcp.NewProtocolErrorResponse(), nil
}

// OnData returns the request/response handler for the Echo Application
func (app *InfluxDbApp) OnData(request udcp.UdcpRequest, session udcp.Session) (udcp.UdcpResponse, error) {
	if request.HasMoreToSend() {
		// Waiting for Client to send more data
		return udcp.NewReceiveReadyResponse(), nil
	}
	// Handle the decoding of the data here
	//
	// This is the point at which you may send data to an external service
	// since at this point all the data the client intended to send is complete
	data, err := session.RecvBuffer().Read()
	if err != nil {
		return udcp.NewProtocolErrorResponse(), nil
	}
	if err = app.onDataWriteToInflux(data); err != nil {
		return udcp.NewProtocolErrorResponse(), nil
	}
	// We're ready to receive more data
	return udcp.NewReceiveReadyResponse(), nil
}

// OnReceiveReady returns data when a Client is waiting for server data
func (app *InfluxDbApp) OnReceiveReady(request udcp.UdcpRequest, session udcp.Session) (udcp.UdcpResponse, error) {
	return udcp.NewReceiveReadyResponse(), nil
}

// OnReleaseDialogue returns the request/response handler for the Echo Application
func (app *InfluxDbApp) OnReleaseDialogue(request udcp.UdcpRequest, session udcp.Session) (udcp.UdcpResponse, error) {
	return udcp.NewUserAbortReleaseDialogueResponse(), nil
}
