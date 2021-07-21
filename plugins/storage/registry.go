package storage

import "github.com/influxdata/telegraf/config"

func Add(name string, creator config.StorageCreator) {
	config.StoragePlugins[name] = creator
}
