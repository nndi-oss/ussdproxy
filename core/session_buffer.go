package mudcp

/// Buffer implementation for the protocol-level buffering on the
/// server side for UDCP sessions
import (
	"bytes"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/valyala/bytebufferpool"
)

// SessionBuffer is a read/write buffer
type SessionBuffer interface {
	Read() ([]byte, error)
	ReadAt(p []byte, offset int64) (int, error)
	Write(data []byte) error
	Set([]byte) error
	FillWith(SessionBuffer) error
	Purge()
	IsEmpty() bool
	Length() int
}

type sessionBuffer struct {
	bucketKeyName []byte
	dbPath        string
	id            *string
	offset        int32 // Offset is the position in the buffer we're at when reading/writing from the buffer
	// TODO: how to handle concurrent access to boltdb?
	db   *bolt.DB
	data *bytebufferpool.ByteBuffer
}

func getSessionStore(dbPath string) *bolt.DB {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// GetOrCreateSession obtains or creates the session from the sessionStore (bolt or redis)
func GetOrCreateSession(sessionID string) Session {
	bucketKeyName := []byte(sessionID)
	var existingData []byte
	db := getSessionStore(sessionID)
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket(bucketKeyName)
		if b == nil {
			// bucket doesn't exist yet
			return nil
		}
		b = tx.Bucket(bucketKeyName)
		existingData = b.Get([]byte(sessionID))
		return nil
	})
	if existingData == nil {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketKeyName)
			var err error
			if b == nil {
				b, err = tx.CreateBucket(bucketKeyName)
				if err != nil {
					log.Fatalf("Failed to create bucket. %s", err)
					return err
				}
			}
			b.Put([]byte(sessionID), make([]byte, 0)) // empty session data
			return nil
		})
		return &session{
			sessionID:  sessionID,
			recvBuffer: NewEmptySessionBuffer(&sessionID),
			sendBuffer: NewEmptySessionBuffer(&sessionID),
		}
	}
	return &session{
		sessionID:  sessionID,
		recvBuffer: NewSessionBuffer(&sessionID, existingData),
		sendBuffer: NewSessionBuffer(&sessionID, existingData),
	}
}

func (s *session) SessionID() string {
	return s.sessionID
}
func (s *session) RecvBuffer() SessionBuffer {
	return s.recvBuffer
}
func (s *session) SendBuffer() SessionBuffer {
	return s.sendBuffer
}

func (s *session) IsOpen() bool {
	return s.isCommitted == false
}

func (s *session) Reset() {
	s.recvBuffer.Purge()
	s.sendBuffer.Purge()
	s.isCommitted = false
}

func (s *session) Close() {
	s.isCommitted = true
}

func (s *session) Commit() {
	s.isCommitted = true
}

func NewEmptySessionBuffer(sessionID *string) *sessionBuffer {
	return &sessionBuffer{
		id:     sessionID,
		offset: 0,
		data:   bytebufferpool.Get(),
	}
}

func NewSessionBuffer(sessionID *string, data []byte) *sessionBuffer {
	s := &sessionBuffer{
		bucketKeyName: []byte("udcpSessions"),
		dbPath:        "udcp.sessions",
		id:            sessionID,
		offset:        0,
		data:          bytebufferpool.Get(),
	}
	s.data.Write(data)
	return s
}

func (sb *sessionBuffer) Read() ([]byte, error) {
	return sb.data.Bytes(), nil
}

func (sb *sessionBuffer) Write(data []byte) error {
	db = getSessionStore()
	defer db.Close()
	_, err := sb.data.Write(data)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyName)
		b.Put([]byte(*sb.id), sb.data.Bytes())
		return nil
	})
	return err
}

func (sb *sessionBuffer) FillWith(buf SessionBuffer) error {
	data, err := buf.Read()
	if err != nil {
		return err
	}
	sb.data.Set(data)
	return nil
}

func (sb *sessionBuffer) ReadAt(p []byte, offset int64) (int, error) {
	r := bytes.NewReader(sb.data.Bytes())
	return r.ReadAt(p, offset)
}

func (sb *sessionBuffer) Set(data []byte) error {
	db = getSessionStore()
	defer db.Close()
	bytebufferpool.Put(sb.data)
	sb.data = bytebufferpool.Get()
	sb.data.Write(data)
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyName)
		b.Put([]byte(*sb.id), sb.data.Bytes())
		return nil
	})
	return err
}

func (sb *sessionBuffer) Purge() {
	db = getSessionStore()
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyName)
		b.Put([]byte(fmt.Sprintf("prev_%s", *sb.id)),
			sb.data.Bytes())
		b.Put([]byte(*sb.id), nil) // empty the session data
		return nil
	})
	bytebufferpool.Put(sb.data)
}

func (sb *sessionBuffer) IsEmpty() bool {
	return sb.data.Len() < 1
}

func (sb *sessionBuffer) Length() int {
	return sb.data.Len()
}
