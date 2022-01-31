package core

// Session provides read and write buffers for a specific Session
type Session interface {
	SessionID() string
	RecvBuffer() SessionBuffer
	SendBuffer() SessionBuffer
	IsOpen() bool
	Close()
	Reset()
	Commit()
}

type session struct {
	Session
	sessionID   string
	isCommitted bool
	recvBuffer  *sessionBuffer
	sendBuffer  *sessionBuffer
}
