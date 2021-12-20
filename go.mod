module github.com/influxdata/telegraf

go 1.17

require (
	cloud.google.com/go v0.93.3 // indirect
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/monitoring v0.2.0
	cloud.google.com/go/pubsub v1.17.0
	code.cloudfoundry.org/clock v1.0.0 // indirect
	collectd.org v0.5.0
	github.com/Azure/azure-amqp-common-go/v3 v3.1.0 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.3.13
	github.com/Azure/azure-kusto-go v0.4.0
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v55.0.0+incompatible // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/azure-storage-queue-go v0.0.0-20191125232315-636801874cdd
	github.com/Azure/go-amqp v0.13.12 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.16
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v0.4.1
	github.com/Mellanox/rdmamap v0.0.0-20191106181932-7c3c4763a6ee
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/Microsoft/hcsshim v0.8.21 // indirect
	github.com/Shopify/sarama v1.29.1
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/aerospike/aerospike-client-go v1.27.0
	github.com/alecthomas/participle v0.4.1 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1004
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.1.4
	github.com/antchfx/xmlquery v1.3.6
	github.com/antchfx/xpath v1.1.11
	github.com/apache/arrow/go/arrow v0.0.0-20211006091945-a69884db78f4 // indirect
	github.com/apache/thrift v0.15.0
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/aws/aws-sdk-go-v2 v1.9.2
	github.com/aws/aws-sdk-go-v2/config v1.8.3
	github.com/aws/aws-sdk-go-v2/credentials v1.4.3
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.6.0
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.5.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.6.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.2
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.3.2
	github.com/aws/smithy-go v1.8.0
	github.com/benbjohnson/clock v1.1.0
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20190531143454-82441e232cf6
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/couchbase/go-couchbase v0.1.0
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/devigned/tab v0.1.1 // indirect
	github.com/dimchansky/utfbom v1.1.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.9+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/doclambda/protobufquery v0.0.0-20210317203640-88ffabe06a60
	github.com/dynatrace-oss/dynatrace-metric-utils-go v0.3.0
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/echlebek/timeproxy v1.0.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/fatih/color v1.10.0
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-ping/ping v0.0.0-20210201095549-52eed920f98c
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v32 v32.1.0
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gopcua/opcua v0.2.3
	github.com/gophercloud/gophercloud v0.16.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/gosnmp/gosnmp v1.33.0
	github.com/grid-x/modbus v0.0.0-20210224155242-c4a3d042e99b
	github.com/grid-x/serial v0.0.0-20191104121038-e24bc9bf6f08 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/gwos/tcg/sdk v0.0.0-20211130162655-32ad77586ccf
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/harlow/kinesis-consumer v0.3.6-0.20210911031324-5a873d6e9fec
	github.com/hashicorp/consul/api v1.9.1
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.9.5 // indirect
	github.com/influxdata/go-syslog/v3 v3.0.0
	github.com/influxdata/influxdb-observability/common v0.2.8
	github.com/influxdata/influxdb-observability/influx2otel v0.2.8
	github.com/influxdata/influxdb-observability/otel2influx v0.2.8
	github.com/influxdata/tail v1.0.1-0.20210707231403-b283181d1fa7
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/intel/iaevents v1.0.0
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.5.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200307190119-3430c5407db8 // indirect
	github.com/jackc/pgtype v1.3.0 // indirect
	github.com/jackc/pgx/v4 v4.6.0
	github.com/jaegertracing/jaeger v1.26.0 // indirect
	github.com/james4k/rcon v0.0.0-20120923215419-8fbb8268b60a
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jhump/protoreflect v1.8.3-0.20210616212123-6cc1efa697ca
	github.com/jmespath/go-jmespath v0.4.0
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kardianos/service v1.0.0
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/mdlayher/genetlink v1.0.0 // indirect
	github.com/mdlayher/netlink v1.1.0 // indirect
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/miekg/dns v1.1.43
	github.com/minio/highwayhash v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/ipvs v1.0.1
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.4.1 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/multiplay/go-ts3 v1.0.0
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/jwt/v2 v2.1.0 // indirect
	github.com/nats-io/nats-server/v2 v2.6.5
	github.com/nats-io/nats.go v1.13.1-0.20211018182449-f2416a8b1483
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/newrelic/newrelic-telemetry-sdk-go v0.5.1
	github.com/nsqio/go-nsq v1.0.8
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pion/dtls/v2 v2.0.9
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport v0.12.3 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.31.1
	github.com/prometheus/procfs v0.6.0
	github.com/prometheus/prometheus v1.8.2-0.20210430082741-2a4b8e12bbf2
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/riemann/riemann-go-client v0.5.0
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/safchain/ethtool v0.0.0-20200218184317-f459e2d13664
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/sensu/sensu-go/api/core/v2 v2.12.0
	github.com/shirou/gopsutil/v3 v3.21.10
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114 // indirect
	github.com/showwin/speedtest-go v1.1.4
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.2 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/golib/v3 v3.3.38
	github.com/signalfx/sapm-proto v0.7.2 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/sleepinggenius2/gosmi v0.4.3
	github.com/snowflakedb/gosnowflake v1.6.2
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/testcontainers/testcontainers-go v0.11.1
	github.com/tidwall/gjson v1.10.2
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.6
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/vapourismo/knx-go v0.0.0-20211128234507-8198fa17db36
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852 // indirect
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae // indirect
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.26.0
	github.com/wavefronthq/wavefront-sdk-go v0.9.9
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/xdg/scram v1.0.3
	github.com/xdg/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/yuin/gopher-lua v0.0.0-20200603152657-dc2b0ca8b37e // indirect
	go.etcd.io/etcd/api/v3 v3.5.0 // indirect
	go.mongodb.org/mongo-driver v1.7.3
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/collector/model v0.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.24.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/sdk/metric v0.24.0
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/net v0.0.0-20211005215030-d2e5035098b3
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	golang.zx2c4.com/wireguard v0.0.20200121 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/api v0.54.0
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210827211047-25e5f791fe06
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/djherbis/times.v1 v1.2.0
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/ldap.v3 v3.1.0
	gopkg.in/olivere/elastic.v5 v5.0.70
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	modernc.org/cc/v3 v3.33.5 // indirect
	modernc.org/ccgo/v3 v3.9.4 // indirect
	modernc.org/libc v1.9.5 // indirect
	modernc.org/mathutil v1.2.2 // indirect
	modernc.org/memory v1.0.4 // indirect
	modernc.org/opt v0.1.1 // indirect
	modernc.org/sqlite v1.10.8
	modernc.org/strutil v1.1.0 // indirect
	modernc.org/token v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

require github.com/libp2p/go-reuseport v0.1.0

require (
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.4.0 // indirect
	github.com/awslabs/kinesis-aggregation/go v0.0.0-20210630091500-54e17340d32f // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	go.opentelemetry.io/otel v1.0.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.24.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.0.1 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.0.1 // indirect
	go.opentelemetry.io/proto/otlp v0.9.0 // indirect
)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible

// replaced due to https//github.com/mdlayher/apcupsd/issues/10
replace github.com/mdlayher/apcupsd => github.com/influxdata/apcupsd v0.0.0-20210427145308-694d5caead0e

//proxy.golang.org has versions of golang.zx2c4.com/wireguard with leading v's, whereas the git repo has tags without leading v's: https://git.zx2c4.com/wireguard-go/refs/tags
//So, fetching this module with version v0.0.20200121 (as done by the transitive dependency
//https://github.com/WireGuard/wgctrl-go/blob/e35592f146e40ce8057113d14aafcc3da231fbac/go.mod#L12 ) was not working when using GOPROXY=direct.
//Replacing with the pseudo-version works around this.
replace golang.zx2c4.com/wireguard v0.0.20200121 => golang.zx2c4.com/wireguard v0.0.0-20200121152719-05b03c675090

// replaced due to open PR updating protobuf https://github.com/cisco-ie/nx-telemetry-proto/pull/1
replace github.com/cisco-ie/nx-telemetry-proto => github.com/sbezverk/nx-telemetry-proto v0.0.0-20210629125746-3c19a51b1abc

// replaced due to open PR updating protobuf https://github.com/riemann/riemann-go-client/pull/27
replace github.com/riemann/riemann-go-client => github.com/dstrand1/riemann-go-client v0.5.1-0.20211028194734-b5eb11fb5754
