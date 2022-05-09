package config

type UssdProxyConfig struct {
	ServerConfig
	UdcpConfig
	DatabaseConfig
	LoggingConfig
}

type ServerConfig struct {
	Host string `viper:"host"`
	Port int    `viper:"port"`
}

// DatabaseConfig database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	Username string
	Password string
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

type UdcpConfig struct {
	RequestTimeout    uint64
	KeepAlive         bool     // Whether to wait for data
	ReceiveReadyLimit uint8    // Number of RR pdus to send to the server
	MinBufferSize     uint16   // default: 512 # Minimum size of the buffer on the server and client side
	MaxBufferSize     uint16   // default: 8096 # Maximum size of the buffer on the server and client side
	Apps              []string // Services or Apps are applications running on the UDCP server
}
