module github.com/influxdata/telegraf

go 1.18

require (
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/monitoring v1.2.0
	cloud.google.com/go/pubsub v1.18.0
	collectd.org v0.5.0
	github.com/Azure/azure-event-hubs-go/v3 v3.3.17
	github.com/Azure/azure-kusto-go v0.5.2
	github.com/Azure/azure-storage-queue-go v0.0.0-20191125232315-636801874cdd
	github.com/Azure/go-autorest/autorest v0.11.24
	github.com/Azure/go-autorest/autorest/adal v0.9.18
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11
	github.com/BurntSushi/toml v0.4.1
	github.com/ClickHouse/clickhouse-go v1.5.4
	github.com/Mellanox/rdmamap v0.0.0-20191106181932-7c3c4763a6ee
	github.com/Shopify/sarama v1.32.0
	github.com/aerospike/aerospike-client-go/v5 v5.7.0
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1483
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.1.5
	github.com/antchfx/xmlquery v1.3.9
	github.com/antchfx/xpath v1.2.0
	github.com/apache/thrift v0.15.0
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/aws/aws-sdk-go-v2 v1.14.0
	github.com/aws/aws-sdk-go-v2/config v1.13.1
	github.com/aws/aws-sdk-go-v2/credentials v1.8.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.11.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.13.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.14.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.13.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.14.0
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.3.2
	github.com/aws/smithy-go v1.11.0
	github.com/benbjohnson/clock v1.3.0
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20190531143454-82441e232cf6
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f
	github.com/couchbase/go-couchbase v0.1.1
	github.com/denisenkom/go-mssqldb v0.12.0
	github.com/dimchansky/utfbom v1.1.1
	github.com/djherbis/times v1.5.0
	github.com/docker/docker v20.10.11+incompatible
	github.com/doclambda/protobufquery v0.0.0-20210317203640-88ffabe06a60
	github.com/dynatrace-oss/dynatrace-metric-utils-go v0.3.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/fatih/color v1.10.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ldap/ldap/v3 v3.4.1
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-ping/ping v0.0.0-20210201095549-52eed920f98c
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.4.1
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.5.7
	github.com/google/go-github/v32 v32.1.0
	github.com/gopcua/opcua v0.3.1
	github.com/gophercloud/gophercloud v0.24.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/gosnmp/gosnmp v1.34.0
	github.com/grid-x/modbus v0.0.0-20211113184042-7f2251c342c9
	github.com/gwos/tcg/sdk v0.0.0-20211223101342-35fbd1ae683c
	github.com/harlow/kinesis-consumer v0.3.6-0.20210911031324-5a873d6e9fec
	github.com/hashicorp/consul/api v1.12.0
	github.com/hashicorp/go-uuid v1.0.2
	github.com/influxdata/go-syslog/v3 v3.0.0
	github.com/influxdata/influxdb-observability/common v0.2.10
	github.com/influxdata/influxdb-observability/influx2otel v0.2.10
	github.com/influxdata/influxdb-observability/otel2influx v0.2.10
	github.com/influxdata/line-protocol/v2 v2.2.1
	github.com/influxdata/tail v1.0.1-0.20210707231403-b283181d1fa7
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/intel/iaevents v1.0.0
	github.com/jackc/pgx/v4 v4.15.0
	github.com/james4k/rcon v0.0.0-20120923215419-8fbb8268b60a
	github.com/jhump/protoreflect v1.8.3-0.20210616212123-6cc1efa697ca
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kardianos/service v1.2.1
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/miekg/dns v1.1.46
	github.com/moby/ipvs v1.0.1
	github.com/multiplay/go-ts3 v1.0.1
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/nsqio/go-nsq v1.1.0
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/pborman/ansi v1.0.0
	github.com/pion/dtls/v2 v2.0.13
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.32.1
	github.com/prometheus/procfs v0.7.3
	github.com/prometheus/prometheus v1.8.2-0.20210430082741-2a4b8e12bbf2
	github.com/riemann/riemann-go-client v0.5.0
	github.com/safchain/ethtool v0.0.0-20200218184317-f459e2d13664
	github.com/sensu/sensu-go/api/core/v2 v2.13.0
	github.com/shirou/gopsutil/v3 v3.22.3
	github.com/showwin/speedtest-go v1.1.4
	github.com/signalfx/golib/v3 v3.3.43
	github.com/sirupsen/logrus v1.8.1
	github.com/sleepinggenius2/gosmi v0.4.4
	github.com/snowflakedb/gosnowflake v1.6.2
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.7.1
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/testcontainers/testcontainers-go v0.12.0
	github.com/tidwall/gjson v1.10.2
	github.com/tinylib/msgp v1.1.6
	github.com/vapourismo/knx-go v0.0.0-20211128234507-8198fa17db36
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.27.3
	github.com/wavefronthq/wavefront-sdk-go v0.9.10
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/xdg/scram v1.0.3
	github.com/yuin/goldmark v1.4.1
	go.mongodb.org/mongo-driver v1.8.3
	go.opentelemetry.io/collector/model v0.44.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.27.0
	go.opentelemetry.io/otel/metric v0.27.0
	go.opentelemetry.io/otel/sdk/metric v0.27.0
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27
	golang.org/x/text v0.3.7
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20211230205640-daad0b7ba671
	gonum.org/v1/gonum v0.11.0
	google.golang.org/api v0.67.0
	google.golang.org/genproto v0.0.0-20220207164111-0872dc986b00
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/olivere/elastic.v5 v5.0.86
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.3
	modernc.org/sqlite v1.10.8
)

require (
	cloud.google.com/go/iam v0.1.1 // indirect
	code.cloudfoundry.org/clock v1.0.0 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20211006091945-a69884db78f4 // indirect
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.5.3 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/containerd/containerd v1.5.9 // indirect
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/grid-x/serial v0.0.0-20211107191517-583c7356b3aa // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.zx2c4.com/wireguard v0.0.0-20211209221555-9c9e7e272434 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect

)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible

// replaced due to https//github.com/mdlayher/apcupsd/issues/10
replace github.com/mdlayher/apcupsd => github.com/influxdata/apcupsd v0.0.0-20210427145308-694d5caead0e

// replaced due to open PR updating protobuf https://github.com/cisco-ie/nx-telemetry-proto/pull/1
replace github.com/cisco-ie/nx-telemetry-proto => github.com/sbezverk/nx-telemetry-proto v0.0.0-20210629125746-3c19a51b1abc

// replaced due to open PR updating protobuf https://github.com/riemann/riemann-go-client/pull/27
replace github.com/riemann/riemann-go-client => github.com/dstrand1/riemann-go-client v0.5.1-0.20211028194734-b5eb11fb5754
