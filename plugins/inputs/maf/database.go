package maf

import (
	"database/sql"
	"sync"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	// go-mssqldb initialization
	//_ "github.com/zensqlmonitor/go-mssqldb"
	_ "github.com/denisenkom/go-mssqldb"
)

type Database struct {
	Servers []string `toml: "servers"`
}

type