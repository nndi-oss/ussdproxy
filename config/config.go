package config

// DatabaseConfig database configuration
type DatabaseConfig struct {
	host     string
	port     int8
	name     string
	username string
	password string
}

// SessionConfig is configuration for session management
type SessionConfig struct {
	Database string
	Path     string
	Name     string
	Username string
	Password string
}

// LoggingConfig is configuration for logging in the server
type LoggingConfig struct {
	Enabled bool
	Level   string
	Syslog  bool
}

// Config configuration for the server
type Config struct {
	Host     string
	Port     uint8
	Database *DatabaseConfig
	Session  *SessionConfig
	Logging  *LoggingConfig
	Udcp     struct {
		RequestTimeout    uint64
		KeepAlive         bool     // Whether to wait for data
		ReceiveReadyLimit uint8    // Number of RR pdus to send to the server
		MinBufferSize     uint16   // default: 512 # Minimum size of the buffer on the server and client side
		MaxBufferSize     uint16   // default: 8096 # Maximum size of the buffer on the server and client side
		Apps              []string // Services or Apps are applications running on the UDCP server
	}
}