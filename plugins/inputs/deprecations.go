package inputs

import "github.com/influxdata/telegraf"

var Deprecations = map[string]telegraf.DeprecationInfo{
	"cassandra": {
		Since:  "1.7",
		Notice: "use 'inputs.jolokia2' with the 'cassandra.conf' example configuration instead",
	},
	"http_listener_v2": {
		Since:  "1.9",
		Notice: "has been renamed to 'influxdb_listener', use 'inputs.influxdb_listener' or 'inputs.influxdb_listener_v2' instead",
	},
	"httpjson": {
		Since:  "1.6",
		Notice: "use 'inputs.http' instead",
	},
	"jolokia": {
		Since:  "1.5",
		Notice: "use 'inputs.jolokia2' instead",
	},
	"kafka_consumer_legacy": {
		Since:  "1.20",
		Notice: "use 'inputs.kafka_consumer' instead, NOTE: 'kafka_consumer' only supports Kafka v0.8+",
	},
	"logparser": {
		Since:  "1.15",
		Notice: "use 'inputs.tail' with 'grok' data format instead",
	},
	"snmp_legacy": {
		Since:  "1.20",
		Notice: "use 'inputs.snmp' instead",
	},
	"tcp_listener": {
		Since:  "1.3",
		Notice: "use 'inputs.socket_listener' instead",
	},
	"udp_listener": {
		Since:  "1.3",
		Notice: "use 'inputs.socket_listener' instead",
	},
}
