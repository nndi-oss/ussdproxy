package core

import "errors"

var (
	// errors
	InvalidHeaderError      = errors.New("Failed to parse UdcpHeader")
	VersionError            = errors.New("Version Not Supported. Only version 0x00 is supported")
	MoreToSendWithDataError = errors.New("Request with MoreToSendFlag was sent with data")
	LengthNotValidError     = errors.New("Length of data inconsistent with Len value in header")
	TooMuchDataError        = errors.New("The request contained too much data") // this could be a good thing, in another lifetime?
	FailedToSaveSession     = errors.New("Failed to save session in Session store")
	InvalidPhoneNumberError = errors.New("PhoneNumber found in the request was invalid")
	UnknownParseError       = errors.New("Failed to parse UssdRequest")
)
