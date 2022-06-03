package snmp

import (
	"github.com/influxdata/telegraf/config"
)

type ClientConfig struct {
	// Timeout to wait for a response.
	Timeout config.Duration `toml:"timeout"`
	Retries int             `toml:"retries"`
	// Values: 1, 2, 3
	Version uint8 `toml:"version"`
	// Path to mib files
	Path []string `toml:"path"`
	// Translator implementation
	Translator string `toml:"-"`

	// Parameters for Version 1 & 2
	Community string `toml:"community"`

	// Parameters for Version 2 & 3
	MaxRepetitions uint32 `toml:"max_repetitions"`

	// Parameters for Version 3
	ContextName string `toml:"context_name"`
	// Values: "noAuthNoPriv", "authNoPriv", "authPriv"
	SecLevel string `toml:"sec_level"`
	SecName  string `toml:"sec_name"`
	// Values: "MD5", "SHA", "". Default: ""
	AuthProtocol string `toml:"auth_protocol"`
	AuthPassword string `toml:"auth_password"`
	// Values: "DES", "AES", "". Default: ""
	PrivProtocol string `toml:"priv_protocol"`
	PrivPassword string `toml:"priv_password"`
	EngineID     string `toml:"-"`
	EngineBoots  uint32 `toml:"-"`
	EngineTime   uint32 `toml:"-"`
}
