package easedbautl

import (
	"github.com/influxdata/telegraf"
	"os"
)

var (
	//SchemaThroughput  = "mysql-throughput"
	//SchemaConnection  = "mysql-connection"
	//SchemaInnodb      = "mysql-innodb"
	//SchemaDbSize      = "mysql-dbsize"
	//SchemaReplication = "mysql-replication"
	//SchemaSnapshot    = "mysql-snapshot"
	SchemaThroughput  = "throughput"
	SchemaConnection  = "connection"
	SchemaInnodb      = "innodb"
	SchemaDbSize      = "dbsize"
	SchemaReplication = "replication"
	SchemaSnapshot    = "snapshot"

	SchemaCpu    = "cpu"
	SchemaMem    = "mem"
	SchemaDisk   = "disk"
	SchemaDiskIO = "diskio"
	SchemaNet    = "net"
)

// Add global tags.
//  The input parameter, measurement, will be add as a tag too, then the output plugin elasticsearch has chance to embedded
//  The measurement name into the index name
//  If the input map, tags, is not nil, new tags will be appended, otherwise a new tags map created.
func AddGlobalTags(measurement string, metric *telegraf.Metric) error {
	category := "platform";
	switch measurement {
	case SchemaCpu, SchemaMem, SchemaDisk, SchemaDiskIO, SchemaNet:
		category = "infrastructure"
	case SchemaThroughput, SchemaConnection, SchemaInnodb,
			SchemaDbSize, SchemaReplication, SchemaSnapshot:
		category = "platform"
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	(*metric).AddTag("category", category)
	(*metric).AddTag("hostname", hostname)
	(*metric).AddTag("measurement", measurement)
	// todo : add other global tags

	return nil
}
