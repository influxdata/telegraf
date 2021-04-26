module github.com/influxdata/telegraf

go 1.16

require (
	cloud.google.com/go v0.56.0
	cloud.google.com/go/bigquery v1.4.0
	cloud.google.com/go/pubsub v1.2.0
	code.cloudfoundry.org/clock v1.0.0 // indirect
	collectd.org v0.5.0
	github.com/Azure/azure-event-hubs-go/v3 v3.2.0
	github.com/Azure/azure-storage-queue-go v0.0.0-20181215014128-6ed74e755687
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/Azure/go-autorest/autorest/adal v0.9.10
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.6
	github.com/BurntSushi/toml v0.3.1
	github.com/Mellanox/rdmamap v0.0.0-20191106181932-7c3c4763a6ee
	github.com/Microsoft/ApplicationInsights-Go v0.4.2
	github.com/Shopify/sarama v1.27.2
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/aerospike/aerospike-client-go v1.27.0
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1004
	github.com/amir/raidman v0.0.0-20170415203553-1ccc43bfb9c9
	github.com/antchfx/xmlquery v1.3.5
	github.com/antchfx/xpath v1.1.11
	github.com/apache/thrift v0.13.0
	github.com/aristanetworks/glog v0.0.0-20191112221043-67e8567f59f3 // indirect
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/aws/aws-sdk-go v1.34.34
	github.com/aws/aws-sdk-go-v2 v1.1.0
	github.com/aws/aws-sdk-go-v2/config v1.1.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.0
	github.com/aws/smithy-go v1.0.0
	github.com/benbjohnson/clock v1.0.3
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/caio/go-tdigest v3.1.0+incompatible
	github.com/cisco-ie/nx-telemetry-proto v0.0.0-20190531143454-82441e232cf6
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/couchbase/go-couchbase v0.0.0-20180501122049-16db1f1fe037
	github.com/couchbase/gomemcached v0.0.0-20180502221210-0da75df14530 // indirect
	github.com/couchbase/goutils v0.0.0-20180530154633-e865a1461c8a // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/dimchansky/utfbom v1.1.1
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/docker/libnetwork v0.8.0-dev.2.0.20181012153825-d7b61745d166
	github.com/eclipse/paho.mqtt.golang v1.3.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-ping/ping v0.0.0-20210201095549-52eed920f98c
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v2.1.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/golang/geo v0.0.0-20190916061304-5b978397cfec
	github.com/golang/protobuf v1.5.1
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.5.5
	github.com/google/go-github/v32 v32.1.0
	github.com/gopcua/opcua v0.1.13
	github.com/gorilla/mux v1.7.3
	github.com/gosnmp/gosnmp v1.31.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/harlow/kinesis-consumer v0.3.1-0.20181230152818-2f58b136fee0
	github.com/hashicorp/consul/api v1.6.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/influxdata/go-syslog/v2 v2.0.1
	github.com/influxdata/tail v1.0.1-0.20200707181643-03a791b270e4
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/james4k/rcon v0.0.0-20120923215419-8fbb8268b60a
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kardianos/service v1.0.0
	github.com/karrick/godirwalk v1.16.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/lib/pq v1.3.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mdlayher/apcupsd v0.0.0-20200608131503-2bf01da7bf1b
	github.com/miekg/dns v1.1.31
	github.com/multiplay/go-ts3 v1.0.0
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/nats-io/nats-server/v2 v2.1.4
	github.com/nats-io/nats.go v1.10.0
	github.com/newrelic/newrelic-telemetry-sdk-go v0.5.1
	github.com/nsqio/go-nsq v1.0.8
	github.com/openconfig/gnmi v0.0.0-20180912164834-33a1865c3029
	github.com/openzipkin/zipkin-go-opentracing v0.3.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.13.0
	github.com/prometheus/procfs v0.1.3
	github.com/prometheus/prometheus v1.8.2-0.20200911110723-e83ef207b6c2
	github.com/riemann/riemann-go-client v0.5.0
	github.com/safchain/ethtool v0.0.0-20200218184317-f459e2d13664
	github.com/sensu/sensu-go/api/core/v2 v2.6.0
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114 // indirect
	github.com/signalfx/golib/v3 v3.3.0
	github.com/sirupsen/logrus v1.6.0
	github.com/soniah/gosnmp v1.25.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.7.0
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	github.com/tidwall/gjson v1.6.0
	github.com/tinylib/msgp v1.1.5
	github.com/vapourismo/knx-go v0.0.0-20201122213738-75fe09ace330
	github.com/vishvananda/netlink v0.0.0-20171020171820-b2de5d10e38e // indirect
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc // indirect
	github.com/vjeantet/grok v1.0.1
	github.com/vmware/govmomi v0.19.0
	github.com/wavefronthq/wavefront-sdk-go v0.9.7
	github.com/wvanbergen/kafka v0.0.0-20171203153745-e2edea948ddf
	github.com/wvanbergen/kazoo-go v0.0.0-20180202103751-f72d8611297a // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/yuin/gopher-lua v0.0.0-20180630135845-46796da1b0b4 // indirect
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68
	golang.org/x/text v0.3.4
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/api v0.29.0
	google.golang.org/genproto v0.0.0-20200815001618-f69a88009b70
	google.golang.org/grpc v1.33.1
	gopkg.in/djherbis/times.v1 v1.2.0
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5
	gopkg.in/ldap.v3 v3.1.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/olivere/elastic.v5 v5.0.70
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	modernc.org/sqlite v1.7.4
)

// replaced due to https://github.com/satori/go.uuid/issues/73
replace github.com/satori/go.uuid => github.com/gofrs/uuid v3.2.0+incompatible
