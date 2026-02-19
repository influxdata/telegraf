module github.com/influxdata/telegraf

go 1.25.7

godebug x509negativeserial=1

require (
	cloud.google.com/go/auth v0.18.2
	cloud.google.com/go/bigquery v1.73.1
	cloud.google.com/go/monitoring v1.24.3
	cloud.google.com/go/pubsub/v2 v2.4.0
	cloud.google.com/go/storage v1.60.0
	collectd.org v0.6.0
	github.com/99designs/keyring v1.2.2
	github.com/Azure/azure-event-hubs-go/v3 v3.6.2
	github.com/Azure/azure-kusto-go v0.16.1
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2 v2.0.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue v1.0.1
	github.com/Azure/go-autorest/autorest v0.11.30
	github.com/Azure/go-autorest/autorest/adal v0.9.24
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13
	github.com/BurntSushi/toml v1.6.0
	github.com/ClickHouse/clickhouse-go/v2 v2.43.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/IBM/nzgo/v12 v12.0.11
	github.com/IBM/sarama v1.46.3
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/Mellanox/rdmamap v1.1.0
	github.com/PaesslerAG/gval v1.2.4
	github.com/SAP/go-hdb v1.15.0
	github.com/aerospike/aerospike-client-go/v5 v5.11.0
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b
	github.com/alitto/pond v1.9.2
	github.com/alitto/pond/v2 v2.6.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.107
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/jsonquery v1.3.6
	github.com/antchfx/xmlquery v1.5.0
	github.com/antchfx/xpath v1.3.5
	github.com/apache/arrow-go/v18 v18.5.1
	github.com/apache/inlong/inlong-sdk/dataproxy-sdk-twins/dataproxy-sdk-golang v1.0.7
	github.com/apache/iotdb-client-go v1.3.5
	github.com/apache/thrift v0.22.0
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/awnumar/memguard v0.23.0
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.4
	github.com/aws/aws-sdk-go-v2 v1.41.1
	github.com/aws/aws-sdk-go-v2/config v1.32.7
	github.com/aws/aws-sdk-go-v2/credentials v1.19.7
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.17
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.54.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.63.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.55.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.289.1
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.43.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.6
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.35.16
	github.com/aws/smithy-go v1.24.0
	github.com/benbjohnson/clock v1.3.5
	github.com/bluenviron/gomavlib/v3 v3.3.0
	github.com/blues/jsonata-go v1.5.4
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/boschrexroth/ctrlx-datalayer-golang v1.3.1
	github.com/bufbuild/protocompile v0.14.1
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20230117155933-f64c045c77df
	github.com/clarify/clarify-go v0.4.1
	github.com/cloudevents/sdk-go/v2 v2.16.2
	github.com/compose-spec/compose-go v1.20.2
	github.com/coocood/freecache v1.2.5
	github.com/coreos/go-semver v0.3.1
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/couchbase/go-couchbase v0.1.1
	github.com/datadope-io/go-zabbix/v2 v2.0.1
	github.com/digitalocean/go-libvirt v0.0.0-20250417173424-a6a66ef779d6
	github.com/dimchansky/utfbom v1.1.1
	github.com/djherbis/times v1.6.0
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-connections v0.6.0
	github.com/dustin/go-humanize v1.0.1
	github.com/dynatrace-oss/dynatrace-metric-utils-go v0.5.0
	github.com/eclipse/paho.golang v0.23.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/emiago/sipgo v1.2.0
	github.com/facebook/time v0.0.0-20250903103710-a5911c32cdb9
	github.com/fatih/color v1.18.0
	github.com/go-ldap/ldap/v3 v3.4.12
	github.com/go-logfmt/logfmt v0.6.1
	github.com/go-ole/go-ole v1.3.0
	github.com/go-redis/redis/v7 v7.4.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.9.3
	github.com/go-stomp/stomp v2.1.4+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid/v5 v5.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/snappy v1.0.0
	github.com/google/cel-go v0.27.0
	github.com/google/gnxi v0.0.0-20231026134436-d82d9936af15
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v32 v32.1.0
	github.com/google/licensecheck v0.3.1
	github.com/google/uuid v1.6.0
	github.com/gopacket/gopacket v1.5.0
	github.com/gopcua/opcua v0.8.0
	github.com/gophercloud/gophercloud/v2 v2.10.0
	github.com/gorcon/rcon v1.4.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/gosnmp/gosnmp v1.43.2
	github.com/grid-x/modbus v0.0.0-20240503115206-582f2ab60a18
	github.com/gwos/tcg/sdk v0.0.0-20240830123415-f8a34bba6358
	github.com/hashicorp/consul/api v1.33.2
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/hashicorp/vault/api v1.22.0
	github.com/hashicorp/vault/api/auth/approle v0.11.0
	github.com/influxdata/influxdb-observability/common v0.5.12
	github.com/influxdata/influxdb-observability/influx2otel v0.5.12
	github.com/influxdata/influxdb-observability/otel2influx v0.5.12
	github.com/influxdata/line-protocol/v2 v2.2.1
	github.com/influxdata/tail v1.0.1-0.20241014115250-3e0015cb677a
	github.com/influxdata/toml v0.0.0-20251106153700-c381e153d076
	github.com/intel/iaevents v1.1.0
	github.com/intel/powertelemetry v1.0.2
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgio v1.0.0
	github.com/jackc/pgtype v1.14.4
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jedib0t/go-pretty/v6 v6.7.8
	github.com/jeremywohl/flatten/v2 v2.0.0-20211013061545-07e4a09fb8e4
	github.com/jmespath/go-jmespath v0.4.0
	github.com/karrick/godirwalk v1.16.2
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/compress v1.18.4
	github.com/klauspost/pgzip v1.2.6
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b
	github.com/leodido/go-syslog/v4 v4.3.0
	github.com/likexian/whois v1.15.7
	github.com/likexian/whois-parser v1.24.21
	github.com/linkedin/goavro/v2 v2.15.0
	github.com/logzio/azure-monitor-metrics-receiver v1.1.0
	github.com/lxc/incus/v6 v6.21.0
	github.com/mdlayher/apcupsd v0.0.0-20220319200143-473c7b5f3c6a
	github.com/mdlayher/vsock v1.2.1
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/microsoft/go-mssqldb v1.9.6
	github.com/miekg/dns v1.1.72
	github.com/moby/ipvs v1.1.0
	github.com/multiplay/go-ts3 v1.2.0
	github.com/nats-io/nats-server/v2 v2.12.4
	github.com/nats-io/nats.go v1.48.0
	github.com/netsampler/goflow2/v2 v2.2.6
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/nsqio/go-nsq v1.1.0
	github.com/nwaples/tacplus v0.0.3
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/openconfig/gnmi v0.14.1
	github.com/openconfig/goyang v1.6.3
	github.com/opensearch-project/opensearch-go/v2 v2.3.0
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.5.0
	github.com/openzipkin/zipkin-go v0.4.3
	github.com/p4lang/p4runtime v1.5.0
	github.com/pavlo-v-chernykh/keystore-go/v4 v4.5.0
	github.com/pborman/ansi v1.0.0
	github.com/pcolladosoto/goslurm v0.1.0
	github.com/peterbourgon/unixtransport v0.0.7
	github.com/pion/dtls/v3 v3.1.2
	github.com/prometheus-community/pro-bing v0.8.0
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.67.5
	github.com/prometheus/procfs v0.19.2
	github.com/prometheus/prometheus v0.308.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/rclone/rclone v1.69.3
	github.com/redis/go-redis/v9 v9.18.0
	github.com/riemann/riemann-go-client v0.5.1-0.20211206220514-f58f10cdce16
	github.com/robbiet480/go.nut v0.0.0-20220219091450-bd8f121e1fa1
	github.com/robinson/gos7 v0.0.0-20240315073918-1f14519e4846
	github.com/safchain/ethtool v0.7.0
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/seancfoley/ipaddress-go v1.7.1
	github.com/sensu/sensu-go/api/core/v2 v2.16.0
	github.com/shirou/gopsutil/v4 v4.26.1
	github.com/showwin/speedtest-go v1.7.10
	github.com/signalfx/golib/v3 v3.3.54
	github.com/sijms/go-ora/v2 v2.9.0
	github.com/sirupsen/logrus v1.9.4
	github.com/sleepinggenius2/gosmi v0.4.4
	github.com/snowflakedb/gosnowflake v1.19.0
	github.com/srebhan/cborquery v1.0.4
	github.com/srebhan/protobufquery v1.0.4
	github.com/stretchr/testify v1.11.1
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/tdrn-org/go-hue v0.3.0
	github.com/tdrn-org/go-nsdp v0.5.0
	github.com/tdrn-org/go-tr064 v0.2.3
	github.com/testcontainers/testcontainers-go v0.40.0
	github.com/testcontainers/testcontainers-go/modules/azure v0.40.0
	github.com/testcontainers/testcontainers-go/modules/kafka v0.40.0
	github.com/testcontainers/testcontainers-go/modules/vault v0.40.0
	github.com/thomasklein94/packer-plugin-libvirt v0.5.0
	github.com/tidwall/gjson v1.18.0
	github.com/tidwall/wal v1.2.1
	github.com/tinylib/msgp v1.6.3
	github.com/urfave/cli/v2 v2.27.7
	github.com/vapourismo/knx-go v0.0.0-20240915133544-a6ab43471c11
	github.com/vertica/vertica-sql-go v1.3.5
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.52.0
	github.com/wavefronthq/wavefront-sdk-go v0.15.0
	github.com/x448/float16 v0.8.4
	github.com/xdg/scram v1.0.5
	github.com/yuin/goldmark v1.7.16
	go.mongodb.org/mongo-driver v1.17.9
	go.opentelemetry.io/collector/pdata v1.46.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0
	go.opentelemetry.io/otel/sdk/metric v1.39.0
	go.opentelemetry.io/proto/otlp v1.9.0
	go.opentelemetry.io/proto/otlp/collector/profiles/v1development v0.2.0
	go.opentelemetry.io/proto/otlp/profiles/v1development v0.2.0
	go.starlark.net v0.0.0-20260102030733-3fee463870c9
	go.step.sm/crypto v0.76.0
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/crypto v0.48.0
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96
	golang.org/x/mod v0.33.0
	golang.org/x/net v0.50.0
	golang.org/x/oauth2 v0.35.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.41.0
	golang.org/x/term v0.40.0
	golang.org/x/text v0.34.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20211230205640-daad0b7ba671
	gonum.org/v1/gonum v0.17.0
	google.golang.org/api v0.266.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/olivere/elastic.v5 v5.0.86
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	k8s.io/api v0.35.1
	k8s.io/apimachinery v0.35.1
	k8s.io/client-go v0.35.1
	layeh.com/radius v0.0.0-20221205141417-e7fbddd11d68
	modernc.org/sqlite v1.45.0
	software.sslmate.com/src/go-pkcs12 v0.7.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	code.cloudfoundry.org/clock v1.2.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.1 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.1 // indirect
	github.com/Azure/azure-storage-queue-go v0.0.0-20230531184854-c06a8eff66fe // indirect
	github.com/Azure/go-amqp v1.4.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/ClickHouse/ch-go v0.71.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.55.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.55.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Max-Sum/base32768 v0.0.0-20230304063302-18e6ce5945fd // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/abbot/go-http-auth v0.4.0 // indirect
	github.com/alecthomas/participle v0.4.1 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/antithesishq/antithesis-sdk-go v0.5.0-default-no-op // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/apex/log v1.9.0 // indirect
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/awnumar/memcall v0.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.71.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.13 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/brutella/dnssd v1.2.14 // indirect
	github.com/bwmarrin/snowflake v0.3.0 // indirect
	github.com/caio/go-tdigest/v4 v4.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.1 // indirect
	github.com/cncf/xds/go v0.0.0-20251210132809-ee656c7534f5 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/couchbase/gomemcached v0.1.3 // indirect
	github.com/couchbase/goutils v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/creack/goselect v0.1.3 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.7.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/echlebek/timeproxy v1.0.0 // indirect
	github.com/elastic/go-sysinfo v1.8.1 // indirect
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.36.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8-0.20250403174932-29230038a667 // indirect
	github.com/go-chi/chi/v5 v5.2.4 // indirect
	github.com/go-darwin/apfs v0.0.0-20211011131704-f84b94dbf348 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-git/go-billy/v5 v5.6.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.1-0.20211022031912-bfb69110f8dd // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.11 // indirect
	github.com/googleapis/gax-go/v2 v2.17.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/grid-x/serial v0.0.0-20211107191517-583c7356b3aa // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/hashicorp/packer-plugin-sdk v0.3.2 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/icholy/digest v1.1.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/jaegertracing/jaeger v1.47.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmhodges/clock v1.2.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jzelinskie/whirlpool v0.0.0-20201016144138-0675e54bb004 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/likexian/gokit v0.25.16 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-ieproxy v0.0.11 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mdlayher/genetlink v1.2.0 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/minio/highwayhash v1.0.4-0.20251030100505-070ab1a87a76 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/jwt/v2 v2.8.0 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/ncw/swift/v2 v2.0.3 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.139.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.3.0 // indirect
	github.com/opencontainers/umoci v0.6.1-0.20251213054154-70fc5ee1f4df // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/panjf2000/ants/v2 v2.11.3 // indirect
	github.com/panjf2000/gnet/v2 v2.9.7 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v4 v4.0.1 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rfjakob/eme v1.1.2 // indirect
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rootless-containers/proto/go-proto v0.0.0-20260109132551-5f4e706f2d5d // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/seancfoley/bintree v1.3.1 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/sapm-proto v0.12.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/twmb/murmur3 v1.1.7 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/urfave/cli v1.22.17 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vbatts/go-mtree v0.7.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/assert v1.3.1 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	github.com/zentures/cityhash v0.0.0-20131128155616-cdd6a94144ab // indirect
	github.com/zitadel/logging v0.7.0 // indirect
	github.com/zitadel/oidc/v3 v3.45.3 // indirect
	github.com/zitadel/schema v1.3.2 // indirect
	go.bug.st/serial v1.6.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.135.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.46.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.140.0 // indirect
	go.opentelemetry.io/collector/semconv v0.128.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.39.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/telemetry v0.0.0-20260109210033-bd525da824e2 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	golang.zx2c4.com/wireguard v0.0.0-20211209221555-9c9e7e272434 // indirect
	google.golang.org/genproto v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.2.2 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250910181357-589584f1c912 // indirect
	k8s.io/utils v0.0.0-20260108192941-914a6e750570 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
