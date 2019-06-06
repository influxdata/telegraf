package easedbautl

import (
	"github.com/influxdata/telegraf"
	"os"
)

// Add global tags.
//  The input parameter, measurement, will be add as a tag too, then the output plugin elasticsearch has chance to embedded
//  The measurement name into the index name
//  If the input map, tags, is not nil, new tags will be appended, otherwise a new tags map created.
func AddGlobalTags(measurement string, metric *telegraf.Metric) error {
	category := "platform";
	switch measurement {
	case "cpu", "mem", "disk", "diskio", "net":
		category = "infrastructure"
	case "mysql-throughput", "mysql-connections", "mysql-innodb", "mysql-snapshot":
		category = "platform"
	}

	hostname, err := os.Hostname()
	if err != nil {
		return erql
	}

	(*metric).AddTag("category", category)
	(*metric).AddTag("hostname", hostname)
	(*metric).AddTag("measurement", measurement)
	// todo : add other global tags

	return nil
}
