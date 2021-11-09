package inputs

import "github.com/influxdata/telegraf"

var Deprecations = map[string]telegraf.DeprecationInfo{
	"cassandra":             {"1.7", "", "use 'inputs.jolokia2' with the 'cassandra.conf' example configuration instead"},
	"http_listener_v2":      {"1.9", "", "has been renamed to 'influxdb_listener', use 'inputs.influxdb_listener' or 'inputs.influxdb_listener_v2' instead"},
	"httpjson":              {"1.6", "", "use 'inputs.http' instead"},
	"jolokia":               {"1.5", "", "use 'inputs.jolokia2' instead"},
	"kafka_consumer_legacy": {"1.20", "", "use 'inputs.kafka_consumer' instead, NOTE: 'kafka_consumer' only supports Kafka v0.8+"},
	"logparser":             {"1.15", "", "use 'inputs.tail' with 'grok' data format instead"},
	"snmp_legacy":           {"1.20", "", "use 'inputs.snmp' instead"},
	"tcp_listener":          {"1.3", "", "use 'inputs.socket_listener' instead"},
	"udp_listener":          {"1.3", "", "use 'inputs.socket_listener' instead"},
}
