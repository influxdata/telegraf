module github.com/influxdata/telegraf

go 1.16

require (
	cloud.google.com/go v0.88.0
	cloud.google.com/go/bigquery v1.19.0
	cloud.google.com/go/pubsub v1.13.0
	code.cloudfoundry.org/clock v1.0.0 // indirect
	collectd.org v0.5.0
	github.com/Azure/azure-event-hubs-go/v3 v3.3.11
	github.com/Azure/azure-kusto-go v0.3.2
	github.com/Azure/azure-storage-queue-go v0.0.0-20191125232315-636801874cdd
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/Azure/go-autorest/autorest/adal v0.9.10
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.6
	github.com/BurntSushi/toml v0.3.1
	github.com/Mellanox/rdmamap v1.0.0
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Shopify/sarama v1.29.1
	github.com/StackExchange/wmi v1.2.0 // indirect
	github.com/aerospike/aerospike-client-go v1.27.0
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1194
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.1.4
	github.com/antchfx/xmlquery v1.3.6
	github.com/antchfx/xpath v1.2.0
	github.com/apache/thrift v0.13.0
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/aws/aws-sdk-go v1.38.69
	github.com/aws/aws-sdk-go-v2 v1.3.2
	github.com/aws/aws-sdk-go-v2/config v1.1.5
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.6
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/smithy-go v1.3.1
	github.com/benbjohnson/clock v1.0.3
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20190531143454-82441e232cf6
	github.com/couchbase/go-couchbase v0.1.0
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/dimchansky/utfbom v1.1.1
	github.com/docker/docker v20.10.7+incompatible
	github.com/doclambda/protobufquery v0.0.0-20210317203640-88ffabe06a60
	github.com/dynatrace-oss/dynatrace-metric-utils-go v0.2.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-ping/ping v0.0.0-20210506233800-ff8be3320020
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v32 v32.1.0
	github.com/gopcua/opcua v0.1.13
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/gosnmp/gosnmp v1.32.0
	github.com/grid-x/modbus v0.0.0-20210714071042-7af2b65ec03b
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/harlow/kinesis-consumer v0.3.1-0.20181230152818-2f58b136fee0
	github.com/hashicorp/consul/api v1.9.1
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/influxdata/go-syslog/v3 v3.0.0
	github.com/influxdata/influxdb-observability/common v0.2.4
	github.com/influxdata/influxdb-observability/influx2otel v0.2.4
	github.com/influxdata/influxdb-observability/otel2influx v0.2.4
	github.com/influxdata/tail v1.0.1-0.20210707231403-b283181d1fa7
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/jackc/pgx/v4 v4.13.0
	github.com/james4k/rcon v0.0.0-20210222224819-34a67ca2b2d6
	github.com/jhump/protoreflect v1.9.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kardianos/service v1.2.0
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/miekg/dns v1.1.43
	github.com/moby/ipvs v1.0.1
	github.com/multiplay/go-ts3 v1.0.0
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/nats-server/v2 v2.3.2
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.0
	github.com/nsqio/go-nsq v1.0.8
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/openzipkin/zipkin-go-opentracing v0.3.4
	github.com/pion/dtls/v2 v2.0.9
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/procfs v0.7.1
	github.com/prometheus/prometheus v1.8.2-0.20200911110723-e83ef207b6c2
	github.com/riemann/riemann-go-client v0.5.0
	github.com/safchain/ethtool v0.0.0-20201023143004-874930cb3ce0
	github.com/sensu/sensu-go/api/core/v2 v2.9.0
	github.com/shirou/gopsutil v3.21.6+incompatible
	github.com/signalfx/golib/v3 v3.3.36
	github.com/sirupsen/logrus v1.8.1
	github.com/sleepinggenius2/gosmi v0.4.3
	github.com/snowflakedb/gosnowflake v1.6.0
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/tbrandon/mbserver v0.0.0-20210320091329-a1f8ae952881
	github.com/testcontainers/testcontainers-go v0.11.1
	github.com/tidwall/gjson v1.8.1
	github.com/tinylib/msgp v1.1.6
	github.com/tklauser/go-sysconf v0.3.7 // indirect
	github.com/vapourismo/knx-go v0.0.0-20201122213738-75fe09ace330
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852 // indirect
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.26.0
	github.com/wavefronthq/wavefront-sdk-go v0.9.9
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/xdg/scram v1.0.3
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	go.mongodb.org/mongo-driver v1.5.3
	go.opentelemetry.io/collector/model v0.0.0-20210723184018-3b7d6ce4830c
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/text v0.3.6
	golang.org/x/tools v0.1.5
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/api v0.51.0
	google.golang.org/genproto v0.0.0-20210726200206-e7812ac95cc0
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/djherbis/times.v1 v1.3.0
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/ldap.v3 v3.1.0
	gopkg.in/olivere/elastic.v5 v5.0.86
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	modernc.org/sqlite v1.12.0
)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible

// replaced due to https//github.com/mdlayher/apcupsd/issues/10
replace github.com/mdlayher/apcupsd => github.com/influxdata/apcupsd v0.0.0-20210427145308-694d5caead0e
