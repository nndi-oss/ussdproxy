package ussdproxy

import "errors"

var (
	// errors
	ErrInvalidHeader       = errors.New("Failed to parse UdcpHeader")
	ErrVersion             = errors.New("Version Not Supported. Only version 0x00 is supported")
	ErrMoreToSendWithData  = errors.New("Request with MoreToSendFlag was sent with data")
	ErrLengthNotValid      = errors.New("Length of data inconsistent with Len value in header")
	ErrTooMuchData         = errors.New("The request contained too much data") // this could be a good thing, in another lifetime?
	ErrFailedToSaveSession = errors.New("Failed to save session in Session store")
	ErrInvalidPhoneNumber  = errors.New("PhoneNumber found in the request was invalid")
	ErrUnknownParse        = errors.New("Failed to parse UssdRequest")
)
