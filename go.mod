module github.com/influxdata/telegraf

go 1.18

require (
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/monitoring v1.2.0
	cloud.google.com/go/pubsub v1.18.0
	collectd.org v0.5.0
	github.com/Azure/azure-event-hubs-go/v3 v3.3.17
	github.com/Azure/azure-kusto-go v0.6.0
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
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1529
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.1.5
	github.com/antchfx/xmlquery v1.3.9
	github.com/antchfx/xpath v1.2.0
	github.com/apache/thrift v0.15.0
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/aws/aws-sdk-go-v2 v1.16.2
	github.com/aws/aws-sdk-go-v2/config v1.15.3
	github.com/aws/aws-sdk-go-v2/credentials v1.11.2
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.15.4
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.15.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.13.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.3
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.3.2
	github.com/aws/smithy-go v1.11.2
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
	github.com/docker/docker v20.10.14+incompatible
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
	github.com/google/gnxi v0.0.0-20220411075422-cd6b043b7fd0
	github.com/google/go-cmp v0.5.7
	github.com/google/go-github/v32 v32.1.0
	github.com/gopcua/opcua v0.3.3
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
	github.com/influxdata/influxdb-observability/common v0.2.18
	github.com/influxdata/influxdb-observability/influx2otel v0.2.18
	github.com/influxdata/influxdb-observability/otel2influx v0.2.18
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
	github.com/miekg/dns v1.1.48
	github.com/moby/ipvs v1.0.1
	github.com/multiplay/go-ts3 v1.0.1
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/nsqio/go-nsq v1.1.0
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/openconfig/gnmi v0.0.0-20200617225440-d2b4e6a45802
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
	github.com/xdg/scram v1.0.5
	github.com/yuin/goldmark v1.4.1
	go.mongodb.org/mongo-driver v1.9.0
	go.opentelemetry.io/collector/pdata v0.49.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.28.0
	go.opentelemetry.io/otel/metric v0.28.0
	go.opentelemetry.io/otel/sdk/metric v0.28.0
	go.starlark.net v0.0.0-20220328144851-d1966c6b9fcd
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27
	golang.org/x/text v0.3.7
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20211230205640-daad0b7ba671
	gonum.org/v1/gonum v0.11.0
	google.golang.org/api v0.67.0
	google.golang.org/genproto v0.0.0-20220207164111-0872dc986b00
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
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
	cloud.google.com/go v0.100.2 // indirect
	cloud.google.com/go/compute v0.1.0 // indirect
	cloud.google.com/go/iam v0.1.1 // indirect
	code.cloudfoundry.org/clock v1.0.0 // indirect
	github.com/Azure/azure-amqp-common-go/v3 v3.2.3 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v61.2.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.21.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v0.8.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.3.0 // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/go-amqp v0.17.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/Microsoft/hcsshim v0.8.24 // indirect
	github.com/alecthomas/participle v0.4.1 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20211006091945-a69884db78f4 // indirect
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.3 // indirect
	github.com/awslabs/kinesis-aggregation/go v0.0.0-20210630091500-54e17340d32f // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudflare/golz4 v0.0.0-20150217214814-ef862a3cdc58 // indirect
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.5.11 // indirect
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/echlebek/timeproxy v1.0.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.0.0-20170517235910-f1bb20e5a188 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/grid-x/serial v0.0.0-20211107191517-583c7356b3aa // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.9.6 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.11.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.10.0 // indirect
	github.com/jaegertracing/jaeger v1.26.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v0.0.0-20200817173448-b6b71def0850 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.14.4 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mdlayher/genetlink v1.1.0 // indirect
	github.com/mdlayher/netlink v1.4.2 // indirect
	github.com/mdlayher/socket v0.0.0-20211102153432-57e3fa563ecb // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.5.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220113022732-58e87895b296 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport v0.13.0 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.2 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/sapm-proto v0.7.2 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852 // indirect
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae // indirect
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.0 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/etcd/api/v3 v3.5.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/collector/model v0.49.0 // indirect
	go.opentelemetry.io/otel v1.6.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.6.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.6.0 // indirect
	go.opentelemetry.io/otel/trace v1.6.0 // indirect
	go.opentelemetry.io/proto/otlp v0.12.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/exp v0.0.0-20200513190911-00229845015e // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	golang.org/x/tools v0.1.9 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20211209221555-9c9e7e272434 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	honnef.co/go/tools v0.2.2 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	modernc.org/cc/v3 v3.33.5 // indirect
	modernc.org/ccgo/v3 v3.9.4 // indirect
	modernc.org/libc v1.9.5 // indirect
	modernc.org/mathutil v1.2.2 // indirect
	modernc.org/memory v1.0.4 // indirect
	modernc.org/opt v0.1.1 // indirect
	modernc.org/strutil v1.1.0 // indirect
	modernc.org/token v1.0.0 // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect

)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible

// replaced due to https//github.com/mdlayher/apcupsd/issues/10
replace github.com/mdlayher/apcupsd => github.com/influxdata/apcupsd v0.0.0-20210427145308-694d5caead0e

// replaced due to open PR updating protobuf https://github.com/cisco-ie/nx-telemetry-proto/pull/1
replace github.com/cisco-ie/nx-telemetry-proto => github.com/sbezverk/nx-telemetry-proto v0.0.0-20210629125746-3c19a51b1abc

// replaced due to open PR updating protobuf https://github.com/riemann/riemann-go-client/pull/27
replace github.com/riemann/riemann-go-client => github.com/dstrand1/riemann-go-client v0.5.1-0.20211028194734-b5eb11fb5754
