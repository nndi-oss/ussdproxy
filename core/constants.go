package core

import (
	"time"
)

// pduType is the Protocol Data Unit type for messages of the UDCP protocol
type pduType uint8

const (
	DusstVersion             = 0x1
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

	ApplicationPduType    pduType = 0x41
	CommandPduType        pduType = 0x43
	CommandPduWithMtsType pduType = 0x63
	DataLongPduType       pduType = 0x44
	DataPduWithMtsType    pduType = 0x64
	ReceiveReadyPduType   pduType = 0x52
	ErrorPduType          pduType = 0x45
	ReleaseDialogPduType  pduType = 0x58
	QueryPduType          pduType = 0x51
	QueryPduWithMtsType   pduType = 0x71
	UdcpProtocolPduType   pduType = 0x55

	ApplicationPduAscii    = "A;"
	CommandPduAscii        = "C;"
	CommandPduWithMtsAscii = "c;"
	DataPduAscii           = "D;"
	DataPduWithMtsAscii    = "d;"
	ReceiveReadyPduAscii   = "R;"
	ErrorPduAscii          = "E;"
	QueryPduAscii          = "Q;"
	QueryPduWithMtsAscii   = "q;"
	UdcpProtocolPduAscii   = "U;"
	ReleaseDialogPduAscii  = "X;"

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

func (p pduType) HasMoreToSend() bool {
	if p == CommandPduWithMtsType ||
		p == DataLongPduType ||
		p == DataPduWithMtsType ||
		p == QueryPduWithMtsType {
		return true
	}
	return false
}

// String returns the string representation of the PDU Type
func (p pduType) String() string {
	switch p {
	case ApplicationPduType:
		return ApplicationPduAscii
	case CommandPduType:
		return CommandPduAscii
	case CommandPduWithMtsType:
		return CommandPduWithMtsAscii
	case DataLongPduType:
		return DataPduAscii
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
