package mqtt

import (
	"fmt"
	"net"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	ussdproxy "github.com/nndi-oss/ussdproxy/lib"
)

// MQTTApplication provides an application that sends metrics to mqtt
type MQTTApplication struct {
	ussdproxy.UdcpApplication

	Client   MQTT.Client
	Addr     string
	Topic    string
	Username string
	Password string
	Database string
	session  ussdproxy.Session
}

func NewMQTTApplication(addr, user, password, topic string) *MQTTApplication {
	conn1, err := net.Dial("tcp", addr) // test the conneciotn
	if err != nil {
		panic(err)
	}
	defer conn1.Close()

	return &MQTTApplication{
		Username: user,
		Password: password,
		Addr:     addr,
		Topic:    topic,
	}
}

func (app *MQTTApplication) GetOrCreateSession() ussdproxy.Session {
	return app.session
}

func (app *MQTTApplication) UseSession(session ussdproxy.Session) {
	app.session = session
}

func (app *MQTTApplication) onDataWriteToMQTT(data []byte) error {
	opts := MQTT.NewClientOptions().AddBroker(app.Addr)
	opts.SetClientID(app.Name())

	// Noop handler for messages
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
	})

	if app.Client == nil {
		app.Client = MQTT.NewClient(opts)
	} else {
		if !app.Client.IsConnected() {
			app.Client = MQTT.NewClient(opts)
		}
	}

	if token := app.Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to the server, got %v", token.Error())
	}

	if token := app.Client.Publish(app.Topic, 0, false, string(data)); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to the server, got %v", token.Error())
	}

	return nil
}

// Name the name of the application
func (app *MQTTApplication) Name() string {
	return "MQTTDB Forwarder"
}

// ApplicationID the unique identifier for the application
func (app *MQTTApplication) ApplicationID() string {
	return "mqtt"
}

// Author the author of the application
func (app *MQTTApplication) Author() string {
	return "NNDI"
}

// Register the MQTTApplication with the server
func (app *MQTTApplication) Register() {
	// noop
}

// OnError returns the request/response handler for the Echo Application
func (app *MQTTApplication) OnError(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	fmt.Printf("Received ErrorPdu, %s", request.Data())
	return ussdproxy.NewProtocolErrorResponse(), nil
}

// OnData returns the request/response handler for the Echo Application
func (app *MQTTApplication) OnData(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
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
	if err = app.onDataWriteToMQTT(data); err != nil {
		return ussdproxy.NewProtocolErrorResponse(), nil
	}
	// We're ready to receive more data
	return ussdproxy.NewReceiveReadyResponse(), nil
}

// OnReceiveReady returns data when a Client is waiting for server data
func (app *MQTTApplication) OnReceiveReady(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	return ussdproxy.NewReceiveReadyResponse(), nil
}

// OnReleaseDialogue returns the request/response handler for the Echo Application
func (app *MQTTApplication) OnReleaseDialogue(request ussdproxy.UdcpRequest, session ussdproxy.Session) (ussdproxy.UdcpResponse, error) {
	return ussdproxy.NewUserAbortReleaseDialogueResponse(), nil
}
