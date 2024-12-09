package snmp

import (
	"time"

	"github.com/influxdata/telegraf/config"
)

type ClientConfig struct {
	// Timeout to wait for a response.
	Timeout              config.Duration `toml:"timeout"`
	Retries              int             `toml:"retries"`
	Version              uint8           `toml:"version"`
	UnconnectedUDPSocket bool            `toml:"unconnected_udp_socket"`

	// Parameters for Version 1 & 2
	Community string `toml:"community"`

	// Parameters for Version 2 & 3
	MaxRepetitions uint32 `toml:"max_repetitions"`

	// Parameters for Version 3
	ContextName  string        `toml:"context_name"`
	SecLevel     string        `toml:"sec_level"`
	SecName      string        `toml:"sec_name"`
	AuthProtocol string        `toml:"auth_protocol"`
	AuthPassword config.Secret `toml:"auth_password"`
	PrivProtocol string        `toml:"priv_protocol"`
	PrivPassword config.Secret `toml:"priv_password"`
	EngineID     string        `toml:"-"`
	EngineBoots  uint32        `toml:"-"`
	EngineTime   uint32        `toml:"-"`

	// Path to mib files
	Path       []string `toml:"path"`
	Translator string   `toml:"-"`
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:        config.Duration(5 * time.Second),
		Retries:        3,
		Version:        2,
		Path:           []string{"/usr/share/snmp/mibs"},
		Translator:     "gosmi",
		Community:      "public",
		MaxRepetitions: 10,
		SecLevel:       "authNoPriv",
		SecName:        "myuser",
		AuthProtocol:   "MD5",
		AuthPassword:   config.NewSecret([]byte("pass")),
	}
}
