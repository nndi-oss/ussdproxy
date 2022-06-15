package ussdproxy

import (
	"strings"
	"time"
)

// PduType is the Protocol Data Unit type for messages of the UDCP protocol
type PduType uint8
type PduTypeAscii string

const (
	ProtocolVersion          = 0x1
	UssdProcessResponseTimer = 1 * time.Minute
	UssdRequestTimer         = 1 * time.Minute
	// USSD strings can have a length of 160 octets but can also be
	// configured differently in different networks.
	// So we go for 140 characters (old twitter style)
	MaxUssdLength        = 140
	MaxDataLength        = 127
	MaxSmsLength         = 160
	MaxReceiveReadyCount = 5
	UssdContinue         = "CON"
	UssdEnd              = "END"

	EchoApplication = "echo"

	ApplicationPduType    PduType = 0x41
	CommandPduType        PduType = 0x43
	CommandPduWithMtsType PduType = 0x63
	DataLongPduType       PduType = 0x44
	DataPduWithMtsType    PduType = 0x64
	ReceiveReadyPduType   PduType = 0x52
	ErrorPduType          PduType = 0x45
	ErrorNotAsciiPduType  PduType = 0x46
	ReleaseDialogPduType  PduType = 0x58
	QueryPduType          PduType = 0x51
	QueryPduWithMtsType   PduType = 0x71
	UdcpProtocolPduType   PduType = 0x55
	InvalidPduType        PduType = 0x99

	ApplicationPduAscii    PduTypeAscii = "A;"
	CommandPduAscii        PduTypeAscii = "C;"
	CommandPduWithMtsAscii PduTypeAscii = "CX;"
	DataLongPduAscii       PduTypeAscii = "D;"
	DataPduWithMtsAscii    PduTypeAscii = "DX;"
	ReceiveReadyPduAscii   PduTypeAscii = "R;"
	ErrorPduAscii          PduTypeAscii = "E;"
	QueryPduAscii          PduTypeAscii = "Q;"
	QueryPduWithMtsAscii   PduTypeAscii = "QX;"
	UdcpProtocolPduAscii   PduTypeAscii = "U;"
	ReleaseDialogPduAscii  PduTypeAscii = "X;"

	ErrorCodeUnknownMask    = 0x66
	ErrorCodeProtoErrorMask = 0x67
	ErrorCodeVersionMask    = 0x68
	ErrorCodeExtAddrMask    = 0x69

	ReleaseCodeUnknownMask     = 0x77
	ReleaseCodeUssdTimeoutMask = 0x76
	ReleaseCodeIdleDialogMask  = 0x75
	ReleaseCodeUserAbortMask   = 0x74

	NoDataResponse    = "__NODATA__"
	NoDataResponseLen = len("__NODATA__")
)

func (p PduType) HasMoreToSend() bool {
	if p == CommandPduWithMtsType ||
		p == DataLongPduType ||
		p == DataPduWithMtsType ||
		p == QueryPduWithMtsType {
		return true
	}
	return false
}

func (p PduTypeAscii) HasMoreToSend() bool {
	if p == CommandPduWithMtsAscii ||
		p == DataLongPduAscii ||
		p == DataPduWithMtsAscii ||
		p == QueryPduWithMtsAscii {
		return true
	}
	return false
}

func RequestPduType(data string) PduType {
	if strings.HasPrefix(data, "A;") {
		return ApplicationPduType
	} else if strings.HasPrefix(data, "C;") {
		return CommandPduType
	} else if strings.HasPrefix(data, "CX;") {
		return CommandPduWithMtsType
	} else if strings.HasPrefix(data, "D;") {
		return DataLongPduType
	} else if strings.HasPrefix(data, "DX;") {
		return DataPduWithMtsType
	} else if strings.HasPrefix(data, "R;") {
		return ReceiveReadyPduType
	} else if strings.HasPrefix(data, "E;") {
		return ErrorPduType
	} else if strings.HasPrefix(data, "Q;") {
		return QueryPduType
	} else if strings.HasPrefix(data, "QX;") {
		return QueryPduWithMtsType
	} else if strings.HasPrefix(data, "U;") {
		return UdcpProtocolPduType
	} else if strings.HasPrefix(data, "X;") {
		return ReleaseDialogPduType
	}

	return InvalidPduType
}

func StripPdu(data string, pduType PduType) string {
	return strings.Replace(data, pduType.String(), "", 1)
}

// String returns the string representation of the PDU Type
func (p PduType) Ascii() PduTypeAscii {
	switch p {
	case ApplicationPduType:
		return ApplicationPduAscii
	case CommandPduType:
		return CommandPduAscii
	case CommandPduWithMtsType:
		return CommandPduWithMtsAscii
	case DataLongPduType:
		return DataLongPduAscii
	case DataPduWithMtsType:
		return DataPduWithMtsAscii
	case ReceiveReadyPduType:
		return ReceiveReadyPduAscii
	case ReleaseDialogPduType:
		return ReleaseDialogPduAscii
	case QueryPduType:
		return QueryPduAscii
	case QueryPduWithMtsType:
		return QueryPduWithMtsAscii
	case UdcpProtocolPduType:
		return UdcpProtocolPduAscii
	case ErrorPduType:
	default:
		return ErrorPduAscii
	}
	return ErrorPduAscii
}

// String returns the string representation of the PDU Type
func (p PduType) String() string {
	return string(p.Ascii())
}
