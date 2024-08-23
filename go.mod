module github.com/influxdata/telegraf

go 1.20

require (
	cloud.google.com/go/bigquery v1.55.0
	cloud.google.com/go/monitoring v1.16.0
	cloud.google.com/go/pubsub v1.33.0
	collectd.org v0.4.0
	github.com/Azure/azure-event-hubs-go/v3 v3.3.17
	github.com/Azure/azure-kusto-go v0.5.2
	github.com/Azure/azure-storage-queue-go v0.0.0-20191125232315-636801874cdd
	github.com/Azure/go-autorest/autorest v0.11.28
	github.com/Azure/go-autorest/autorest/adal v0.9.22
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11
	github.com/BurntSushi/toml v1.2.1
	github.com/ClickHouse/clickhouse-go v1.5.4
	github.com/Mellanox/rdmamap v0.0.0-20191106181932-7c3c4763a6ee
	github.com/Shopify/sarama v1.32.0
	github.com/aerospike/aerospike-client-go/v5 v5.7.0
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1483
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.1.5
	github.com/antchfx/xmlquery v1.3.9
	github.com/antchfx/xpath v1.2.0
	github.com/apache/thrift v0.16.0
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/aws/aws-sdk-go-v2 v1.17.7
	github.com/aws/aws-sdk-go-v2/config v1.18.19
	github.com/aws/aws-sdk-go-v2/credentials v1.13.18
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.8.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.13.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.14.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.13.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.7
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.13.6
	github.com/aws/smithy-go v1.13.5
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
	github.com/docker/docker v25.0.6+incompatible
	github.com/doclambda/protobufquery v0.0.0-20210317203640-88ffabe06a60
	github.com/dynatrace-oss/dynatrace-metric-utils-go v0.3.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/fatih/color v1.15.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ldap/ldap/v3 v3.4.1
	github.com/go-logfmt/logfmt v0.6.0
	github.com/go-ping/ping v0.0.0-20210201095549-52eed920f98c
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.6.0
	github.com/google/go-github/v32 v32.1.0
	github.com/gopcua/opcua v0.3.1
	github.com/gophercloud/gophercloud v1.7.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/gosnmp/gosnmp v1.34.0
	github.com/grid-x/modbus v0.0.0-20211113184042-7f2251c342c9
	github.com/gwos/tcg/sdk v0.0.0-20211223101342-35fbd1ae683c
	github.com/harlow/kinesis-consumer v0.3.6-0.20211204214318-c2b9f79d7ab6
	github.com/hashicorp/consul/api v1.25.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/influxdata/go-syslog/v3 v3.0.0
	github.com/influxdata/influxdb-observability/common v0.2.10
	github.com/influxdata/influxdb-observability/influx2otel v0.2.10
	github.com/influxdata/influxdb-observability/otel2influx v0.2.10
	github.com/influxdata/line-protocol/v2 v2.2.1
	github.com/influxdata/tail v1.0.1-0.20210707231403-b283181d1fa7
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/intel/iaevents v1.0.0
	github.com/jackc/pgx/v4 v4.18.3
	github.com/james4k/rcon v0.0.0-20120923215419-8fbb8268b60a
	github.com/jhump/protoreflect v1.8.3-0.20210616212123-6cc1efa697ca
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kardianos/service v1.2.1
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/matttproud/golang_protobuf_extensions v1.0.4
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/miekg/dns v1.1.56
	github.com/moby/ipvs v1.0.1
	github.com/multiplay/go-ts3 v1.0.1
	github.com/nats-io/nats-server/v2 v2.10.18
	github.com/nats-io/nats.go v1.36.0
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/nsqio/go-nsq v1.1.0
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/pion/dtls/v2 v2.2.6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/client_model v0.5.0
	github.com/prometheus/common v0.44.0
	github.com/prometheus/procfs v0.11.1
	github.com/prometheus/prometheus v0.48.0
	github.com/riemann/riemann-go-client v0.5.0
	github.com/safchain/ethtool v0.0.0-20200218184317-f459e2d13664
	github.com/sensu/sensu-go/api/core/v2 v2.13.0
	github.com/shirou/gopsutil/v3 v3.23.9
	github.com/showwin/speedtest-go v1.1.4
	github.com/signalfx/golib/v3 v3.3.53
	github.com/sirupsen/logrus v1.9.3
	github.com/sleepinggenius2/gosmi v0.4.4
	github.com/snowflakedb/gosnowflake v1.6.21
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.8.4
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/testcontainers/testcontainers-go v0.26.0
	github.com/tidwall/gjson v1.10.2
	github.com/tinylib/msgp v1.1.6
	github.com/vapourismo/knx-go v0.0.0-20211128234507-8198fa17db36
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.27.3
	github.com/wavefronthq/wavefront-sdk-go v0.9.10
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/xdg/scram v1.0.3
	go.mongodb.org/mongo-driver v1.15.0
	go.opentelemetry.io/collector/model v0.44.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0
	go.opentelemetry.io/otel/sdk/metric v1.21.0
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	golang.org/x/net v0.27.0
	golang.org/x/oauth2 v0.13.0
	golang.org/x/sync v0.7.0
	golang.org/x/sys v0.22.0
	golang.org/x/text v0.16.0
	gonum.org/v1/gonum v0.11.0
	google.golang.org/api v0.147.0
	google.golang.org/genproto v0.0.0-20231002182017-d307bd883b97
	google.golang.org/genproto/googleapis/api v0.0.0-20231012201019-e917dd12ba7a
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/olivere/elastic.v5 v5.0.86
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.28.2
	k8s.io/apimachinery v0.28.2
	k8s.io/client-go v0.28.2
	modernc.org/sqlite v1.18.2
)

require (
	cloud.google.com/go v0.110.8 // indirect
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.2 // indirect
	code.cloudfoundry.org/clock v1.0.0 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/Azure/azure-amqp-common-go/v3 v3.2.3 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v65.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.0.0 // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/go-amqp v0.17.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/alecthomas/participle v0.4.1 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/arrow/go/v12 v12.0.0 // indirect
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.59 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.31 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.31.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.6 // indirect
	github.com/awslabs/kinesis-aggregation/go v0.0.0-20210630091500-54e17340d32f // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/golz4 v0.0.0-20150217214814-ef862a3cdc58 // indirect
	github.com/containerd/containerd v1.7.11 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/echlebek/timeproxy v1.0.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.0.0-20170517235910-f1bb20e5a188 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/flatbuffers v23.1.21+incompatible // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.1 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/grid-x/serial v0.0.0-20211107191517-583c7356b3aa // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jaegertracing/jaeger v1.38.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/jwt/v2 v2.5.8 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.0.2 // indirect
	github.com/pion/udp/v2 v2.0.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/sapm-proto v0.12.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twmb/murmur3 v1.1.7 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/vishvananda/netlink v1.2.1-beta.2 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.etcd.io/etcd/api/v3 v3.5.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/term v0.22.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231009173412-8bfb1ae86b6c // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.1.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible

// replaced due to https//github.com/mdlayher/apcupsd/issues/10
replace github.com/mdlayher/apcupsd => github.com/influxdata/apcupsd v0.0.0-20210427145308-694d5caead0e

// replaced due to open PR updating protobuf https://github.com/cisco-ie/nx-telemetry-proto/pull/1
replace github.com/cisco-ie/nx-telemetry-proto => github.com/sbezverk/nx-telemetry-proto v0.0.0-20210629125746-3c19a51b1abc

// replaced due to open PR updating protobuf https://github.com/riemann/riemann-go-client/pull/27
replace github.com/riemann/riemann-go-client => github.com/dstrand1/riemann-go-client v0.5.1-0.20211028194734-b5eb11fb5754

// replace gopsutil with local copy with LVM related change for temorory fix
replace github.com/shirou/gopsutil/v3 => ./patches/gopsutil/v3
