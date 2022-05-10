package config

import (
	"fmt"

	"github.com/nndi-oss/ussdproxy/pkg/ussd"
	"github.com/nndi-oss/ussdproxy/pkg/ussd/africastalking"
	"github.com/nndi-oss/ussdproxy/pkg/ussd/flares"
	"github.com/nndi-oss/ussdproxy/pkg/ussd/truroute"
)

// ServerConfig configuraiton
type ServerConfig struct {
	Name           string `mapstructure:"name"`
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	RequestTimeout int    `mapstructure:"request_timeout"`
}

// DatabaseConfig database configuration
type DatabaseConfig struct {
	Name     string `mapstructure:"name"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LoggingConfig is configuration for logging in the server
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"log_file"`
}

// UdcpCommandsConfig configuraiton
type UdcpCommandsConfig struct {
	QuerySessionID         bool `mapstructure:"query_session_id"`
	QueryKeepAlive         bool `mapstructure:"query_keep_alive"`
	QueryReceiveReadyLimit bool `mapstructure:"query_receive_ready_limit"`
	QueryMaxBufferSize     bool `mapstructure:"query_max_buffer_size"`
	ClearBuffer            bool `mapstructure:"clear_buffer"`
	GrowBuffer             bool `mapstructure:"grow_buffer"`
	ShrinkBuffer           bool `mapstructure:"shrink_buffer"`
	CacheSession           bool `mapstructure:"cache_session"`
	CloseSession           bool `mapstructure:"close_session"`
}

// SessionConfig is configuration for session management
type SessionConfig struct {
	Database string
	Path     string
	Name     string
	Username string
	Password string
}

// AppConfig configuration
type AppConfig struct {
	Name string `mapstructure:"name"`
}

// AppConfig configuration
type UssdConfig struct {
	Provider    string `mapstructure:"provider"`
	CallbackURL string `mapstructure:"callback_url"`
}

// UdcpConfig configuration
type UdcpConfig struct {
	RequestTimeout    uint64             `mapstructure:"request_timeout"`
	KeepAlive         bool               `mapstructure:"keep_alive"`          // Whether to wait for data
	ReceiveReadyLimit uint8              `mapstructure:"receive_ready_limit"` // Number of RR pdus to send to the server
	MinBufferSize     uint16             `mapstructure:"min_buffer_size"`     // default: 512 # Minimum size of the buffer on the server and client side
	MaxBufferSize     uint16             `mapstructure:"max_buffer_size"`     // default: 8096 # Maximum size of the buffer on the server and client side
	Commands          UdcpCommandsConfig `mapstructure:"commands"`
	Session           SessionConfig      `mapstructure:"session"`
	Apps              []AppConfig        `mapstructure:"apps"` // Services or Apps are applications running on the UDCP server
}

// UssdProxyConfig main configuration struct
type UssdProxyConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	Ussd    UssdConfig    `mapstructure:"ussd"`
	Udcp    UdcpConfig    `mapstructure:"udcp"`
	Logging LoggingConfig `mapstructure:"logging"`
}

func Validate(cfg UssdProxyConfig) error {
	// TODO: implement me!
	return fmt.Errorf("not implemented")
}

func (c *UssdProxyConfig) GetProvider() ussd.UssdProvider {
	if c.Ussd.Provider == "africastalking" {
		return africastalking.New()
	}

	if c.Ussd.Provider == "truroute" {
		return truroute.New()
	}

	if c.Ussd.Provider == "flares" {
		return flares.New()
	}

	return nil
}
