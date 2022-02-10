package ussdproxy

// UssdRequestInterface struct represents a request coming in from the Network
type UssdRequestInterface interface {
	PhoneNumber() string
	Data() []byte
	RawText() string
	SessionID() string
	Channel() string
	Provider() string
}
