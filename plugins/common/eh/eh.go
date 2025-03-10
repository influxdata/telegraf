package eh

import (
	"github.com/influxdata/telegraf/config"
)


type Config struct {
	ConnectionString string          `toml:"connection_string"`
	Timeout          config.Duration `toml:"timeout"`
	PartitionKey     string          `toml:"partition_key"`
	MaxMessageSize   int             `toml:"max_message_size"`
}