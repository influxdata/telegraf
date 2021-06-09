module github.com/golang-migrate/migrate/v4

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/spanner v1.18.0 // indirect
	cloud.google.com/go/storage v1.10.0 // indirect
	collectd.org v0.5.0
	github.com/Azure/azure-event-hubs-go/v3 v3.2.0
	github.com/Azure/azure-storage-queue-go v0.0.0-20181215014128-6ed74e755687
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/Azure/go-autorest/autorest/adal v0.9.10
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.6
	github.com/BurntSushi/toml v0.3.1
	github.com/ClickHouse/clickhouse-go v1.4.3 // indirect
	github.com/Mellanox/rdmamap v0.0.0-20191106181932-7c3c4763a6ee
	github.com/Shopify/sarama v1.27.2
	github.com/aerospike/aerospike-client-go v1.27.0
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1004
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/xmlquery v1.3.5
	github.com/antchfx/xpath v1.1.11
	github.com/apache/arrow/go/arrow v0.0.0-20210521153258-78c88a9f517b // indirect
	github.com/apache/thrift v0.13.0
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/aws/aws-sdk-go v1.34.34
	github.com/aws/aws-sdk-go-v2 v1.6.0
	github.com/aws/aws-sdk-go-v2/config v1.3.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/smithy-go v1.4.0
	github.com/benbjohnson/clock v1.0.3
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cenkalti/backoff/v4 v4.0.2 // indirect
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20190531143454-82441e232cf6
	github.com/cockroachdb/cockroach-go v0.0.0-20190925194419-606b3d062051 // indirect
	github.com/couchbase/go-couchbase v0.0.0-20180501122049-16db1f1fe037
	github.com/cznic/mathutil v0.0.0-20180504122225-ca4c9f2c1369 // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/dhui/dktest v0.3.4 // indirect
	github.com/dimchansky/utfbom v1.1.1
	github.com/docker/docker v17.12.0-ce-rc1.0.20210128214336-420b1d36250f+incompatible
	github.com/eclipse/paho.mqtt.golang v1.3.0
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsouza/fake-gcs-server v1.17.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-ping/ping v0.0.0-20210201095549-52eed920f98c
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/here v0.6.0 // indirect
	github.com/goburrow/modbus v0.1.0
	github.com/gobwas/glob v0.2.3
	github.com/gocql/gocql v0.0.0-20190301043612-f6df8288f9b4 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.3
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.5
	github.com/google/go-github/v32 v32.1.0
	github.com/google/go-github/v35 v35.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gopcua/opcua v0.1.13
	github.com/gorilla/mux v1.7.4
	github.com/gosnmp/gosnmp v1.32.0
	github.com/harlow/kinesis-consumer v0.3.1-0.20181230152818-2f58b136fee0
	github.com/hashicorp/consul/api v1.8.1
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/influxdata/go-syslog/v2 v2.0.1
	github.com/influxdata/influxdb-observability/common v0.0.0-20210429174543-86ae73cafd31
	github.com/influxdata/influxdb-observability/otel2influx v0.0.0-20210429174543-86ae73cafd31
	github.com/influxdata/influxdb-observability/otlp v0.0.0-20210429174543-86ae73cafd31
	github.com/influxdata/tail v1.0.1-0.20200707181643-03a791b270e4
	github.com/influxdata/telegraf v1.18.3
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/jackc/pgconn v1.8.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451 // indirect
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/jackc/pgx/v4 v4.10.1 // indirect
	github.com/james4k/rcon v0.0.0-20120923215419-8fbb8268b60a
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kardianos/service v1.0.0
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/compress v1.12.2 // indirect
	github.com/ktrysmt/go-bitbucket v0.6.4 // indirect
	github.com/lib/pq v1.8.0 // indirect
	github.com/markbates/pkger v0.15.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/microsoft/ApplicationInsights-Go v0.4.4
	github.com/miekg/dns v1.1.31
	github.com/moby/ipvs v1.0.1
	github.com/multiplay/go-ts3 v1.0.0
	github.com/mutecomm/go-sqlcipher/v4 v4.4.0 // indirect
	github.com/nakagami/firebirdsql v0.0.0-20190310045651-3c02a58cfed8 // indirect
	github.com/nats-io/nats-server/v2 v2.1.4
	github.com/nats-io/nats.go v1.10.0
	github.com/neo4j/neo4j-go-driver v1.8.1-0.20200803113522-b626aa943eba // indirect
	github.com/newrelic/newrelic-telemetry-sdk-go v0.5.1
	github.com/nsqio/go-nsq v1.0.8
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/openzipkin/zipkin-go-opentracing v0.3.4
	github.com/pierrec/lz4/v4 v4.1.7 // indirect
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	github.com/prometheus/procfs v0.1.3
	github.com/prometheus/prometheus v1.8.2-0.20200911110723-e83ef207b6c2
	github.com/riemann/riemann-go-client v0.5.0
	github.com/safchain/ethtool v0.0.0-20200218184317-f459e2d13664
	github.com/sensu/sensu-go/api/core/v2 v2.6.0
	github.com/shirou/gopsutil v3.21.3+incompatible
	github.com/signalfx/golib/v3 v3.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/sleepinggenius2/gosmi v0.4.3
	github.com/snowflakedb/gosnowflake v1.4.3 // indirect
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.7.0
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/tidwall/gjson v1.6.0
	github.com/tinylib/msgp v1.1.5
	github.com/vapourismo/knx-go v0.0.0-20201122213738-75fe09ace330
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.19.0
	github.com/wavefronthq/wavefront-sdk-go v0.9.7
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/xanzy/go-gitlab v0.15.0 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/xdg/stringprep v1.0.0 // indirect
	gitlab.com/nyarla/go-crypt v0.0.0-20160106005555-d9a5dc2b789b // indirect
	go.mongodb.org/mongo-driver v1.3.2 // indirect
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023
	golang.org/x/oauth2 v0.0.0-20210413134643-5e61552d6c78
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210521090106-6ca3eb03dfc2
	golang.org/x/text v0.3.6
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.1.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/api v0.45.0
	google.golang.org/genproto v0.0.0-20210427215850-f767ed18ee4d
	google.golang.org/grpc v1.37.0
	gopkg.in/djherbis/times.v1 v1.2.0
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/ldap.v3 v3.1.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/olivere/elastic.v5 v5.0.70
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	modernc.org/b v1.0.0 // indirect
	modernc.org/db v1.0.0 // indirect
	modernc.org/file v1.0.0 // indirect
	modernc.org/fileutil v1.0.0 // indirect
	modernc.org/golex v1.0.0 // indirect
	modernc.org/internal v1.0.0 // indirect
	modernc.org/lldb v1.0.0 // indirect
	modernc.org/ql v1.0.0 // indirect
	modernc.org/sortutil v1.1.0 // indirect
	modernc.org/sqlite v1.10.6
	modernc.org/zappy v1.0.0 // indirect
)

go 1.16
