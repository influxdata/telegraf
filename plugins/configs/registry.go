package configs

import "github.com/influxdata/telegraf/config"

func Add(name string, creator config.ConfigCreator) {
	config.ConfigPlugins[name] = creator
}
