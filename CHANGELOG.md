<a name="unreleased"></a>
## [Unreleased]




<a name="1.22.0"></a>
## [1.22.0] - 2022-02-24

### Bug Fixes

- Accept non-standard OPC UA OK status by implementing a configurable workaround ([#10384](https://github.com/influxdata/telegraf/issues/10384))
- Fix batching logic with write records, introduce concurrent requests ([#8947](https://github.com/influxdata/telegraf/issues/8947))
- Fix panic in logparser due to missing Log. ([#10296](https://github.com/influxdata/telegraf/issues/10296))
- Fix panic in parsers due to missing Log for all plugins using SetParserFunc. ([#10288](https://github.com/influxdata/telegraf/issues/10288))
- Handle duplicate registration of protocol-buffer files gracefully. ([#10188](https://github.com/influxdata/telegraf/issues/10188))
- Implement NaN and inf handling for elasticsearch output ([#10196](https://github.com/influxdata/telegraf/issues/10196))
- Improve parser tests by using go-cmp/cmp ([#10497](https://github.com/influxdata/telegraf/issues/10497))
- Linter fixes for config/config.go ([#10710](https://github.com/influxdata/telegraf/issues/10710))
- Make telegraf compile on Windows with golang 1.16.2 ([#10246](https://github.com/influxdata/telegraf/issues/10246))
- **dedup:** Modifying slice while iterating is dangerous ([#10684](https://github.com/influxdata/telegraf/issues/10684))
- Print loaded plugins and deprecations for once and test ([#10205](https://github.com/influxdata/telegraf/issues/10205))
- Remove verbose logging from disk input plugin ([#10527](https://github.com/influxdata/telegraf/issues/10527))
- Revert deprecation of http_listener_v2 ([#10648](https://github.com/influxdata/telegraf/issues/10648))
- Revert unintented corruption of the Makefile from [#10200](https://github.com/influxdata/telegraf/issues/10200). ([#10203](https://github.com/influxdata/telegraf/issues/10203))
- Set NextCheckTime to LastCheckTime to avoid GroundWork to invent a value ([#10623](https://github.com/influxdata/telegraf/issues/10623))
- Statefull parser handling ([#10575](https://github.com/influxdata/telegraf/issues/10575))
- Sudden close of Telegraf caused by OPC UA input plugin ([#10230](https://github.com/influxdata/telegraf/issues/10230))
- Update go-sensu to v2.12.0 ([#10247](https://github.com/influxdata/telegraf/issues/10247))
- Update modbus readme ([#10501](https://github.com/influxdata/telegraf/issues/10501))
- add RFC3164 to RFC5424 translation to docs ([#10480](https://github.com/influxdata/telegraf/issues/10480))
- add comment to logparser ([#10479](https://github.com/influxdata/telegraf/issues/10479))
- add graylog toml tags ([#10660](https://github.com/influxdata/telegraf/issues/10660))
- **inputs.opcua:** add more data to error log ([#10465](https://github.com/influxdata/telegraf/issues/10465))
- add newline in execd for prometheus parsing ([#10463](https://github.com/influxdata/telegraf/issues/10463))
- address flaky tests in cookie_test.go and graylog_test.go ([#10326](https://github.com/influxdata/telegraf/issues/10326))
- **parsers.json_v2:** allow optional paths and handle wrong paths correctly ([#10468](https://github.com/influxdata/telegraf/issues/10468))
- bump all go.opentelemetry.io dependencies ([#10647](https://github.com/influxdata/telegraf/issues/10647))
- bump cloud.google.com/go/monitoring from 0.2.0 to 1.2.0 ([#10454](https://github.com/influxdata/telegraf/issues/10454))
- bump cloud.google.com/go/pubsub from 1.17.0 to 1.17.1 ([#10504](https://github.com/influxdata/telegraf/issues/10504))
- bump cloud.google.com/go/pubsub from 1.17.1 to 1.18.0 ([#10714](https://github.com/influxdata/telegraf/issues/10714))
- bump github.com/Azure/azure-event-hubs-go/v3 from 3.3.13 to 3.3.17 ([#10449](https://github.com/influxdata/telegraf/issues/10449))
- bump github.com/Azure/azure-kusto-go from 0.5.0 to 0.5.2 ([#10598](https://github.com/influxdata/telegraf/issues/10598))
- bump github.com/ClickHouse/clickhouse-go from 1.5.1 to 1.5.4 ([#10717](https://github.com/influxdata/telegraf/issues/10717))
- bump github.com/aliyun/alibaba-cloud-sdk-go ([#10653](https://github.com/influxdata/telegraf/issues/10653))
- bump github.com/antchfx/jsonquery from 1.1.4 to 1.1.5 ([#10433](https://github.com/influxdata/telegraf/issues/10433))
- bump github.com/antchfx/xmlquery from 1.3.6 to 1.3.9 ([#10507](https://github.com/influxdata/telegraf/issues/10507))
- bump github.com/antchfx/xpath from 1.1.11 to 1.2.0 ([#10436](https://github.com/influxdata/telegraf/issues/10436))
- bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs from 1.5.2 to 1.12.0 ([#10415](https://github.com/influxdata/telegraf/issues/10415))
- bump github.com/aws/aws-sdk-go-v2/service/dynamodb from 1.5.0 to 1.13.0 ([#10692](https://github.com/influxdata/telegraf/issues/10692))
- bump github.com/aws/aws-sdk-go-v2/service/kinesis from 1.6.0 to 1.13.0 ([#10601](https://github.com/influxdata/telegraf/issues/10601))
- bump github.com/aws/aws-sdk-go-v2/service/sts from 1.7.2 to 1.14.0 ([#10602](https://github.com/influxdata/telegraf/issues/10602))
- bump github.com/benbjohnson/clock from 1.1.0 to 1.3.0 ([#10588](https://github.com/influxdata/telegraf/issues/10588))
- bump github.com/couchbase/go-couchbase from 0.1.0 to 0.1.1 ([#10417](https://github.com/influxdata/telegraf/issues/10417))
- bump github.com/denisenkom/go-mssqldb from 0.10.0 to 0.12.0 ([#10503](https://github.com/influxdata/telegraf/issues/10503))
- bump github.com/eclipse/paho.mqtt.golang from 1.3.0 to 1.3.5 ([#9913](https://github.com/influxdata/telegraf/issues/9913))
- bump github.com/google/go-cmp from 0.5.6 to 0.5.7 ([#10563](https://github.com/influxdata/telegraf/issues/10563))
- bump github.com/gopcua/opcua from 0.2.3 to 0.3.1 ([#10626](https://github.com/influxdata/telegraf/issues/10626))
- bump github.com/gophercloud/gophercloud from 0.16.0 to 0.24.0 ([#10693](https://github.com/influxdata/telegraf/issues/10693))
- bump github.com/gosnmp/gosnmp from 1.33.0 to 1.34.0 ([#10450](https://github.com/influxdata/telegraf/issues/10450))
- bump github.com/hashicorp/consul/api from 1.9.1 to 1.12.0 ([#10435](https://github.com/influxdata/telegraf/issues/10435))
- bump github.com/influxdata/influxdb-observability/influx2otel from 0.2.8 to 0.2.10 ([#10432](https://github.com/influxdata/telegraf/issues/10432))
- bump github.com/jackc/pgx/v4 from 4.14.1 to 4.15.0 ([#10702](https://github.com/influxdata/telegraf/issues/10702))
- bump github.com/jackc/pgx/v4 from 4.6.0 to 4.14.1 ([#10453](https://github.com/influxdata/telegraf/issues/10453))
- bump github.com/kardianos/service from 1.0.0 to 1.2.1 ([#10416](https://github.com/influxdata/telegraf/issues/10416))
- bump github.com/multiplay/go-ts3 from 1.0.0 to 1.0.1 ([#10538](https://github.com/influxdata/telegraf/issues/10538))
- bump github.com/nats-io/nats-server/v2 from 2.6.5 to 2.7.2 ([#10638](https://github.com/influxdata/telegraf/issues/10638))
- bump github.com/newrelic/newrelic-telemetry-sdk-go ([#10715](https://github.com/influxdata/telegraf/issues/10715))
- bump github.com/nsqio/go-nsq from 1.0.8 to 1.1.0 ([#10521](https://github.com/influxdata/telegraf/issues/10521))
- bump github.com/pion/dtls/v2 from 2.0.9 to 2.0.13 ([#10418](https://github.com/influxdata/telegraf/issues/10418))
- bump github.com/prometheus/client_golang from 1.11.0 to 1.12.1 ([#10572](https://github.com/influxdata/telegraf/issues/10572))
- bump github.com/prometheus/common from 0.31.1 to 0.32.1 ([#10506](https://github.com/influxdata/telegraf/issues/10506))
- bump github.com/prometheus/procfs from 0.6.0 to 0.7.3 ([#10414](https://github.com/influxdata/telegraf/issues/10414))
- bump github.com/sensu/sensu-go/api/core/v2 from 2.12.0 to 2.13.0 ([#10704](https://github.com/influxdata/telegraf/issues/10704))
- bump github.com/shirou/gopsutil/v3 from 3.21.10 to 3.21.12 ([#10451](https://github.com/influxdata/telegraf/issues/10451))
- bump github.com/signalfx/golib/v3 from 3.3.38 to 3.3.43 ([#10652](https://github.com/influxdata/telegraf/issues/10652))
- bump github.com/vmware/govmomi from 0.26.0 to 0.27.2 ([#10536](https://github.com/influxdata/telegraf/issues/10536))
- bump github.com/vmware/govmomi from 0.27.2 to 0.27.3 ([#10571](https://github.com/influxdata/telegraf/issues/10571))
- bump github.com/wavefronthq/wavefront-sdk-go from 0.9.9 to 0.9.10 ([#10718](https://github.com/influxdata/telegraf/issues/10718))
- bump go.mongodb.org/mongo-driver from 1.7.3 to 1.8.3 ([#10564](https://github.com/influxdata/telegraf/issues/10564))
- bump go.opentelemetry.io/collector/model from 0.39.0 to 0.43.2 ([#10562](https://github.com/influxdata/telegraf/issues/10562))
- bump google.golang.org/api from 0.54.0 to 0.65.0 ([#10434](https://github.com/influxdata/telegraf/issues/10434))
- bump k8s.io/api from 0.23.3 to 0.23.4 ([#10713](https://github.com/influxdata/telegraf/issues/10713))
- bump k8s.io/client-go from 0.22.2 to 0.23.3 ([#10589](https://github.com/influxdata/telegraf/issues/10589))
- check for nil client before closing in amqp ([#10635](https://github.com/influxdata/telegraf/issues/10635))
- check index before assignment ([#10299](https://github.com/influxdata/telegraf/issues/10299))
- collapsed fields by calling more indepth function ([#10430](https://github.com/influxdata/telegraf/issues/10430))
- correctly set ASCII trailer for syslog output ([#10393](https://github.com/influxdata/telegraf/issues/10393))
- cumulative interval start times for stackdriver output ([#10097](https://github.com/influxdata/telegraf/issues/10097))
- do not require networking during tests ([#10321](https://github.com/influxdata/telegraf/issues/10321))
- do not save cache on i386 builds ([#10464](https://github.com/influxdata/telegraf/issues/10464))
- eliminate MIB dependency for ifname processor ([#10214](https://github.com/influxdata/telegraf/issues/10214))
- empty import tzdata for Windows binaries ([#10377](https://github.com/influxdata/telegraf/issues/10377))
- ensure CI tests runs against i386 ([#10457](https://github.com/influxdata/telegraf/issues/10457))
- ensure folders do not get loaded more than once ([#10551](https://github.com/influxdata/telegraf/issues/10551))
- ensure graylog spec fields not prefixed with '_' ([#10209](https://github.com/influxdata/telegraf/issues/10209))
- ensure http body is empty ([#10396](https://github.com/influxdata/telegraf/issues/10396))
- error msg for missing env variables in config ([#10681](https://github.com/influxdata/telegraf/issues/10681))
- fix missing storage in container with disk plugin ([#10318](https://github.com/influxdata/telegraf/issues/10318))
- **http_listener_v2:** fix panic on close ([#10132](https://github.com/influxdata/telegraf/issues/10132))
- flush wavefront output sender on error to clean up broken connections ([#10225](https://github.com/influxdata/telegraf/issues/10225))
- grab table columns more accurately ([#10295](https://github.com/influxdata/telegraf/issues/10295))
- graylog readme to use graylog 3 URLs ([#10481](https://github.com/influxdata/telegraf/issues/10481))
- improve ignore list and list of author names
- include influxdb bucket name in error messages ([#10706](https://github.com/influxdata/telegraf/issues/10706))
- incorrect handling of json_v2 timestamp_path ([#10618](https://github.com/influxdata/telegraf/issues/10618))
- inputs.snmp to respect number of retries configured ([#10268](https://github.com/influxdata/telegraf/issues/10268))
- ipset crash when command not found ([#10474](https://github.com/influxdata/telegraf/issues/10474))
- json_v2 parser timestamp setting ([#10221](https://github.com/influxdata/telegraf/issues/10221))
- license doc outdated causing CI failure ([#10630](https://github.com/influxdata/telegraf/issues/10630))
- linter fixes for "import-shadowing: The name '...' shadows an import name" ([#10689](https://github.com/influxdata/telegraf/issues/10689))
- mac signing issue with arm64 ([#10293](https://github.com/influxdata/telegraf/issues/10293))
- mark TestGatherUDPCert as an integration test ([#10279](https://github.com/influxdata/telegraf/issues/10279))
- mdstat when sync is less than 10% ([#10701](https://github.com/influxdata/telegraf/issues/10701))
- move "Starting Telegraf" log ([#10528](https://github.com/influxdata/telegraf/issues/10528))
- move author thank yous
- mqtt topic extracting no longer requires all three fields ([#10208](https://github.com/influxdata/telegraf/issues/10208))
- **parsers.nagios:** nagios parser now uses real error for logging [#10472](https://github.com/influxdata/telegraf/issues/10472) ([#10473](https://github.com/influxdata/telegraf/issues/10473))
- openweathermap add feels_like field ([#10705](https://github.com/influxdata/telegraf/issues/10705))
- panic due to no module ([#10303](https://github.com/influxdata/telegraf/issues/10303))
- panic is no mibs folder is found ([#10301](https://github.com/influxdata/telegraf/issues/10301))
- parallelism fix for ifname processor ([#10007](https://github.com/influxdata/telegraf/issues/10007))
- pool detection and metrics gathering for ZFS >= 2.1.x ([#10099](https://github.com/influxdata/telegraf/issues/10099))
- prometheusremotewrite wrong timestamp unit ([#10547](https://github.com/influxdata/telegraf/issues/10547))
- re-enable OpenBSD modbus support ([#10385](https://github.com/influxdata/telegraf/issues/10385))
- remove duplicate addition of fields ([#10478](https://github.com/influxdata/telegraf/issues/10478))
- remove signed macOS dotfile artifacts ([#10560](https://github.com/influxdata/telegraf/issues/10560))
- run go mod tidy ([#10273](https://github.com/influxdata/telegraf/issues/10273))
- run gofmt ([#10274](https://github.com/influxdata/telegraf/issues/10274))
- snmp input plugin errors if mibs folder doesn't exist ([#10346](https://github.com/influxdata/telegraf/issues/10346)) ([#10354](https://github.com/influxdata/telegraf/issues/10354))
- snmp marshal error ([#10322](https://github.com/influxdata/telegraf/issues/10322))
- timestamp change during execution of json_v2 parser. ([#10657](https://github.com/influxdata/telegraf/issues/10657))
- typo in docs ([#10441](https://github.com/influxdata/telegraf/issues/10441))
- typo in openstack neutron input plugin (newtron) ([#10284](https://github.com/influxdata/telegraf/issues/10284))
- update GroundWork SDK and improve logging ([#10255](https://github.com/influxdata/telegraf/issues/10255))
- update bug template
- update containerd to 1.5.9 ([#10402](https://github.com/influxdata/telegraf/issues/10402))
- update djherbis/times and fix dependabot ([#10332](https://github.com/influxdata/telegraf/issues/10332))
- update docker memory usage calculation ([#10491](https://github.com/influxdata/telegraf/issues/10491))
- update go-ldap to v3.4.1 ([#10343](https://github.com/influxdata/telegraf/issues/10343))
- update gosmi from v0.4.3 to v0.4.4 ([#10686](https://github.com/influxdata/telegraf/issues/10686))
- use current time as ecs timestamp ([#10636](https://github.com/influxdata/telegraf/issues/10636))
- **json_v2:** use raw values for timestamps ([#10413](https://github.com/influxdata/telegraf/issues/10413))
- use sha256 for RPM digest ([#10272](https://github.com/influxdata/telegraf/issues/10272))
- warning output when running with --test ([#10329](https://github.com/influxdata/telegraf/issues/10329))
- wavefront_disable_prefix_conversion case missing from missingTomlField func ([#10442](https://github.com/influxdata/telegraf/issues/10442))
- windows service - graceful shutdown of telegraf ([#9616](https://github.com/influxdata/telegraf/issues/9616))

### Features

- add compression to Datadog Output ([#9963](https://github.com/influxdata/telegraf/issues/9963))
- Add ClickHouse driver to sql inputs/outputs plugins ([#9671](https://github.com/influxdata/telegraf/issues/9671))
- Add SMART plugin concurrency configuration option, nvme-cli v1.14+ support and lint fixes. ([#10150](https://github.com/influxdata/telegraf/issues/10150))
- Add additional stats to bond collector ([#10137](https://github.com/influxdata/telegraf/issues/10137))
- Add autorestart and restartdelay flags to Windows service ([#10559](https://github.com/influxdata/telegraf/issues/10559))
- Add caching to internet_speed ([#10530](https://github.com/influxdata/telegraf/issues/10530))
- Add noise plugin ([#10057](https://github.com/influxdata/telegraf/issues/10057))
- Add option to disable Wavefront prefix conversion ([#10252](https://github.com/influxdata/telegraf/issues/10252))
- Bump github.com/aerospike/aerospike-client-go from 1.27.0 to 5.7.0 ([#10604](https://github.com/influxdata/telegraf/issues/10604))
- Implemented support for reading raw values, added tests and doc ([#6501](https://github.com/influxdata/telegraf/issues/6501))
- Improve error logging on plugin initialization ([#10307](https://github.com/influxdata/telegraf/issues/10307))
- Modbus add per-request tags ([#10231](https://github.com/influxdata/telegraf/issues/10231))
- Modbus support multiple slaves (gateway feature) ([#9279](https://github.com/influxdata/telegraf/issues/9279))
- Optimize locking for SNMP MIBs loading. ([#10206](https://github.com/influxdata/telegraf/issues/10206))
- Parser plugin restructuring ([#8791](https://github.com/influxdata/telegraf/issues/8791))
- Update underlying KNX library to support new types. ([#10263](https://github.com/influxdata/telegraf/issues/10263))
- Xtremio input ([#9697](https://github.com/influxdata/telegraf/issues/9697))
- add FileVersion and icon to Win exe ([#10487](https://github.com/influxdata/telegraf/issues/10487))
- **mongodb:** add FsTotalSize and FsUsedSize informations ([#10625](https://github.com/influxdata/telegraf/issues/10625))
- add Redis Sentinel input plugin ([#10042](https://github.com/influxdata/telegraf/issues/10042))
- add Vault input plugin ([#10198](https://github.com/influxdata/telegraf/issues/10198))
- add bearer token support to elasticsearch output ([#10399](https://github.com/influxdata/telegraf/issues/10399))
- add builds for riscv64 ([#10262](https://github.com/influxdata/telegraf/issues/10262))
- add consul metrics input plugin ([#10258](https://github.com/influxdata/telegraf/issues/10258))
- add dynamic tagging to gnmi plugin ([#7484](https://github.com/influxdata/telegraf/issues/7484))
- add exclude_root_certs option to x509_cert plugin ([#9822](https://github.com/influxdata/telegraf/issues/9822))
- add heap_size_limit field for input.kibana ([#10243](https://github.com/influxdata/telegraf/issues/10243))
- add mock input plugin ([#9782](https://github.com/influxdata/telegraf/issues/9782))
- add more functionality to template processor ([#10316](https://github.com/influxdata/telegraf/issues/10316))
- add nomad input plugin ([#10106](https://github.com/influxdata/telegraf/issues/10106))
- add option to disable prepared statements for PostgreSQL ([#9710](https://github.com/influxdata/telegraf/issues/9710))
- add option to skip errors during CSV parsing ([#10267](https://github.com/influxdata/telegraf/issues/10267))
- add socks5 proxy support for kafka output plugin ([#8192](https://github.com/influxdata/telegraf/issues/8192))
- add systemd notify support ([#10340](https://github.com/influxdata/telegraf/issues/10340))
- add timeout-setting to Graylog-plugin ([#10220](https://github.com/influxdata/telegraf/issues/10220))
- adds optional list of non retryable http statuscodes to http output plugin ([#10186](https://github.com/influxdata/telegraf/issues/10186))
- aggregator histogram add expiration ([#10520](https://github.com/influxdata/telegraf/issues/10520))
- **inputs.win_perf_counter:** allow errors to be ignored ([#10535](https://github.com/influxdata/telegraf/issues/10535))
- changelog generation
- check TLSConfig early to catch missing certificates ([#10341](https://github.com/influxdata/telegraf/issues/10341))
- collection offset implementation ([#10545](https://github.com/influxdata/telegraf/issues/10545))
- deprecate unused snmp_trap timeout configuration option ([#10339](https://github.com/influxdata/telegraf/issues/10339))
- gather additional stats from memcached ([#10641](https://github.com/influxdata/telegraf/issues/10641))
- ignore bot messages
- process group tag for groundwork output plugin ([#10499](https://github.com/influxdata/telegraf/issues/10499))
- reworked varnish_cache plugin ([#9432](https://github.com/influxdata/telegraf/issues/9432))
- socketstat input plugin ([#3649](https://github.com/influxdata/telegraf/issues/3649))
- socks5 proxy support for websocket ([#10672](https://github.com/influxdata/telegraf/issues/10672))
- support aws managed service for prometheus ([#10202](https://github.com/influxdata/telegraf/issues/10202))
- support darwin arm64 ([#10239](https://github.com/influxdata/telegraf/issues/10239))
- support headers for http plugin with cookie auth ([#10404](https://github.com/influxdata/telegraf/issues/10404))
- update docker client API version ([#10382](https://github.com/influxdata/telegraf/issues/10382))


Thank you for your contributions!



R290 ,
Nirmesh ,
Sven Rebhan ,
Paweł Żak ,
Ted M Lin ,
reimda ,
Thomas Casteleyn ,
Sergey Vilgelm ,
Joshua Powers ,
Sebastian Spaink ,
dependabot[bot] ,
AsafMah ,
Mya ,
Nathan J Mehl ,
Patryk Małek ,
Mikołaj Przybysz ,
Mark Rushakoff ,
Sakerdotes ,
Alan Pope ,
Aaron Wood ,
hulucc ,
Martin Reindl ,
Laurent Sesquès ,
Grimsby ,
Vladislav ,
Jason Heard ,
zachmares ,
JC ,
Jeremy Yang ,
Anatoly Laskaris ,
Kuba Trojan ,
John Seekins ,
Jim Eagle ,
Christian ,
crflanigan ,
Arati Kulkarni ,
Vlasta Hajek ,
cthiel42 ,
Bastien Dronneau ,
Petar Obradović ,
Francesco Bartolini ,
Todd Persen ,
bewing ,
Jarno Huuskonen ,
Irina Vasileva ,
Alexander Krantz ,
Alberto Fernandez ,
Alexander Olekhnovich ,
Josef Johansson ,
Sebastian Thörn ,
Nico Vinzens ,
Eugene Komarov ,
sspaink ,
LINKIWI ,
Pavlo Sumkin ,
Robert Hajek ,
Michael Hoffmann ,
Arthur Gautier ,
Conor Evans ,


<a name="v1.21.4"></a>
## [v1.21.4] - 2022-02-16

### Bug Fixes

- Revert deprecation of http_listener_v2 ([#10648](https://github.com/influxdata/telegraf/issues/10648))
- **parsers.json_v2:** allow optional paths and handle wrong paths correctly ([#10468](https://github.com/influxdata/telegraf/issues/10468))
- bump all go.opentelemetry.io dependencies ([#10647](https://github.com/influxdata/telegraf/issues/10647))
- bump cloud.google.com/go/monitoring from 0.2.0 to 1.2.0 ([#10454](https://github.com/influxdata/telegraf/issues/10454))
- bump github.com/Azure/azure-kusto-go from 0.5.0 to 0.5.2 ([#10598](https://github.com/influxdata/telegraf/issues/10598))
- bump github.com/aliyun/alibaba-cloud-sdk-go ([#10653](https://github.com/influxdata/telegraf/issues/10653))
- bump github.com/aws/aws-sdk-go-v2/service/kinesis from 1.6.0 to 1.13.0 ([#10601](https://github.com/influxdata/telegraf/issues/10601))
- bump github.com/benbjohnson/clock from 1.1.0 to 1.3.0 ([#10588](https://github.com/influxdata/telegraf/issues/10588))
- bump github.com/denisenkom/go-mssqldb from 0.10.0 to 0.12.0 ([#10503](https://github.com/influxdata/telegraf/issues/10503))
- bump github.com/google/go-cmp from 0.5.6 to 0.5.7 ([#10563](https://github.com/influxdata/telegraf/issues/10563))
- bump github.com/gopcua/opcua from 0.2.3 to 0.3.1 ([#10626](https://github.com/influxdata/telegraf/issues/10626))
- bump github.com/multiplay/go-ts3 from 1.0.0 to 1.0.1 ([#10538](https://github.com/influxdata/telegraf/issues/10538))
- bump github.com/nats-io/nats-server/v2 from 2.6.5 to 2.7.2 ([#10638](https://github.com/influxdata/telegraf/issues/10638))
- bump github.com/prometheus/client_golang from 1.11.0 to 1.12.1 ([#10572](https://github.com/influxdata/telegraf/issues/10572))
- bump github.com/signalfx/golib/v3 from 3.3.38 to 3.3.43 ([#10652](https://github.com/influxdata/telegraf/issues/10652))
- bump github.com/vmware/govmomi from 0.26.0 to 0.27.2 ([#10536](https://github.com/influxdata/telegraf/issues/10536))
- bump github.com/vmware/govmomi from 0.27.2 to 0.27.3 ([#10571](https://github.com/influxdata/telegraf/issues/10571))
- bump go.mongodb.org/mongo-driver from 1.7.3 to 1.8.3 ([#10564](https://github.com/influxdata/telegraf/issues/10564))
- bump go.opentelemetry.io/collector/model from 0.39.0 to 0.43.2 ([#10562](https://github.com/influxdata/telegraf/issues/10562))
- bump k8s.io/client-go from 0.22.2 to 0.23.3 ([#10589](https://github.com/influxdata/telegraf/issues/10589))
- check for nil client before closing in amqp ([#10635](https://github.com/influxdata/telegraf/issues/10635))
- ensure folders do not get loaded more than once ([#10551](https://github.com/influxdata/telegraf/issues/10551))
- incorrect handling of json_v2 timestamp_path ([#10618](https://github.com/influxdata/telegraf/issues/10618))
- license doc outdated causing CI failure ([#10630](https://github.com/influxdata/telegraf/issues/10630))
- prometheusremotewrite wrong timestamp unit ([#10547](https://github.com/influxdata/telegraf/issues/10547))
- remove signed macOS dotfile artifacts ([#10560](https://github.com/influxdata/telegraf/issues/10560))
- timestamp change during execution of json_v2 parser. ([#10657](https://github.com/influxdata/telegraf/issues/10657))
- update docker memory usage calculation ([#10491](https://github.com/influxdata/telegraf/issues/10491))
- update go.mod
- use current time as ecs timestamp ([#10636](https://github.com/influxdata/telegraf/issues/10636))


Thank you for your contributions!



reimda ,
Sebastian Spaink ,
dependabot[bot] ,
AsafMah ,
Joshua Powers ,
Mya ,
Mark Rushakoff ,
hulucc ,
Grimsby ,
Jason Heard ,
Josh Powers ,


<a name="v1.21.3"></a>
## [v1.21.3] - 2022-01-27

### Bug Fixes

- Fix batching logic with write records, introduce concurrent requests ([#8947](https://github.com/influxdata/telegraf/issues/8947))
- Make telegraf compile on Windows with golang 1.16.2 ([#10246](https://github.com/influxdata/telegraf/issues/10246))
- Update modbus readme ([#10501](https://github.com/influxdata/telegraf/issues/10501))
- add RFC3164 to RFC5424 translation to docs ([#10480](https://github.com/influxdata/telegraf/issues/10480))
- add comment to logparser ([#10479](https://github.com/influxdata/telegraf/issues/10479))
- add newline in execd for prometheus parsing ([#10463](https://github.com/influxdata/telegraf/issues/10463))
- address flaky tests in cookie_test.go and graylog_test.go ([#10326](https://github.com/influxdata/telegraf/issues/10326))
- bump cloud.google.com/go/pubsub from 1.17.0 to 1.17.1 ([#10504](https://github.com/influxdata/telegraf/issues/10504))
- bump github.com/Azure/azure-event-hubs-go/v3 from 3.3.13 to 3.3.17 ([#10449](https://github.com/influxdata/telegraf/issues/10449))
- bump github.com/antchfx/jsonquery from 1.1.4 to 1.1.5 ([#10433](https://github.com/influxdata/telegraf/issues/10433))
- bump github.com/antchfx/xmlquery from 1.3.6 to 1.3.9 ([#10507](https://github.com/influxdata/telegraf/issues/10507))
- bump github.com/antchfx/xpath from 1.1.11 to 1.2.0 ([#10436](https://github.com/influxdata/telegraf/issues/10436))
- bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs from 1.5.2 to 1.12.0 ([#10415](https://github.com/influxdata/telegraf/issues/10415))
- bump github.com/couchbase/go-couchbase from 0.1.0 to 0.1.1 ([#10417](https://github.com/influxdata/telegraf/issues/10417))
- bump github.com/gosnmp/gosnmp from 1.33.0 to 1.34.0 ([#10450](https://github.com/influxdata/telegraf/issues/10450))
- bump github.com/hashicorp/consul/api from 1.9.1 to 1.12.0 ([#10435](https://github.com/influxdata/telegraf/issues/10435))
- bump github.com/influxdata/influxdb-observability/influx2otel from 0.2.8 to 0.2.10 ([#10432](https://github.com/influxdata/telegraf/issues/10432))
- bump github.com/jackc/pgx/v4 from 4.6.0 to 4.14.1 ([#10453](https://github.com/influxdata/telegraf/issues/10453))
- bump github.com/kardianos/service from 1.0.0 to 1.2.1 ([#10416](https://github.com/influxdata/telegraf/issues/10416))
- bump github.com/nsqio/go-nsq from 1.0.8 to 1.1.0 ([#10521](https://github.com/influxdata/telegraf/issues/10521))
- bump github.com/pion/dtls/v2 from 2.0.9 to 2.0.13 ([#10418](https://github.com/influxdata/telegraf/issues/10418))
- bump github.com/prometheus/common from 0.31.1 to 0.32.1 ([#10506](https://github.com/influxdata/telegraf/issues/10506))
- bump github.com/prometheus/procfs from 0.6.0 to 0.7.3 ([#10414](https://github.com/influxdata/telegraf/issues/10414))
- bump github.com/shirou/gopsutil/v3 from 3.21.10 to 3.21.12 ([#10451](https://github.com/influxdata/telegraf/issues/10451))
- bump google.golang.org/api from 0.54.0 to 0.65.0 ([#10434](https://github.com/influxdata/telegraf/issues/10434))
- collapsed fields by calling more indepth function ([#10430](https://github.com/influxdata/telegraf/issues/10430))
- correctly set ASCII trailer for syslog output ([#10393](https://github.com/influxdata/telegraf/issues/10393))
- cumulative interval start times for stackdriver output ([#10097](https://github.com/influxdata/telegraf/issues/10097))
- do not save cache on i386 builds ([#10464](https://github.com/influxdata/telegraf/issues/10464))
- ensure CI tests runs against i386 ([#10457](https://github.com/influxdata/telegraf/issues/10457))
- ensure http body is empty ([#10396](https://github.com/influxdata/telegraf/issues/10396))
- graylog readme to use graylog 3 URLs ([#10481](https://github.com/influxdata/telegraf/issues/10481))
- ipset crash when command not found ([#10474](https://github.com/influxdata/telegraf/issues/10474))
- **parsers.nagios:** nagios parser now uses real error for logging [#10472](https://github.com/influxdata/telegraf/issues/10472) ([#10473](https://github.com/influxdata/telegraf/issues/10473))
- remove duplicate addition of fields ([#10478](https://github.com/influxdata/telegraf/issues/10478))
- snmp input plugin errors if mibs folder doesn't exist ([#10346](https://github.com/influxdata/telegraf/issues/10346)) ([#10354](https://github.com/influxdata/telegraf/issues/10354))
- typo in docs ([#10441](https://github.com/influxdata/telegraf/issues/10441))
- update containerd to 1.5.9 ([#10402](https://github.com/influxdata/telegraf/issues/10402))
- **json_v2:** use raw values for timestamps ([#10413](https://github.com/influxdata/telegraf/issues/10413))
- wavefront_disable_prefix_conversion case missing from missingTomlField func ([#10442](https://github.com/influxdata/telegraf/issues/10442))


Thank you for your contributions!



Nirmesh ,
Sven Rebhan ,
reimda ,
Joshua Powers ,
Sebastian Spaink ,
dependabot[bot] ,
Mya ,
Nathan J Mehl ,
Sakerdotes ,
R290 ,
Laurent Sesquès ,
zachmares ,


<a name="v1.21.2"></a>
## [v1.21.2] - 2022-01-05

### Bug Fixes

- Fix panic in logparser due to missing Log. ([#10296](https://github.com/influxdata/telegraf/issues/10296))
- Fix panic in parsers due to missing Log for all plugins using SetParserFunc. ([#10288](https://github.com/influxdata/telegraf/issues/10288))
- Update go-sensu to v2.12.0 ([#10247](https://github.com/influxdata/telegraf/issues/10247))
- check index before assignment ([#10299](https://github.com/influxdata/telegraf/issues/10299))
- do not require networking during tests ([#10321](https://github.com/influxdata/telegraf/issues/10321))
- empty import tzdata for Windows binaries ([#10377](https://github.com/influxdata/telegraf/issues/10377))
- fix missing storage in container with disk plugin ([#10318](https://github.com/influxdata/telegraf/issues/10318))
- grab table columns more accurately ([#10295](https://github.com/influxdata/telegraf/issues/10295))
- mac signing issue with arm64 ([#10293](https://github.com/influxdata/telegraf/issues/10293))
- mark TestGatherUDPCert as an integration test ([#10279](https://github.com/influxdata/telegraf/issues/10279))
- panic due to no module ([#10303](https://github.com/influxdata/telegraf/issues/10303))
- panic is no mibs folder is found ([#10301](https://github.com/influxdata/telegraf/issues/10301))
- snmp marshal error ([#10322](https://github.com/influxdata/telegraf/issues/10322))
- typo in openstack neutron input plugin (newtron) ([#10284](https://github.com/influxdata/telegraf/issues/10284))
- update GroundWork SDK and improve logging ([#10255](https://github.com/influxdata/telegraf/issues/10255))
- update bug template
- update djherbis/times and fix dependabot ([#10332](https://github.com/influxdata/telegraf/issues/10332))
- update go-ldap to v3.4.1 ([#10343](https://github.com/influxdata/telegraf/issues/10343))
- warning output when running with --test ([#10329](https://github.com/influxdata/telegraf/issues/10329))

### Features

- Add SMART plugin concurrency configuration option, nvme-cli v1.14+ support and lint fixes. ([#10150](https://github.com/influxdata/telegraf/issues/10150))
- add builds for riscv64 ([#10262](https://github.com/influxdata/telegraf/issues/10262))
- support darwin arm64 ([#10239](https://github.com/influxdata/telegraf/issues/10239))


Thank you for your contributions!



Sven Rebhan ,
Sergey Vilgelm ,
Mya ,
Joshua Powers ,
Sebastian Spaink ,
Laurent Sesquès ,
Vladislav ,
Kuba Trojan ,


<a name="v1.21.1"></a>
## [v1.21.1] - 2021-12-16

### Bug Fixes

- Fix panic in parsers due to missing Log for all plugins using SetParserFunc. ([#10288](https://github.com/influxdata/telegraf/issues/10288))
- Update go-sensu to v2.12.0 ([#10247](https://github.com/influxdata/telegraf/issues/10247))
- mark TestGatherUDPCert as an integration test ([#10279](https://github.com/influxdata/telegraf/issues/10279))
- typo in openstack neutron input plugin (newtron) ([#10284](https://github.com/influxdata/telegraf/issues/10284))

### Features

- Add SMART plugin concurrency configuration option, nvme-cli v1.14+ support and lint fixes. ([#10150](https://github.com/influxdata/telegraf/issues/10150))
- support darwin arm64 ([#10239](https://github.com/influxdata/telegraf/issues/10239))


Thank you for your contributions!



Sven Rebhan ,
Sergey Vilgelm ,
Sebastian Spaink ,
Laurent Sesquès ,
Kuba Trojan ,
Mya ,


<a name="v1.21.0"></a>
## [v1.21.0] - 2021-12-15

### Bug Fixes

- Add error message logging to outputs.http ([#9727](https://github.com/influxdata/telegraf/issues/9727))
- Add metric name is a label with name "__name" to Loki output plugin ([#10001](https://github.com/influxdata/telegraf/issues/10001))
- Add setting to win_perf_counters input to ignore localization ([#10101](https://github.com/influxdata/telegraf/issues/10101))
- Allow for non x86 macs in Go install script ([#9982](https://github.com/influxdata/telegraf/issues/9982))
- Changed VM ID from string to int ([#10068](https://github.com/influxdata/telegraf/issues/10068))
- Check return code of zfs command for FreeBSD. ([#9956](https://github.com/influxdata/telegraf/issues/9956))
- Correct conversion of int with specific bit size ([#9933](https://github.com/influxdata/telegraf/issues/9933))
- Couchbase insecure certificate validation ([#9458](https://github.com/influxdata/telegraf/issues/9458))
- Fix panic for non-existing metric names ([#9757](https://github.com/influxdata/telegraf/issues/9757))
- Graylog plugin TLS support and message format ([#9862](https://github.com/influxdata/telegraf/issues/9862))
- Handle duplicate registration of protocol-buffer files gracefully. ([#10188](https://github.com/influxdata/telegraf/issues/10188))
- Implement NaN and inf handling for elasticsearch output ([#10196](https://github.com/influxdata/telegraf/issues/10196))
- Linter fixes for plugins/aggregators/[a-z]* ([#10182](https://github.com/influxdata/telegraf/issues/10182))
- Linter fixes for plugins/common/[a-z]* ([#10189](https://github.com/influxdata/telegraf/issues/10189))
- Linter fixes for plugins/inputs/[a-o]* (leftovers) ([#10192](https://github.com/influxdata/telegraf/issues/10192))
- Linter fixes for plugins/inputs/[h-j]* ([#9986](https://github.com/influxdata/telegraf/issues/9986))
- Linter fixes for plugins/inputs/[k-l]* ([#9999](https://github.com/influxdata/telegraf/issues/9999))
- Linter fixes for plugins/inputs/[n-o]* ([#10011](https://github.com/influxdata/telegraf/issues/10011))
- Linter fixes for plugins/inputs/[p-z]* (leftovers) ([#10193](https://github.com/influxdata/telegraf/issues/10193))
- Linter fixes for plugins/inputs/[t-z]* ([#10105](https://github.com/influxdata/telegraf/issues/10105))
- Linter fixes for plugins/inputs/m* ([#10006](https://github.com/influxdata/telegraf/issues/10006))
- Linter fixes for plugins/inputs/p* ([#10066](https://github.com/influxdata/telegraf/issues/10066))
- Linter fixes for plugins/inputs/s* ([#10104](https://github.com/influxdata/telegraf/issues/10104))
- Linter fixes for plugins/outputs/[a-f]* ([#10124](https://github.com/influxdata/telegraf/issues/10124))
- Linter fixes for plugins/outputs/[g-m]* ([#10127](https://github.com/influxdata/telegraf/issues/10127))
- Linter fixes for plugins/outputs/[p-z]* ([#10139](https://github.com/influxdata/telegraf/issues/10139))
- Linter fixes for plugins/parsers/[a-z]* ([#10145](https://github.com/influxdata/telegraf/issues/10145))
- Linter fixes for plugins/processors/[a-z]* ([#10161](https://github.com/influxdata/telegraf/issues/10161))
- Linter fixes for plugins/serializers/[a-z]* ([#10181](https://github.com/influxdata/telegraf/issues/10181))
- Markdown linter fixes for LICENSE_OF_DEPENDENCIES.md ([#10065](https://github.com/influxdata/telegraf/issues/10065))
- Print loaded plugins and deprecations for once and test ([#10205](https://github.com/influxdata/telegraf/issues/10205))
- Rename KNXListener to knx_listener ([#9741](https://github.com/influxdata/telegraf/issues/9741))
- Revert "Reset the flush interval timer when flush is requested or batch is ready. ([#8953](https://github.com/influxdata/telegraf/issues/8953))" ([#9800](https://github.com/influxdata/telegraf/issues/9800))
- Revert unintented corruption of the Makefile from [#10200](https://github.com/influxdata/telegraf/issues/10200). ([#10203](https://github.com/influxdata/telegraf/issues/10203))
- Set the default value correctly ([#9980](https://github.com/influxdata/telegraf/issues/9980))
- Sudden close of Telegraf caused by OPC UA input plugin ([#10230](https://github.com/influxdata/telegraf/issues/10230))
- Update gopcua library to latest version ([#9560](https://github.com/influxdata/telegraf/issues/9560))
- add additional logstash output plugin stats ([#9707](https://github.com/influxdata/telegraf/issues/9707))
- add keep alive config option, add documentation around issue with eclipse/mosquitto version combined with this plugin, update test ([#9803](https://github.com/influxdata/telegraf/issues/9803))
- add normalization of tags for ethtool input plugin ([#9901](https://github.com/influxdata/telegraf/issues/9901))
- add s390x to nightlies ([#9990](https://github.com/influxdata/telegraf/issues/9990))
- bump cloud.google.com/go/pubsub from 1.15.0 to 1.17.0 ([#9769](https://github.com/influxdata/telegraf/issues/9769))
- bump github.com/Azure/azure-event-hubs-go/v3 from 3.2.0 to 3.3.13 ([#9677](https://github.com/influxdata/telegraf/issues/9677))
- bump github.com/Azure/azure-kusto-go from 0.3.2 to 0.4.0 ([#9768](https://github.com/influxdata/telegraf/issues/9768))
- bump github.com/Azure/go-autorest/autorest/adal ([#9791](https://github.com/influxdata/telegraf/issues/9791))
- bump github.com/Azure/go-autorest/autorest/adal from 0.9.10->0.9.15
- bump github.com/Azure/go-autorest/autorest/azure/auth from 0.5.6 to 0.5.8 ([#9678](https://github.com/influxdata/telegraf/issues/9678))
- bump github.com/antchfx/xmlquery from 1.3.5 to 1.3.6 ([#9750](https://github.com/influxdata/telegraf/issues/9750))
- bump github.com/apache/thrift from 0.14.2 to 0.15.0 ([#9921](https://github.com/influxdata/telegraf/issues/9921))
- bump github.com/aws/aws-sdk-go-v2/config from 1.8.2 to 1.8.3 ([#9948](https://github.com/influxdata/telegraf/issues/9948))
- bump github.com/aws/smithy-go from 1.3.1 to 1.8.0 ([#9770](https://github.com/influxdata/telegraf/issues/9770))
- bump github.com/docker/docker from 20.10.7+incompatible to 20.10.9+incompatible ([#9905](https://github.com/influxdata/telegraf/issues/9905))
- bump github.com/eclipse/paho.mqtt.golang from 1.3.0 to 1.3.5 ([#9913](https://github.com/influxdata/telegraf/issues/9913))
- bump github.com/golang-jwt/jwt/v4 from 4.0.0 to 4.1.0 ([#9904](https://github.com/influxdata/telegraf/issues/9904))
- bump github.com/miekg/dns from 1.1.31 to 1.1.43 ([#9656](https://github.com/influxdata/telegraf/issues/9656))
- bump github.com/prometheus/client_golang from 1.7.1 to 1.11.0 ([#9653](https://github.com/influxdata/telegraf/issues/9653))
- bump github.com/prometheus/common from 0.26.0 to 0.31.1 ([#9869](https://github.com/influxdata/telegraf/issues/9869))
- bump github.com/shirou/gopsutil ([#9760](https://github.com/influxdata/telegraf/issues/9760))
- bump github.com/testcontainers/testcontainers-go from 0.11.0 to 0.11.1 ([#9789](https://github.com/influxdata/telegraf/issues/9789))
- bump google.golang.org/grpc from 1.39.1 to 1.40.0 ([#9751](https://github.com/influxdata/telegraf/issues/9751))
- bump k8s.io/apimachinery from 0.21.1 to 0.22.2 ([#9776](https://github.com/influxdata/telegraf/issues/9776))
- **inputs.tail:** change test default watch method to poll when Win
- check error before defer in prometheus k8s ([#10091](https://github.com/influxdata/telegraf/issues/10091))
- correct timezone in intel rdt plugin ([#10026](https://github.com/influxdata/telegraf/issues/10026))
- decode Prometheus scrape path from Kuberentes labels ([#9662](https://github.com/influxdata/telegraf/issues/9662))
- directory monitor input plugin when data format is CSV and csv_skip_rows>0 and csv_header_row_count>=1 ([#9865](https://github.com/influxdata/telegraf/issues/9865))
- do not build modbus on openbsd ([#10047](https://github.com/influxdata/telegraf/issues/10047))
- duplicate line_protocol when using object and fields ([#9872](https://github.com/influxdata/telegraf/issues/9872))
- eliminate MIB dependency for ifname processor ([#10214](https://github.com/influxdata/telegraf/issues/10214))
- ensure graylog spec fields not prefixed with '_' ([#10209](https://github.com/influxdata/telegraf/issues/10209))
- error returned to OpenTelemetry client ([#9797](https://github.com/influxdata/telegraf/issues/9797))
- extra lock on init for safety ([#10199](https://github.com/influxdata/telegraf/issues/10199))
- failing ci on master ([#10175](https://github.com/influxdata/telegraf/issues/10175))
- **http_listener_v2:** fix panic on close ([#10132](https://github.com/influxdata/telegraf/issues/10132))
- flush wavefront output sender on error to clean up broken connections ([#10225](https://github.com/influxdata/telegraf/issues/10225))
- gitignore should ignore .toml/.conf files ([#9818](https://github.com/influxdata/telegraf/issues/9818))
- inconsistent metric types in mysql ([#9403](https://github.com/influxdata/telegraf/issues/9403))
- input plugin statsd bug ([#10116](https://github.com/influxdata/telegraf/issues/10116))
- inputs.snmp to respect number of retries configured ([#10268](https://github.com/influxdata/telegraf/issues/10268))
- internet_speed input plugin not collecting/reporting latency ([#9957](https://github.com/influxdata/telegraf/issues/9957))
- json_v2 parser timestamp setting ([#10221](https://github.com/influxdata/telegraf/issues/10221))
- logging in intel_rdt.go caused service stop timeout even as root ([#9844](https://github.com/influxdata/telegraf/issues/9844)) ([#9850](https://github.com/influxdata/telegraf/issues/9850))
- makefile missing space for i386 tar and rpm ([#9887](https://github.com/influxdata/telegraf/issues/9887))
- markdown: resolve all markdown issues with a-c ([#10169](https://github.com/influxdata/telegraf/issues/10169))
- markdown: resolve all markdown issues with d-f ([#10171](https://github.com/influxdata/telegraf/issues/10171))
- markdown: resolve all markdown issues with g-h ([#10172](https://github.com/influxdata/telegraf/issues/10172))
- memory leak in influx parser ([#9787](https://github.com/influxdata/telegraf/issues/9787))
- migrate aws/credentials.go to use NewSession, same functionality but now supports error ([#9878](https://github.com/influxdata/telegraf/issues/9878))
- migrate to cloud.google.com/go/monitoring/apiv3/v2 ([#9880](https://github.com/influxdata/telegraf/issues/9880))
- mongodb input plugin issue [#9845](https://github.com/influxdata/telegraf/issues/9845) ([#9846](https://github.com/influxdata/telegraf/issues/9846))
- mqtt topic extracting no longer requires all three fields ([#10208](https://github.com/influxdata/telegraf/issues/10208))
- mute graylog UDP/TCP tests by marking them as integration ([#9881](https://github.com/influxdata/telegraf/issues/9881))
- mysql: type conversion follow-up ([#9966](https://github.com/influxdata/telegraf/issues/9966))
- nightly upload requires package steps ([#9795](https://github.com/influxdata/telegraf/issues/9795))
- outputs.opentelemetry use attributes setting ([#9588](https://github.com/influxdata/telegraf/issues/9588))
- pagination error on cloudwatch plugin ([#9693](https://github.com/influxdata/telegraf/issues/9693))
- parallelism fix for ifname processor ([#10007](https://github.com/influxdata/telegraf/issues/10007))
- patched intel rdt to allow sudo ([#9527](https://github.com/influxdata/telegraf/issues/9527))
- pool detection and metrics gathering for ZFS >= 2.1.x ([#10099](https://github.com/influxdata/telegraf/issues/10099))
- procstat missing tags in procstat_lookup metric ([#9808](https://github.com/influxdata/telegraf/issues/9808))
- procstat tags were not getting generated correctly ([#9973](https://github.com/influxdata/telegraf/issues/9973))
- redacts IPMI password in logs ([#9997](https://github.com/influxdata/telegraf/issues/9997))
- register bigquery to output plugins [#10177](https://github.com/influxdata/telegraf/issues/10177) ([#10178](https://github.com/influxdata/telegraf/issues/10178))
- **json_v2:** remove dead code ([#9908](https://github.com/influxdata/telegraf/issues/9908))
- remove eg fix: which breaks label bot functionality ([#9859](https://github.com/influxdata/telegraf/issues/9859))
- remove release.sh script ([#10030](https://github.com/influxdata/telegraf/issues/10030))
- remove telegraflinter from in-tree ([#10053](https://github.com/influxdata/telegraf/issues/10053))
- removed snmptranslate from readme and fix default path ([#10136](https://github.com/influxdata/telegraf/issues/10136))
- resolve [#10027](https://github.com/influxdata/telegraf/issues/10027) ([#10112](https://github.com/influxdata/telegraf/issues/10112))
- run go mod tidy
- run go mod tidy ([#10273](https://github.com/influxdata/telegraf/issues/10273))
- run gofmt ([#10274](https://github.com/influxdata/telegraf/issues/10274))
- segfault in ingress, persistentvolumeclaim, statefulset in kube_inventory ([#9585](https://github.com/influxdata/telegraf/issues/9585))
- set NIGHTLY=1 for correctly named nightly artifacts ([#9987](https://github.com/influxdata/telegraf/issues/9987))
- set location for timezone on failing time tests ([#9877](https://github.com/influxdata/telegraf/issues/9877))
- skip knxlistener when writing the sample config ([#10131](https://github.com/influxdata/telegraf/issues/10131))
- solve compatibility issue for mongodb inputs when using 5.x relicaset ([#9892](https://github.com/influxdata/telegraf/issues/9892))
- starlark pop operation for non-existing keys ([#9954](https://github.com/influxdata/telegraf/issues/9954))
- stop triggering share-artifacts on release/tags ([#9996](https://github.com/influxdata/telegraf/issues/9996))
- super-linter use v4.8.1, issue with latest ([#10108](https://github.com/influxdata/telegraf/issues/10108))
- sysstat use unique temp file vs hard-coded ([#10165](https://github.com/influxdata/telegraf/issues/10165))
- update BurntSushi/toml for hex config support ([#10089](https://github.com/influxdata/telegraf/issues/10089))
- update gjson to v1.10.2 ([#9998](https://github.com/influxdata/telegraf/issues/9998))
- update golanci-lint to v1.42.1 ([#9932](https://github.com/influxdata/telegraf/issues/9932))
- update golang-ci package ([#9817](https://github.com/influxdata/telegraf/issues/9817))
- update influxdb input schema documentation ([#10029](https://github.com/influxdata/telegraf/issues/10029))
- update makefile indents to not always run which ([#10126](https://github.com/influxdata/telegraf/issues/10126))
- update nats-sever to support openbsd ([#10046](https://github.com/influxdata/telegraf/issues/10046))
- update readme to align with other docs ([#10005](https://github.com/influxdata/telegraf/issues/10005))
- update readme.md to point at latest docs URL
- update shirou/gopsutil to v3 ([#10119](https://github.com/influxdata/telegraf/issues/10119))
- update toml tag to match sample config / readme ([#9848](https://github.com/influxdata/telegraf/issues/9848))
- use sha256 for RPM digest ([#10272](https://github.com/influxdata/telegraf/issues/10272))
- windows service - graceful shutdown of telegraf ([#9616](https://github.com/influxdata/telegraf/issues/9616))

### Features

- Add json_timestamp_layout option ([#8229](https://github.com/influxdata/telegraf/issues/8229))
- Add more details to processors.ifname logmessages ([#9984](https://github.com/influxdata/telegraf/issues/9984))
- Add support of aggregator as Starlark script ([#9419](https://github.com/influxdata/telegraf/issues/9419))
- Add use_batch_format for HTTP output plugin ([#8184](https://github.com/influxdata/telegraf/issues/8184))
- Adds the ability to create and name a tag containing the filename using the directory monitor input plugin ([#9860](https://github.com/influxdata/telegraf/issues/9860))
- Allow user to select the source for the metric timestamp. ([#9013](https://github.com/influxdata/telegraf/issues/9013))
- Azure Event Hubs output plugin ([#9346](https://github.com/influxdata/telegraf/issues/9346))
- Extend regexp processor do allow renaming of measurements, tags and fields ([#9561](https://github.com/influxdata/telegraf/issues/9561))
- Implement deprecation infrastructure ([#10200](https://github.com/influxdata/telegraf/issues/10200))
- Internet Speed Monitor Input Plugin ([#9623](https://github.com/influxdata/telegraf/issues/9623))
- Kafka Add metadata full to config ([#9833](https://github.com/influxdata/telegraf/issues/9833))
- Modbus connection settings (serial) ([#9256](https://github.com/influxdata/telegraf/issues/9256))
- Openstack input plugin ([#9236](https://github.com/influxdata/telegraf/issues/9236))
- Optimize locking for SNMP MIBs loading. ([#10206](https://github.com/influxdata/telegraf/issues/10206))
- Starlark processor example for processing sparkplug_b messages ([#9513](https://github.com/influxdata/telegraf/issues/9513))
- add Linux Volume Manager input plugin ([#9771](https://github.com/influxdata/telegraf/issues/9771))
- add additional metrics to support elastic pool (sqlserver plugin) ([#9841](https://github.com/influxdata/telegraf/issues/9841))
- add count of bonded slaves (for easier alerting) ([#9762](https://github.com/influxdata/telegraf/issues/9762))
- add custom time/date format field for elasticsearch_query ([#9838](https://github.com/influxdata/telegraf/issues/9838))
- add debug query output to elasticsearch_query ([#9827](https://github.com/influxdata/telegraf/issues/9827))
- **inputs.win_services:** add exclude filter ([#10144](https://github.com/influxdata/telegraf/issues/10144))
- add graylog plugin TCP support ([#9644](https://github.com/influxdata/telegraf/issues/9644))
- **prometheus:** add ignore_timestamp option ([#9740](https://github.com/influxdata/telegraf/issues/9740))
- add intel_pmu plugin ([#9724](https://github.com/influxdata/telegraf/issues/9724))
- add max_processing_time config to Kafka Consumer input ([#9988](https://github.com/influxdata/telegraf/issues/9988))
- add measurements from puppet 5 ([#9706](https://github.com/influxdata/telegraf/issues/9706))
- add mongodb output plugin  ([#9923](https://github.com/influxdata/telegraf/issues/9923))
- add new groundwork output plugin ([#9891](https://github.com/influxdata/telegraf/issues/9891))
- add option to skip table creation in azure data explorer output ([#9942](https://github.com/influxdata/telegraf/issues/9942))
- add retry to 413 errors with InfluxDB output ([#10130](https://github.com/influxdata/telegraf/issues/10130))
- enable extracting tag values from MQTT topics ([#9995](https://github.com/influxdata/telegraf/issues/9995))
- more fields for papertrail event webhook ([#9940](https://github.com/influxdata/telegraf/issues/9940))
- plugins/common/tls/config.go: Filter client certificates by DNS names ([#9910](https://github.com/influxdata/telegraf/issues/9910))
- **dynatrace-output:** remove special handling from counters ([#9675](https://github.com/influxdata/telegraf/issues/9675))
- telegraf to merge tables with different indexes ([#9241](https://github.com/influxdata/telegraf/issues/9241))


Thank you for your contributions!



Goutham Veeramachaneni ,
AlphaAr ,
reimda ,
Alan Pope ,
atetevoortwis ,
Sven Rebhan ,
Sebastian Spaink ,
Alexander Krantz ,
alespour ,
Paweł Żak ,
Joshua Powers ,
Fan Zhang ,
R290 ,
John Seekins ,
Helen Weller ,
dependabot[bot] ,
trojanku ,
Ehsan ,
Thomas Casteleyn ,
Jacob Marble ,
Mya ,
Patryk Małek ,
Mikołaj Przybysz ,
Felix Edelmann ,
rentiansheng ,
Sanyam Arya ,
Guo Qiao (Joe) ,
Patrick Hemmer ,
Howard Yoo ,
alrex ,
Doron-Bargo ,
xavpaice ,
Aaron Wood ,
Robert Thein ,
n2N8Z ,
Aleksandr Venger ,
alon ,
David B ,
Scott Anderson ,
Pierre Fersing ,
JC ,
Heiko Schlittermann ,
Nicolas Filotto ,
etycomputer ,
Thomas Conté ,
Gerald Quintana ,
singamSrikar ,
James Sorensen ,
Jean-Sébastien Dupuy ,
Yuji Kawamoto ,
bkotlowski ,
Chris Ruscio ,
bustedware ,
Vladislav ,
AsafMah ,
Sam Arnold ,
Josef Johansson ,
Daniel Dyla ,
helotpl ,### Reverts
- Merge branch 'master' into master
- add netflow plugin




<a name="v1.20.4"></a>
## [v1.20.4] - 2021-11-17

### Bug Fixes

- Add metric name is a label with name "__name" to Loki output plugin ([#10001](https://github.com/influxdata/telegraf/issues/10001))
- Changed VM ID from string to int ([#10068](https://github.com/influxdata/telegraf/issues/10068))
- Linter fixes for plugins/inputs/[h-j]* ([#9986](https://github.com/influxdata/telegraf/issues/9986))
- Linter fixes for plugins/inputs/[k-l]* ([#9999](https://github.com/influxdata/telegraf/issues/9999))
- Linter fixes for plugins/inputs/[n-o]* ([#10011](https://github.com/influxdata/telegraf/issues/10011))
- Linter fixes for plugins/inputs/m* ([#10006](https://github.com/influxdata/telegraf/issues/10006))
- Markdown linter fixes for LICENSE_OF_DEPENDENCIES.md ([#10065](https://github.com/influxdata/telegraf/issues/10065))
- Set the default value correctly ([#9980](https://github.com/influxdata/telegraf/issues/9980))
- correct timezone in intel rdt plugin ([#10026](https://github.com/influxdata/telegraf/issues/10026))
- do not build modbus on openbsd ([#10047](https://github.com/influxdata/telegraf/issues/10047))
- mysql: type conversion follow-up ([#9966](https://github.com/influxdata/telegraf/issues/9966))
- remove release.sh script ([#10030](https://github.com/influxdata/telegraf/issues/10030))
- remove telegraflinter from in-tree ([#10053](https://github.com/influxdata/telegraf/issues/10053))
- super-linter use v4.8.1, issue with latest ([#10108](https://github.com/influxdata/telegraf/issues/10108))
- update BurntSushi/toml for hex config support ([#10089](https://github.com/influxdata/telegraf/issues/10089))
- update influxdb input schema documentation ([#10029](https://github.com/influxdata/telegraf/issues/10029))
- update readme.md to point at latest docs URL


Thank you for your contributions!



AlphaAr ,
atetevoortwis ,
Paweł Żak ,
Fan Zhang ,
trojanku ,
Joshua Powers ,
Felix Edelmann ,
Sebastian Spaink ,
David B ,
Scott Anderson ,


<a name="v1.20.3"></a>
## [v1.20.3] - 2021-10-27

### Bug Fixes

- Allow for non x86 macs in Go install script ([#9982](https://github.com/influxdata/telegraf/issues/9982))
- Check return code of zfs command for FreeBSD. ([#9956](https://github.com/influxdata/telegraf/issues/9956))
- Correct conversion of int with specific bit size ([#9933](https://github.com/influxdata/telegraf/issues/9933))
- add normalization of tags for ethtool input plugin ([#9901](https://github.com/influxdata/telegraf/issues/9901))
- add s390x to nightlies ([#9990](https://github.com/influxdata/telegraf/issues/9990))
- bump github.com/Azure/azure-kusto-go from 0.3.2 to 0.4.0 ([#9768](https://github.com/influxdata/telegraf/issues/9768))
- bump github.com/apache/thrift from 0.14.2 to 0.15.0 ([#9921](https://github.com/influxdata/telegraf/issues/9921))
- bump github.com/aws/aws-sdk-go-v2/config from 1.8.2 to 1.8.3 ([#9948](https://github.com/influxdata/telegraf/issues/9948))
- bump github.com/docker/docker from 20.10.7+incompatible to 20.10.9+incompatible ([#9905](https://github.com/influxdata/telegraf/issues/9905))
- bump github.com/golang-jwt/jwt/v4 from 4.0.0 to 4.1.0 ([#9904](https://github.com/influxdata/telegraf/issues/9904))
- bump github.com/prometheus/common from 0.26.0 to 0.31.1 ([#9869](https://github.com/influxdata/telegraf/issues/9869))
- decode Prometheus scrape path from Kuberentes labels ([#9662](https://github.com/influxdata/telegraf/issues/9662))
- inconsistent metric types in mysql ([#9403](https://github.com/influxdata/telegraf/issues/9403))
- internet_speed input plugin not collecting/reporting latency ([#9957](https://github.com/influxdata/telegraf/issues/9957))
- patched intel rdt to allow sudo ([#9527](https://github.com/influxdata/telegraf/issues/9527))
- procstat tags were not getting generated correctly ([#9973](https://github.com/influxdata/telegraf/issues/9973))
- redacts IPMI password in logs ([#9997](https://github.com/influxdata/telegraf/issues/9997))
- segfault in ingress, persistentvolumeclaim, statefulset in kube_inventory ([#9585](https://github.com/influxdata/telegraf/issues/9585))
- set NIGHTLY=1 for correctly named nightly artifacts ([#9987](https://github.com/influxdata/telegraf/issues/9987))
- solve compatibility issue for mongodb inputs when using 5.x relicaset ([#9892](https://github.com/influxdata/telegraf/issues/9892))
- starlark pop operation for non-existing keys ([#9954](https://github.com/influxdata/telegraf/issues/9954))
- stop triggering share-artifacts on release/tags ([#9996](https://github.com/influxdata/telegraf/issues/9996))
- update gjson to v1.10.2 ([#9998](https://github.com/influxdata/telegraf/issues/9998))
- update golanci-lint to v1.42.1 ([#9932](https://github.com/influxdata/telegraf/issues/9932))
- update readme to align with other docs ([#10005](https://github.com/influxdata/telegraf/issues/10005))

### Features

- more fields for papertrail event webhook ([#9940](https://github.com/influxdata/telegraf/issues/9940))


Thank you for your contributions!



Alan Pope ,
Sven Rebhan ,
Sebastian Spaink ,
Joshua Powers ,
dependabot[bot] ,
Alexander Krantz ,
Felix Edelmann ,
Sanyam Arya ,
xavpaice ,
Aleksandr Venger ,
alon ,
Sam Arnold ,


<a name="v1.20.2"></a>
## [v1.20.2] - 2021-10-07

### Bug Fixes

- duplicate line_protocol when using object and fields ([#9872](https://github.com/influxdata/telegraf/issues/9872))
- makefile missing space for i386 tar and rpm ([#9887](https://github.com/influxdata/telegraf/issues/9887))
- memory leak in influx parser ([#9787](https://github.com/influxdata/telegraf/issues/9787))
- migrate aws/credentials.go to use NewSession, same functionality but now supports error ([#9878](https://github.com/influxdata/telegraf/issues/9878))
- migrate to cloud.google.com/go/monitoring/apiv3/v2 ([#9880](https://github.com/influxdata/telegraf/issues/9880))
- set location for timezone on failing time tests ([#9877](https://github.com/influxdata/telegraf/issues/9877))


Thank you for your contributions!



Sebastian Spaink ,
Joshua Powers ,
Patrick Hemmer ,


<a name="v1.20.1"></a>
## [v1.20.1] - 2021-10-06

### Bug Fixes

- Couchbase insecure certificate validation ([#9458](https://github.com/influxdata/telegraf/issues/9458))
- Rename KNXListener to knx_listener ([#9741](https://github.com/influxdata/telegraf/issues/9741))
- Revert "Reset the flush interval timer when flush is requested or batch is ready. ([#8953](https://github.com/influxdata/telegraf/issues/8953))" ([#9800](https://github.com/influxdata/telegraf/issues/9800))
- add keep alive config option, add documentation around issue with eclipse/mosquitto version combined with this plugin, update test ([#9803](https://github.com/influxdata/telegraf/issues/9803))
- bump cloud.google.com/go/pubsub from 1.15.0 to 1.17.0 ([#9769](https://github.com/influxdata/telegraf/issues/9769))
- bump github.com/Azure/go-autorest/autorest/adal ([#9791](https://github.com/influxdata/telegraf/issues/9791))
- bump github.com/Azure/go-autorest/autorest/azure/auth from 0.5.6 to 0.5.8 ([#9678](https://github.com/influxdata/telegraf/issues/9678))
- bump github.com/aws/smithy-go from 1.3.1 to 1.8.0 ([#9770](https://github.com/influxdata/telegraf/issues/9770))
- bump github.com/testcontainers/testcontainers-go from 0.11.0 to 0.11.1 ([#9789](https://github.com/influxdata/telegraf/issues/9789))
- bump k8s.io/apimachinery from 0.21.1 to 0.22.2 ([#9776](https://github.com/influxdata/telegraf/issues/9776))
- error returned to OpenTelemetry client ([#9797](https://github.com/influxdata/telegraf/issues/9797))
- gitignore should ignore .toml/.conf files ([#9818](https://github.com/influxdata/telegraf/issues/9818))
- logging in intel_rdt.go caused service stop timeout even as root ([#9844](https://github.com/influxdata/telegraf/issues/9844)) ([#9850](https://github.com/influxdata/telegraf/issues/9850))
- mongodb input plugin issue [#9845](https://github.com/influxdata/telegraf/issues/9845) ([#9846](https://github.com/influxdata/telegraf/issues/9846))
- nightly upload requires package steps ([#9795](https://github.com/influxdata/telegraf/issues/9795))
- procstat missing tags in procstat_lookup metric ([#9808](https://github.com/influxdata/telegraf/issues/9808))
- remove eg fix: which breaks label bot functionality ([#9859](https://github.com/influxdata/telegraf/issues/9859))
- run go mod tidy
- update golang-ci package ([#9817](https://github.com/influxdata/telegraf/issues/9817))
- update toml tag to match sample config / readme ([#9848](https://github.com/influxdata/telegraf/issues/9848))

### Features

- add custom time/date format field for elasticsearch_query ([#9838](https://github.com/influxdata/telegraf/issues/9838))


Thank you for your contributions!



Alexander Krantz ,
Sven Rebhan ,
Joshua Powers ,
Helen Weller ,
dependabot[bot] ,
Jacob Marble ,
Guo Qiao (Joe) ,
Howard Yoo ,
Sebastian Spaink ,


<a name="v1.20.0"></a>
## [v1.20.0] - 2021-09-17

### Bug Fixes

- Add error message logging to outputs.http ([#9727](https://github.com/influxdata/telegraf/issues/9727))
- Bump github.com/aws/aws-sdk-go-v2 from 1.3.2 to 1.8.0 ([#9636](https://github.com/influxdata/telegraf/issues/9636))
- Bump github.com/aws/aws-sdk-go-v2/config from 1.1.5 to 1.6.0
- Bump github.com/golang/snappy from 0.0.3 to 0.0.4 ([#9637](https://github.com/influxdata/telegraf/issues/9637))
- Bump github.com/sirupsen/logrus from 1.7.0 to 1.8.1 ([#9639](https://github.com/influxdata/telegraf/issues/9639))
- Bump github.com/testcontainers/testcontainers-go from 0.11.0 to 0.11.1 ([#9638](https://github.com/influxdata/telegraf/issues/9638))
- CrateDB replace dots in tag keys with underscores ([#9566](https://github.com/influxdata/telegraf/issues/9566))
- Do not return on disconnect to avoid breaking reconnect ([#9524](https://github.com/influxdata/telegraf/issues/9524))
- Fix panic for non-existing metric names ([#9757](https://github.com/influxdata/telegraf/issues/9757))
- Fixing k8s nodes and pods parsing error ([#9581](https://github.com/influxdata/telegraf/issues/9581))
- Normalize unix socket path ([#9554](https://github.com/influxdata/telegraf/issues/9554))
- Refactor ec2 init for config-api ([#9576](https://github.com/influxdata/telegraf/issues/9576))
- Update gopcua library to latest version ([#9560](https://github.com/influxdata/telegraf/issues/9560))
- Verify checksum of Go download in mac script ([#9335](https://github.com/influxdata/telegraf/issues/9335))
- add additional logstash output plugin stats ([#9707](https://github.com/influxdata/telegraf/issues/9707))
- bump cloud.google.com/go/pubsub from 1.2.0 to 1.15.0 ([#9655](https://github.com/influxdata/telegraf/issues/9655))
- bump github.com/Azure/azure-event-hubs-go/v3 from 3.2.0 to 3.3.13 ([#9677](https://github.com/influxdata/telegraf/issues/9677))
- bump github.com/Azure/go-autorest/autorest/adal from 0.9.10->0.9.15
- bump github.com/antchfx/xmlquery from 1.3.5 to 1.3.6 ([#9750](https://github.com/influxdata/telegraf/issues/9750))
- bump github.com/miekg/dns from 1.1.31 to 1.1.43 ([#9656](https://github.com/influxdata/telegraf/issues/9656))
- bump github.com/prometheus/client_golang from 1.7.1 to 1.11.0 ([#9653](https://github.com/influxdata/telegraf/issues/9653))
- bump github.com/shirou/gopsutil ([#9760](https://github.com/influxdata/telegraf/issues/9760))
- bump github.com/tinylib/msgp from 1.1.5 to 1.1.6 ([#9652](https://github.com/influxdata/telegraf/issues/9652))
- bump runc to v1.0.0-rc95 to address CVE-2021-30465 ([#9713](https://github.com/influxdata/telegraf/issues/9713))
- bump thrift to 0.14.2 and zipkin-go-opentracing 0.4.5 ([#9700](https://github.com/influxdata/telegraf/issues/9700))
- **mongodb:** change command based on server version ([#9674](https://github.com/influxdata/telegraf/issues/9674))
- **inputs.tail:** change test default watch method to poll when Win
- **opcua:** clean client on disconnect so that connect works cleanly ([#9583](https://github.com/influxdata/telegraf/issues/9583))
- cookie test ([#9608](https://github.com/influxdata/telegraf/issues/9608))
- improve Clickhouse corner cases for empty recordset in aggregation queries, fix dictionaries behavior ([#9401](https://github.com/influxdata/telegraf/issues/9401))
- issues with prometheus kubernetes pod discovery ([#9605](https://github.com/influxdata/telegraf/issues/9605))
- migrate dgrijalva/jwt-go to golang-jwt/jwt/v4 ([#9699](https://github.com/influxdata/telegraf/issues/9699))
- muting tests for udp_listener ([#9578](https://github.com/influxdata/telegraf/issues/9578))
- output timestamp with fractional seconds ([#9625](https://github.com/influxdata/telegraf/issues/9625))
- outputs.opentelemetry use attributes setting ([#9588](https://github.com/influxdata/telegraf/issues/9588))
- outputs.opentelemetry use headers config in grpc requests ([#9587](https://github.com/influxdata/telegraf/issues/9587))
- pagination error on cloudwatch plugin ([#9693](https://github.com/influxdata/telegraf/issues/9693))
- prefix dependabot commits with "fix:" ([#9641](https://github.com/influxdata/telegraf/issues/9641))
- race condition in cookie test ([#9659](https://github.com/influxdata/telegraf/issues/9659))
- **dt-output:** remove hardcoded int value ([#9676](https://github.com/influxdata/telegraf/issues/9676))
- run go fmt on inputs.mdstat with go1.17 ([#9702](https://github.com/influxdata/telegraf/issues/9702))
- sort logs by timestamp before writing to Loki ([#9571](https://github.com/influxdata/telegraf/issues/9571))
- support 1.17 & 1.16.7 Go versions ([#9642](https://github.com/influxdata/telegraf/issues/9642))
- upgraded sensu/go to v2.9.0 ([#9577](https://github.com/influxdata/telegraf/issues/9577))
- wireguard unknown revision when using direct ([#9620](https://github.com/influxdata/telegraf/issues/9620))

### Features

- Add rocm_smi input to monitor AMD GPUs ([#9602](https://github.com/influxdata/telegraf/issues/9602))
- Internet Speed Monitor Input Plugin ([#9623](https://github.com/influxdata/telegraf/issues/9623))
- Modbus Rtu over tcp enhancement ([#9570](https://github.com/influxdata/telegraf/issues/9570))
- OpenTelemetry output plugin ([#9228](https://github.com/influxdata/telegraf/issues/9228))
- Pull metrics from multiple AWS CloudWatch namespaces ([#9386](https://github.com/influxdata/telegraf/issues/9386))
- Support AWS Web Identity Provider ([#9411](https://github.com/influxdata/telegraf/issues/9411))
- add bool datatype for sql output plugin ([#9598](https://github.com/influxdata/telegraf/issues/9598))
- add count of bonded slaves (for easier alerting) ([#9762](https://github.com/influxdata/telegraf/issues/9762))
- add inputs.mdstat to gather from /proc/mdstat collection ([#9101](https://github.com/influxdata/telegraf/issues/9101))
- **http_listener_v2:** allows multiple paths and add path_tag ([#9529](https://github.com/influxdata/telegraf/issues/9529))
- **dynatrace-output:** remove special handling from counters ([#9675](https://github.com/influxdata/telegraf/issues/9675))


Thank you for your contributions!



Goutham Veeramachaneni ,
dependabot[bot] ,
Alexander Krantz ,
Sven Rebhan ,
varunjain0606 ,
Sebastian Spaink ,
pierwill ,
John Seekins ,
reimda ,
Marcus Ilgner ,
Eugene Klimov ,
Grace Wehner ,
alrex ,
Doron-Bargo ,
Daniel Dyla ,
JC ,
Helen Weller ,
Matteo Concas ,
Sanyam Arya ,
Marius Bezuidenhout ,
Jacob Marble ,
Nicolai Scheer ,
Dominik Rosiek ,


<a name="v1.19.3"></a>
## [v1.19.3] - 2021-08-18

### Bug Fixes

- Bump github.com/aws/aws-sdk-go-v2 from 1.3.2 to 1.8.0 ([#9636](https://github.com/influxdata/telegraf/issues/9636))
- Bump github.com/golang/snappy from 0.0.3 to 0.0.4 ([#9637](https://github.com/influxdata/telegraf/issues/9637))
- Bump github.com/sirupsen/logrus from 1.7.0 to 1.8.1 ([#9639](https://github.com/influxdata/telegraf/issues/9639))
- Bump github.com/testcontainers/testcontainers-go from 0.11.0 to 0.11.1 ([#9638](https://github.com/influxdata/telegraf/issues/9638))
- CrateDB replace dots in tag keys with underscores ([#9566](https://github.com/influxdata/telegraf/issues/9566))
- Do not return on disconnect to avoid breaking reconnect ([#9524](https://github.com/influxdata/telegraf/issues/9524))
- Fixing k8s nodes and pods parsing error ([#9581](https://github.com/influxdata/telegraf/issues/9581))
- Normalize unix socket path ([#9554](https://github.com/influxdata/telegraf/issues/9554))
- Refactor ec2 init for config-api ([#9576](https://github.com/influxdata/telegraf/issues/9576))
- **opcua:** clean client on disconnect so that connect works cleanly ([#9583](https://github.com/influxdata/telegraf/issues/9583))
- improve Clickhouse corner cases for empty recordset in aggregation queries, fix dictionaries behavior ([#9401](https://github.com/influxdata/telegraf/issues/9401))
- issues with prometheus kubernetes pod discovery ([#9605](https://github.com/influxdata/telegraf/issues/9605))
- muting tests for udp_listener ([#9578](https://github.com/influxdata/telegraf/issues/9578))
- sort logs by timestamp before writing to Loki ([#9571](https://github.com/influxdata/telegraf/issues/9571))
- upgraded sensu/go to v2.9.0 ([#9577](https://github.com/influxdata/telegraf/issues/9577))
- wireguard unknown revision when using direct ([#9620](https://github.com/influxdata/telegraf/issues/9620))


Thank you for your contributions!



dependabot[bot] ,
Alexander Krantz ,
Sven Rebhan ,
varunjain0606 ,
Sebastian Spaink ,
Marcus Ilgner ,
Eugene Klimov ,
Grace Wehner ,
JC ,
Helen Weller ,


<a name="v1.19.2"></a>
## [v1.19.2] - 2021-07-28

- Telegraf v1.19.2
- Update changelog
- update build version
- Update dynatrace output ([#9363](https://github.com/influxdata/telegraf/issues/9363))
- Fix metrics reported as written but not actually written  ([#9526](https://github.com/influxdata/telegraf/issues/9526))
- Prevent segfault in persistent volume claims ([#9549](https://github.com/influxdata/telegraf/issues/9549))
- Fix attempt to connect to an empty list of servers. ([#9503](https://github.com/influxdata/telegraf/issues/9503))
- Fix handling bool in sql input plugin ([#9540](https://github.com/influxdata/telegraf/issues/9540))
- Linter fixes for plugins/inputs/[fg]* ([#9387](https://github.com/influxdata/telegraf/issues/9387))
- [Docs] Clarify tagging behavior ([#9461](https://github.com/influxdata/telegraf/issues/9461))
- Attach the pod labels to the `kubernetes_pod_volume` & `kubernetes_pod_network` metrics. ([#9438](https://github.com/influxdata/telegraf/issues/9438))
- Bug Fix Snmp empty metric name ([#9519](https://github.com/influxdata/telegraf/issues/9519))
- Worktable workfile stats ([#8587](https://github.com/influxdata/telegraf/issues/8587))
- Update Go to v1.16.6 ([#9542](https://github.com/influxdata/telegraf/issues/9542))
- Prevent x509_cert from hanging on UDP connection ([#9323](https://github.com/influxdata/telegraf/issues/9323))
- Simplify how nesting is handled ([#9504](https://github.com/influxdata/telegraf/issues/9504))
- Switch MongoDB libraries ([#9493](https://github.com/influxdata/telegraf/issues/9493))
- [output dynatrace] Initialize loggedMetrics map ([#9491](https://github.com/influxdata/telegraf/issues/9491))
- Fix prometheus cadvisor authentication ([#9497](https://github.com/influxdata/telegraf/issues/9497))
- Add support for large uint64 and int64 numbers ([#9520](https://github.com/influxdata/telegraf/issues/9520))
- fixed percentiles not being able to be ints ([#9447](https://github.com/influxdata/telegraf/issues/9447))
- Detect changes to config and reload telegraf (copy of pr [#8529](https://github.com/influxdata/telegraf/issues/8529)) ([#9485](https://github.com/influxdata/telegraf/issues/9485))
- Provide detailed error message in telegraf log ([#9466](https://github.com/influxdata/telegraf/issues/9466))
- Update the dynatrace metric utils v0.1->v0.2 ([#9399](https://github.com/influxdata/telegraf/issues/9399))
- chore: fixing link in influxdb_listener plugin ([#9431](https://github.com/influxdata/telegraf/issues/9431))
- Allow multiple keys when parsing cgroups ([#8108](https://github.com/influxdata/telegraf/issues/8108))
- Fix json_v2 parser to handle nested objects in arrays properly ([#9479](https://github.com/influxdata/telegraf/issues/9479))
- Add s7comm external input plugin ([#9360](https://github.com/influxdata/telegraf/issues/9360))


Thank you for your contributions!





<a name="v1.19.1"></a>
## [v1.19.1] - 2021-07-07

- Telegraf v1.19.1
- Update changelog
- update build version
- Sqlserver input: require authentication method to be specified ([#9388](https://github.com/influxdata/telegraf/issues/9388))
- Improve documentation ([#9457](https://github.com/influxdata/telegraf/issues/9457))
- Fix typo in perDeviceIncludeDeprecationWarning ([#9442](https://github.com/influxdata/telegraf/issues/9442))
- Fix segfault in kube_inventory ([#9456](https://github.com/influxdata/telegraf/issues/9456))
- Fix Couchbase regression ([#9448](https://github.com/influxdata/telegraf/issues/9448))
- Fix nil pointer error in knx_listener ([#9444](https://github.com/influxdata/telegraf/issues/9444))
- add OpenTelemetry entry ([#9464](https://github.com/influxdata/telegraf/issues/9464))
- updated gopsutil to use a specific commit ([#9446](https://github.com/influxdata/telegraf/issues/9446))
- Fix RabbitMQ regression in [#9383](https://github.com/influxdata/telegraf/issues/9383) ([#9443](https://github.com/influxdata/telegraf/issues/9443))
- nat-server upgrade to v2.2.6 ([#9369](https://github.com/influxdata/telegraf/issues/9369))
- Exclude read-timeout from being an error ([#9429](https://github.com/influxdata/telegraf/issues/9429))
- Don't stop parsing after statsd parsing error ([#9423](https://github.com/influxdata/telegraf/issues/9423))
- apimachinary updated to v0.21.1 ([#9370](https://github.com/influxdata/telegraf/issues/9370))
- chore: readme updates ([#9367](https://github.com/influxdata/telegraf/issues/9367))
- updated jwt to v1.2.2 and updated jwt-go to v3.2.3 ([#9373](https://github.com/influxdata/telegraf/issues/9373))
- Update couchbase dependencies to v0.1.0 ([#9412](https://github.com/influxdata/telegraf/issues/9412))
- added a check for oid and name to prevent empty metrics ([#9366](https://github.com/influxdata/telegraf/issues/9366))
- fixing insecure_skip_verify ([#9413](https://github.com/influxdata/telegraf/issues/9413))
- Fix messing up the 'source' tag for https sources. ([#9400](https://github.com/influxdata/telegraf/issues/9400))
- Update signalfx to v3.3.0->v3.3.34 ([#9375](https://github.com/influxdata/telegraf/issues/9375))
- tags no longer required in included_keys ([#9406](https://github.com/influxdata/telegraf/issues/9406))
- Fix x509_cert input plugin SNI support ([#9289](https://github.com/influxdata/telegraf/issues/9289))
- gjson dependancy updated to v1.8.0 ([#9372](https://github.com/influxdata/telegraf/issues/9372))
- kube_inventory: expand tls key/tls certificate documentation  ([#9357](https://github.com/influxdata/telegraf/issues/9357))
- Adjust link to ceph documentation ([#9378](https://github.com/influxdata/telegraf/issues/9378))
- Linter fixes for plugins/inputs/[de]* ([#9379](https://github.com/influxdata/telegraf/issues/9379))


Thank you for your contributions!





<a name="v1.19.0"></a>
## [v1.19.0] - 2021-06-17

### Bug Fixes

- Beat readme title ([#8938](https://github.com/influxdata/telegraf/issues/8938))
- Verify checksum of Go download in mac script ([#9335](https://github.com/influxdata/telegraf/issues/9335))

### Features

- Add external Big blue button plugin ([#9090](https://github.com/influxdata/telegraf/issues/9090))
- Adding Plex Webhooks external plugin ([#8898](https://github.com/influxdata/telegraf/issues/8898))


Thank you for your contributions!



Russ Savage ,
pierwill ,
LEDUNOIS Simon ,


<a name="v1.18.3"></a>
## [v1.18.3] - 2021-05-21

- Telegraf v1.18.3
- Update changelog
- update build version
- Set user agent when scraping prom metrics ([#9271](https://github.com/influxdata/telegraf/issues/9271))
- Migrate soniah/gosnmp import to gosnmp/gosnmp ([#9203](https://github.com/influxdata/telegraf/issues/9203))
- Add Freebsd armv7 URL for nightly builds / organize ([#9268](https://github.com/influxdata/telegraf/issues/9268))
- Kinesis_consumer input plugin - fix repeating parser error ([#9169](https://github.com/influxdata/telegraf/issues/9169))
- SQL Server - sqlServerRingBufferCPU - removed whitespaces ([#9130](https://github.com/influxdata/telegraf/issues/9130))
- Add ability to enable gzip compression in elasticsearch output ([#8913](https://github.com/influxdata/telegraf/issues/8913))
- Upgrade hashicorp/consul/api to v1.8.1 ([#9238](https://github.com/influxdata/telegraf/issues/9238))
- Migrate ipvs library from docker/libnetwork/ipvs to moby/ipvs ([#9235](https://github.com/influxdata/telegraf/issues/9235))
- Document using group membership to allow access to /dev/pf for pf input plugin ([#9232](https://github.com/influxdata/telegraf/issues/9232))
- Add FreeBSD armv7 package ([#9200](https://github.com/influxdata/telegraf/issues/9200))
- Upgrade gopsutil to v3.21.3 ([#9224](https://github.com/influxdata/telegraf/issues/9224))
- Make microsoft lowercase ([#9209](https://github.com/influxdata/telegraf/issues/9209))
- upgrade gogo protobuf to v1.3.2 ([#9190](https://github.com/influxdata/telegraf/issues/9190))
- Bump github.com/Azure/go-autorest/autorest/azure/auth from 0.4.2 to 0.5.6 ([#8746](https://github.com/influxdata/telegraf/issues/8746))
- Bump collectd.org from 0.3.0 to 0.5.0 ([#8745](https://github.com/influxdata/telegraf/issues/8745))
- Bump github.com/nats-io/nats.go from 1.9.1 to 1.10.0 ([#8716](https://github.com/influxdata/telegraf/issues/8716))
- Change duplicate kubernetes import and update protobuf to v1.5.1 ([#9039](https://github.com/influxdata/telegraf/issues/9039))
- Migrate from github.com/ericchiang/k8s to github.com/kubernetes/client-go ([#8937](https://github.com/influxdata/telegraf/issues/8937))


Thank you for your contributions!





<a name="v1.18.2"></a>
## [v1.18.2] - 2021-04-30

- Telegraf v1.18.2
- append to package list instead of assigning
- don't use the parallel package build section during release builds
- go mod tidy
- Update changelog
- update build version
- Converter processor: add support for large hexadecimal strings ([#9160](https://github.com/influxdata/telegraf/issues/9160))
- Fix apcupsd 'ALARMDEL' bug via forked repo ([#9195](https://github.com/influxdata/telegraf/issues/9195))
- Make JSON format compatible with nulls ([#9110](https://github.com/influxdata/telegraf/issues/9110))
- Fix: sync nfsclient ops map with nfsclient struct ([#9128](https://github.com/influxdata/telegraf/issues/9128))
- Log snmpv3 auth failures ([#8917](https://github.com/influxdata/telegraf/issues/8917))
- Change to NewStreamParser to accept larger inputs from scanner ([#8892](https://github.com/influxdata/telegraf/issues/8892))
- Added MetricLookback setting ([#9045](https://github.com/influxdata/telegraf/issues/9045))
- remove deprecation warning ([#9125](https://github.com/influxdata/telegraf/issues/9125))
- Carbon2 serializer: sanitize metric name ([#9026](https://github.com/influxdata/telegraf/issues/9026))
- Delete log.Fatal calls and replace with error returns. ([#9086](https://github.com/influxdata/telegraf/issues/9086))
- ci config changes ([#9001](https://github.com/influxdata/telegraf/issues/9001))
- Parallelize PR builds by Architecture ([#9172](https://github.com/influxdata/telegraf/issues/9172))
- Speed up package step by running in parallel. ([#9096](https://github.com/influxdata/telegraf/issues/9096))


Thank you for your contributions!





<a name="v1.18.1"></a>
## [v1.18.1] - 2021-04-07

- Telegraf v1.18.1
- Update changelog
- update build version
- Add ability to handle 'binary logs' mySQL query with 3 columns, in case 3 columns are sent (MySQL 8 and greater) ([#9082](https://github.com/influxdata/telegraf/issues/9082))
- Add configurable option for the 'path' tag override in the Tail plugin. ([#9069](https://github.com/influxdata/telegraf/issues/9069))
- fix nfsclient merge to release-1.18 branch
- inputs.nfsclient: use uint64, also update error handling ([#9067](https://github.com/influxdata/telegraf/issues/9067))
- Fix inputs.snmp init when no mibs installed ([#9050](https://github.com/influxdata/telegraf/issues/9050))
- inputs.ping: Always SetPrivileged(true) in native mode ([#9072](https://github.com/influxdata/telegraf/issues/9072))
- Don't walk the entire interface table to just retrieve one field ([#9043](https://github.com/influxdata/telegraf/issues/9043))
- readme fix ([#9064](https://github.com/influxdata/telegraf/issues/9064))
- use correct compute metadata url to get folder-id ([#9056](https://github.com/influxdata/telegraf/issues/9056))
- Handle error when initializing the auth object in Azure Monitor output plugin. ([#9048](https://github.com/influxdata/telegraf/issues/9048))
- update: inputs.sqlserver support version in readme ([#9040](https://github.com/influxdata/telegraf/issues/9040))
- SQLServer - Fixes sqlserver_process_cpu calculation ([#8549](https://github.com/influxdata/telegraf/issues/8549))
- Fix ipmi panic ([#9035](https://github.com/influxdata/telegraf/issues/9035))
- check for length of perusage for stat gathering and removed not used function ([#9009](https://github.com/influxdata/telegraf/issues/9009))
- update new plugins in changelog ([#8991](https://github.com/influxdata/telegraf/issues/8991))
- exec plugins should not truncate messages in debug mode ([#8333](https://github.com/influxdata/telegraf/issues/8333))
- Close running outputs when reloading ([#8769](https://github.com/influxdata/telegraf/issues/8769))


Thank you for your contributions!





<a name="v1.18.0"></a>
## [v1.18.0] - 2021-03-17

### Bug Fixes

- Beat readme title ([#8938](https://github.com/influxdata/telegraf/issues/8938))
- reading multiple holding registers in modbus input plugin ([#8628](https://github.com/influxdata/telegraf/issues/8628))
- remove ambiguity on '\v' from line-protocol parser ([#8720](https://github.com/influxdata/telegraf/issues/8720))

### Features

- Adding Plex Webhooks external plugin ([#8898](https://github.com/influxdata/telegraf/issues/8898))


Thank you for your contributions!



Russ Savage ,
Antonio Garcia ,
Adrian Thurston ,### Reverts
- Update grok package to support for field names containing '-' and '.' ([#8276](https://github.com/influxdata/telegraf/issues/8276))
- disable flakey grok test for now




<a name="v1.17.3"></a>
## [v1.17.3] - 2021-02-17

- Telegraf v1.17.3
- Update changelog
- update build version
- plugins/filestat: Skip missing files ([#7316](https://github.com/influxdata/telegraf/issues/7316))
- Update to 1.15.8 ([#8868](https://github.com/influxdata/telegraf/issues/8868))
- Bump github.com/gopcua/opcua from 0.1.12 to 0.1.13 ([#8744](https://github.com/influxdata/telegraf/issues/8744))
- outputs/warp10: url encode comma in tags value ([#8657](https://github.com/influxdata/telegraf/issues/8657))
- inputs.x509_cert: Fix timeout issue  ([#8824](https://github.com/influxdata/telegraf/issues/8824))
- update min Go version in Telegraf readme ([#8846](https://github.com/influxdata/telegraf/issues/8846))
- Fix reconnection issues mqtt ([#8821](https://github.com/influxdata/telegraf/issues/8821))
- Validate the response from InfluxDB after writing/creating a database to avoid json parsing panics/errors ([#8775](https://github.com/influxdata/telegraf/issues/8775))
- Expose v4/v6-only connection-schemes through GosnmpWrapper ([#8804](https://github.com/influxdata/telegraf/issues/8804))
- adds missing & to flush_jitter output ref ([#8838](https://github.com/influxdata/telegraf/issues/8838))
- Sort and timeout is deadline ([#8839](https://github.com/influxdata/telegraf/issues/8839))
- Update README for inputs.ping with correct cmd for native ping on Linux ([#8787](https://github.com/influxdata/telegraf/issues/8787))
- Update go-ping to latest version ([#8771](https://github.com/influxdata/telegraf/issues/8771))


Thank you for your contributions!





<a name="v1.17.2"></a>
## [v1.17.2] - 2021-01-28

- Telegraf v1.17.2
- Update changelog
- Set interface for native ([#8770](https://github.com/influxdata/telegraf/issues/8770))
- Resolve regression, re-add missing function ([#8764](https://github.com/influxdata/telegraf/issues/8764))


Thank you for your contributions!





<a name="v1.17.1"></a>
## [v1.17.1] - 2021-01-27

- Telegraf v1.17.1
- avoid issues with vendored dependencies when running make
- Update changelog
- [outputs.influxdb_v2] add exponential backoff, and respect client error responses ([#8662](https://github.com/influxdata/telegraf/issues/8662))
- add line about measurement being specified in docs ([#8734](https://github.com/influxdata/telegraf/issues/8734))
- Fix issue with elasticsearch output being really noisy about some errors ([#8748](https://github.com/influxdata/telegraf/issues/8748))
- Add geoip external project reference
- improve mntr regex to match user specific keys. ([#7533](https://github.com/influxdata/telegraf/issues/7533))
- Fix crash in lustre2 input plugin, when field name and value ([#7967](https://github.com/influxdata/telegraf/issues/7967))
- Update grok-library to v1.0.1 with dots and dash-patterns fixed. ([#8673](https://github.com/influxdata/telegraf/issues/8673))
- Use go-ping for "native" execution in Ping plugin ([#8679](https://github.com/influxdata/telegraf/issues/8679))
- fix x509 cert timeout issue ([#8741](https://github.com/influxdata/telegraf/issues/8741))
- Add setting to enable caching in ipmitool ([#8335](https://github.com/influxdata/telegraf/issues/8335))
- Bump github.com/nsqio/go-nsq from 1.0.7 to 1.0.8 ([#8714](https://github.com/influxdata/telegraf/issues/8714))
- Bump github.com/Shopify/sarama from 1.27.1 to 1.27.2 ([#8715](https://github.com/influxdata/telegraf/issues/8715))
- add kafka connect example to jolokia2 input ([#8709](https://github.com/influxdata/telegraf/issues/8709))
- Bump github.com/newrelic/newrelic-telemetry-sdk-go from 0.2.0 to 0.5.1 ([#8712](https://github.com/influxdata/telegraf/issues/8712))
- Add Event Log support for Windows ([#8616](https://github.com/influxdata/telegraf/issues/8616))
- update readme: prometheus remote write ([#8683](https://github.com/influxdata/telegraf/issues/8683))
- GNMI plugin should not take off the first character of field keys when no 'alias path' exists. ([#8659](https://github.com/influxdata/telegraf/issues/8659))
- Use the 'measurement' json field from the particle webhook as the measurment name, or if it's blank, use the 'name' field of the event's json. ([#8609](https://github.com/influxdata/telegraf/issues/8609))
- Procstat input plugin should use the same timestamp in all metrics in the same Gather() cycle. ([#8658](https://github.com/influxdata/telegraf/issues/8658))
- update data formats output docs ([#8674](https://github.com/influxdata/telegraf/issues/8674))
- Add timestamp column support to postgresql_extensible ([#8602](https://github.com/influxdata/telegraf/issues/8602))
- Added ability to define skip values in csv parser ([#8627](https://github.com/influxdata/telegraf/issues/8627))
- fix some annoying tests due to ports in use
- Optimize SeriesGrouper & aggregators.merge ([#8391](https://github.com/influxdata/telegraf/issues/8391))
- Using mime-type in prometheus parser to handle protocol-buffer responses ([#8545](https://github.com/influxdata/telegraf/issues/8545))
- Input SNMP plugin - upgrade gosnmp library to version 1.29.0 ([#8588](https://github.com/influxdata/telegraf/issues/8588))
- Upgrade circle-ci config to v2.1 ([#8621](https://github.com/influxdata/telegraf/issues/8621))
- remove redundant reference to docs in data formats docs ([#8652](https://github.com/influxdata/telegraf/issues/8652))
- alphabetize external plugins list ([#8647](https://github.com/influxdata/telegraf/issues/8647))
- Open Hardware Monitor ([#8646](https://github.com/influxdata/telegraf/issues/8646))
- outputs/http: add option to control idle connection timeout ([#8055](https://github.com/influxdata/telegraf/issues/8055))
- update influxdb_v2 config documentation in main ([#8618](https://github.com/influxdata/telegraf/issues/8618))
- update intel powerstat readme ([#8600](https://github.com/influxdata/telegraf/issues/8600))
- common/tls: Allow specifying SNI hostnames ([#7897](https://github.com/influxdata/telegraf/issues/7897))
- Fix spelling and clarify docs ([#8164](https://github.com/influxdata/telegraf/issues/8164))
- fixed formatting (+1 squashed commit) ([#8541](https://github.com/influxdata/telegraf/issues/8541))
- Provide method to include core count when reporting cpu_usage in procstat input ([#6165](https://github.com/influxdata/telegraf/issues/6165))
- Add support for an inclusive job list in Jenkins plugin ([#8287](https://github.com/influxdata/telegraf/issues/8287))
- improve the error log message for snmp trap ([#8552](https://github.com/influxdata/telegraf/issues/8552))
- [http_listener_v2] Stop() succeeds even if fails to start ([#8502](https://github.com/influxdata/telegraf/issues/8502))
- Unify comments style in the CPU input ([#8605](https://github.com/influxdata/telegraf/issues/8605))
- Fix readme link for line protocol in influx parser ([#8610](https://github.com/influxdata/telegraf/issues/8610))
- Add hex_key parameter for IPMI input plugin connection ([#8524](https://github.com/influxdata/telegraf/issues/8524))
- Add more verbose errors to influxdb output ([#6061](https://github.com/influxdata/telegraf/issues/6061))


Thank you for your contributions!





<a name="v1.17.0"></a>
## [v1.17.0] - 2020-12-18

### Bug Fixes

- **exec:** fix typo in exec readme ([#8265](https://github.com/influxdata/telegraf/issues/8265))
- **ras:** update readme title ([#8266](https://github.com/influxdata/telegraf/issues/8266))

### Features

- add build number field to jenkins_job measurement ([#8038](https://github.com/influxdata/telegraf/issues/8038))


Thank you for your contributions!



Russ Savage ,
alespour ,### Reverts
- Update grok package to support for field names containing '-' and '.' ([#8276](https://github.com/influxdata/telegraf/issues/8276))
- disable flakey grok test for now
- fix to start Telegraf from Linux systemd.service




<a name="v1.16.3"></a>
## [v1.16.3] - 2020-12-01

- Telegraf v1.16.3
- Update changelog
- Log SubscribeResponse_Error message and code. [#8482](https://github.com/influxdata/telegraf/issues/8482) ([#8483](https://github.com/influxdata/telegraf/issues/8483))
- add log warning to starlark drop-fields example
- update godirwalk to v1.16.1 ([#7987](https://github.com/influxdata/telegraf/issues/7987))
- Starlark example dropbytype ([#8438](https://github.com/influxdata/telegraf/issues/8438))
- Fix typo in column name ([#8468](https://github.com/influxdata/telegraf/issues/8468))
- [php-fpm] Fix possible "index out of range" ([#8461](https://github.com/influxdata/telegraf/issues/8461))
- Update mdlayher/apcupsd dependency ([#8444](https://github.com/influxdata/telegraf/issues/8444))
- Show how to return a custom error with the Starlark processor ([#8439](https://github.com/influxdata/telegraf/issues/8439))
- keep field name as is for csv timestamp column ([#8440](https://github.com/influxdata/telegraf/issues/8440))
- Add DriverVersion and CUDA Version to output ([#8436](https://github.com/influxdata/telegraf/issues/8436))
- Show how to return several metrics with the Starlark processor ([#8423](https://github.com/influxdata/telegraf/issues/8423))
- Support logging in starlark ([#8408](https://github.com/influxdata/telegraf/issues/8408))
- add kinesis output to external plugins list ([#8315](https://github.com/influxdata/telegraf/issues/8315))
- [#8405](https://github.com/influxdata/telegraf/issues/8405) add non-retryable debug logging ([#8406](https://github.com/influxdata/telegraf/issues/8406))
- Wavefront output should distinguish between retryable and non-retryable errors ([#8404](https://github.com/influxdata/telegraf/issues/8404))
- Allow to catch errors that occur in the apply function ([#8401](https://github.com/influxdata/telegraf/issues/8401))


Thank you for your contributions!





<a name="v1.16.2"></a>
## [v1.16.2] - 2020-11-13

- Telegraf v1.16.2
- Update changelog
- Fix parsing of multiple files with different headers ([#6318](https://github.com/influxdata/telegraf/issues/6318)). ([#8400](https://github.com/influxdata/telegraf/issues/8400))
- proxmox: ignore QEMU templates and iron out a few bugs ([#8326](https://github.com/influxdata/telegraf/issues/8326))
- systemd_units: add --plain to command invocation ([#7990](https://github.com/influxdata/telegraf/issues/7990)) ([#7991](https://github.com/influxdata/telegraf/issues/7991))
- fix links in external plugins readme ([#8307](https://github.com/influxdata/telegraf/issues/8307))
- Fix minor typos in readmes ([#8370](https://github.com/influxdata/telegraf/issues/8370))
- Fix SMART plugin to recognize all devices from config ([#8374](https://github.com/influxdata/telegraf/issues/8374))
- Add OData-Version header to requests ([#8288](https://github.com/influxdata/telegraf/issues/8288))
- misc tests
- Prydin issue 8169 ([#8357](https://github.com/influxdata/telegraf/issues/8357))
- On-prem fix for [#8324](https://github.com/influxdata/telegraf/issues/8324) ([#8356](https://github.com/influxdata/telegraf/issues/8356))
- [output.wavefront] Introduced "immediate_flush" flag ([#8165](https://github.com/influxdata/telegraf/issues/8165))
- added support for bytes encoding ([#7938](https://github.com/influxdata/telegraf/issues/7938))
- Update jwt-go module to address CVE-2020-26160 ([#8337](https://github.com/influxdata/telegraf/issues/8337))
- fix plugins/input/ras test ([#8350](https://github.com/influxdata/telegraf/issues/8350))
- [#8328](https://github.com/influxdata/telegraf/issues/8328) Fixed a bug with the state map in Dynatrace Plugin ([#8329](https://github.com/influxdata/telegraf/issues/8329))


Thank you for your contributions!





<a name="v1.16.1"></a>
## [v1.16.1] - 2020-10-28

- Telegraf v1.16.1
- Update changelog
- SQL Server Azure PerfCounters Fix ([#8331](https://github.com/influxdata/telegraf/issues/8331))
- kafka sasl-mechanism auth support for SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI ([#8318](https://github.com/influxdata/telegraf/issues/8318))
- SQL Server - PerformanceCounters - removed synthetic counters ([#8325](https://github.com/influxdata/telegraf/issues/8325))
- SQL Server - server_properties added sql_version_desc ([#8324](https://github.com/influxdata/telegraf/issues/8324))
- Disable RAS input plugin on specific Linux architectures: mips64, mips64le, ppc64le, riscv64 ([#8317](https://github.com/influxdata/telegraf/issues/8317))
- processes: fix issue with stat no such file/dir ([#8309](https://github.com/influxdata/telegraf/issues/8309))
- fix issue with PDH_CALC_NEGATIVE_DENOMINATOR error ([#8308](https://github.com/influxdata/telegraf/issues/8308))
- RAS plugin - fix for too many open files handlers ([#8306](https://github.com/influxdata/telegraf/issues/8306))
- Get the build version from a static file


Thank you for your contributions!





<a name="v1.16.0"></a>
## [v1.16.0] - 2020-10-21

### Bug Fixes

- **readmes:** adding code block annotations ([#7963](https://github.com/influxdata/telegraf/issues/7963))
- **exec:** fix typo in exec readme ([#8265](https://github.com/influxdata/telegraf/issues/8265))
- **win_eventlog:** fixing config ([#8209](https://github.com/influxdata/telegraf/issues/8209))
- plugins/parsers/influx: avoid ParseError.Error panic ([#8177](https://github.com/influxdata/telegraf/issues/8177))
- **readmes:** standarize first line of readmes ([#7973](https://github.com/influxdata/telegraf/issues/7973))
- **puppet:** update broken link ([#7977](https://github.com/influxdata/telegraf/issues/7977))
- **ipmi:** update link in readme ([#7975](https://github.com/influxdata/telegraf/issues/7975))
- **readmes:** updates to internal and proxmox readmes ([#7982](https://github.com/influxdata/telegraf/issues/7982))

### Features

- add functionality to get values from redis commands ([#8196](https://github.com/influxdata/telegraf/issues/8196))


Thank you for your contributions!



Russ Savage ,
Roger Peppe ,
Yoofi Quansah ,### Reverts
- update influxdb v2 port




<a name="v1.15.4"></a>
## [v1.15.4] - 2020-10-21

- Telegraf v1.15.4
- Update changelog
- fix issue with loading processor config from execd ([#8274](https://github.com/influxdata/telegraf/issues/8274))
- fix panic on streaming processers using logging ([#8176](https://github.com/influxdata/telegraf/issues/8176))


Thank you for your contributions!





<a name="v1.15.3"></a>
## [v1.15.3] - 2020-09-11

### Bug Fixes

- **readmes:** standarize first line of readmes ([#7973](https://github.com/influxdata/telegraf/issues/7973))
- **puppet:** update broken link ([#7977](https://github.com/influxdata/telegraf/issues/7977))
- **ipmi:** update link in readme ([#7975](https://github.com/influxdata/telegraf/issues/7975))


Thank you for your contributions!



Russ Savage ,### Reverts
- fix cloudwatch tests




<a name="v1.15.2"></a>
## [v1.15.2] - 2020-07-31

- Telegraf 1.15.2
- Update changelog
- fixes issue with rpm /var/log/telegraf permissions ([#7909](https://github.com/influxdata/telegraf/issues/7909))
- Fix tail following on EOF ([#7927](https://github.com/influxdata/telegraf/issues/7927))


Thank you for your contributions!





<a name="v1.15.1"></a>
## [v1.15.1] - 2020-07-22

- Telegraf 1.15.1
- Update changelog
- Fix arch name in deb/rpm builds ([#7877](https://github.com/influxdata/telegraf/issues/7877))


Thank you for your contributions!





<a name="v1.15.0"></a>
## [v1.15.0] - 2020-07-22

- Telegraf 1.15.0
- Set 1.15.0 release date
- Add logic starlark example ([#7864](https://github.com/influxdata/telegraf/issues/7864))
- shim logger improvements ([#7865](https://github.com/influxdata/telegraf/issues/7865))
- Fix defaults processor readme typos ([#7873](https://github.com/influxdata/telegraf/issues/7873))
- Recv next message after send returns EOF ([#7872](https://github.com/influxdata/telegraf/issues/7872))
- fix issue with execd restart_delay being ignored ([#7867](https://github.com/influxdata/telegraf/issues/7867))
- clarify docs and add warning if execd is misconfigured ([#7866](https://github.com/influxdata/telegraf/issues/7866))
- fix bug with loading plugins in shim with no config ([#7816](https://github.com/influxdata/telegraf/issues/7816))
- Telegraf 1.15.0-rc4
- Fix suricata input docs ([#7856](https://github.com/influxdata/telegraf/issues/7856))
- ifname: avoid unpredictable conditions in getMap test ([#7848](https://github.com/influxdata/telegraf/issues/7848))
- Log after interval has elapsed; skip short intervals ([#7854](https://github.com/influxdata/telegraf/issues/7854))
- Initialize aggregation processors ([#7853](https://github.com/influxdata/telegraf/issues/7853))
- Update redfish docs with link ([#7846](https://github.com/influxdata/telegraf/issues/7846))
- Telegraf 1.15.0-rc3
- Update telegraf.conf
- ifname processor: expire old cached entries ([#7838](https://github.com/influxdata/telegraf/issues/7838))
- update go versions: 1.14.5, 1.13.13 ([#7837](https://github.com/influxdata/telegraf/issues/7837))
- Edit Starlark README ([#7832](https://github.com/influxdata/telegraf/issues/7832))
- Send metrics in FIFO order ([#7814](https://github.com/influxdata/telegraf/issues/7814))
- Set log output before starting plugin ([#7820](https://github.com/influxdata/telegraf/issues/7820))
- Fix darwin package build flags ([#7818](https://github.com/influxdata/telegraf/issues/7818))
- Close file to ensure it has been flushed ([#7819](https://github.com/influxdata/telegraf/issues/7819))
- Add minimum version for new plugins ([#7810](https://github.com/influxdata/telegraf/issues/7810))
- Fix markdown syntax ([#7806](https://github.com/influxdata/telegraf/issues/7806))
- Fix typo in 1.15 release notes ([#7804](https://github.com/influxdata/telegraf/issues/7804))
- Telegraf 1.15.0-rc2
- Fix tag package version
- Telegraf 1.15.0-rc1
- Update sample configuration
- Update readme and changelog
- Update changelog
- Add ifname processor plugin ([#7763](https://github.com/influxdata/telegraf/issues/7763))
- Traverse redfish api using resource links ([#7722](https://github.com/influxdata/telegraf/issues/7722))
- Fix test race in kafka_consumer ([#7797](https://github.com/influxdata/telegraf/issues/7797))
- Update changelog
- Support utf-16 in file and tail inputs ([#7792](https://github.com/influxdata/telegraf/issues/7792))
- Run all Go tests with flag -race ([#7783](https://github.com/influxdata/telegraf/issues/7783))
- Update changelog
- Add v3 metadata support to ecs input ([#7154](https://github.com/influxdata/telegraf/issues/7154))
- Fix inputs.execd readme links ([#7791](https://github.com/influxdata/telegraf/issues/7791))
- Fix data race in input plugin ping_windows
- streaming processors docs update ([#7786](https://github.com/influxdata/telegraf/issues/7786))
- switch mac tests to Go 1.14 ([#7784](https://github.com/influxdata/telegraf/issues/7784))
- Fix flakey processors.execd test
- Do not enable -race for GOARCH=386
- Run all Go tests with flag -race
- Fix data race in plugin output pubsub tests ([#7782](https://github.com/influxdata/telegraf/issues/7782))
- Shim refactor to support processors and output
- Fix data race in tail input tests ([#7780](https://github.com/influxdata/telegraf/issues/7780))
- Update CHANGELOG.md
- execd output ([#7761](https://github.com/influxdata/telegraf/issues/7761))
- Set user agent when requesting http config ([#7752](https://github.com/influxdata/telegraf/issues/7752))
- Update changelog
- Accept decimal point when parsing kibana uptime ([#7768](https://github.com/influxdata/telegraf/issues/7768))
- Update common/tls import path
- Update nginx_sts plugin readme
- Update changelog
- Add nginx_sts input plugin ([#7205](https://github.com/influxdata/telegraf/issues/7205))
- Update readme and changelog
- Rename cisco_telemetry_gnmi input to gnmi ([#7695](https://github.com/influxdata/telegraf/issues/7695))
- Update changelog
- Allow overriding the collection_jitter and precision per input ([#7762](https://github.com/influxdata/telegraf/issues/7762))
- Fix data race in phpfpm initializing http client ([#7738](https://github.com/influxdata/telegraf/issues/7738))
- Set 1.14.5 release date
- Update changelog
- Allow histograms with no buckets and summary without quantiles ([#7740](https://github.com/influxdata/telegraf/issues/7740))
- Fix typo in elasticsearch input docs ([#7764](https://github.com/influxdata/telegraf/issues/7764))
- Only set version ldflags on tags
- Update changelog
- Update release notes
- Allow any key usage type on x509 certificate ([#7760](https://github.com/influxdata/telegraf/issues/7760))
- Build packages in makefile ([#7759](https://github.com/influxdata/telegraf/issues/7759))
- Update changelog
- Update github.com/tidwall/gjson ([#7756](https://github.com/influxdata/telegraf/issues/7756))
- reverse dns lookup processor ([#7639](https://github.com/influxdata/telegraf/issues/7639))
- Execd processor ([#7640](https://github.com/influxdata/telegraf/issues/7640))
- remove streaming processors docs
- clean up tests
- address feedback
- Update changelog
- Return on toml parse errors instead of logging ([#7751](https://github.com/influxdata/telegraf/issues/7751))
- Update tls import path
- Export internal/tls package for use in execd plugins ([#7697](https://github.com/influxdata/telegraf/issues/7697))
- Update changelog
- Add laundry to mem input plugin on FreeBSD ([#7736](https://github.com/influxdata/telegraf/issues/7736))
- Fix data race in plugins/inputs/stackdriver/stackdriver_test.go ([#7744](https://github.com/influxdata/telegraf/issues/7744))
- Fix data race in plugins/inputs/suricata/suricata_test.go ([#7745](https://github.com/influxdata/telegraf/issues/7745))
- Fix data race in kafka_consumer_test.go ([#7737](https://github.com/influxdata/telegraf/issues/7737))
- Fix SNMP trap test race ([#7731](https://github.com/influxdata/telegraf/issues/7731))
- Update changelog
- Fix incorrect Azure SQL DB server properties ([#7715](https://github.com/influxdata/telegraf/issues/7715))
- fix race
- fix after rebase
- remove processors/execd/examples/count.go
- execd processor
- Fix license check
- Update readme/changelog
- Add starlark processor ([#7660](https://github.com/influxdata/telegraf/issues/7660))
- Update changelog
- Add missing nvme attributes to smart plugin ([#7575](https://github.com/influxdata/telegraf/issues/7575))
- Update changelog
- Add counter type to perfmon collector ([#7712](https://github.com/influxdata/telegraf/issues/7712))
- Update changelog
- Skip overs errors in the output of the sensors command ([#7718](https://github.com/influxdata/telegraf/issues/7718))
- Remove master/slave terminology from tests ([#7719](https://github.com/influxdata/telegraf/issues/7719))
- Update changelog
- Fix ping exit code handling on non-Linux ([#7658](https://github.com/influxdata/telegraf/issues/7658))
- Update changelog and redfish docs
- Add redfish input plugin ([#7082](https://github.com/influxdata/telegraf/issues/7082))
- Update changelog
- Add ability to add selectors as tags in kube_inventory ([#7267](https://github.com/influxdata/telegraf/issues/7267))
- Document that string fields do not produce prometheus metrics ([#7644](https://github.com/influxdata/telegraf/issues/7644))
- Remove trailing backslash management in sqlserver input ([#7700](https://github.com/influxdata/telegraf/issues/7700))
- Link to GJSON playground in json parser documentation ([#7698](https://github.com/influxdata/telegraf/issues/7698))
- Add 'batch' to mqtt output optional parameters ([#7690](https://github.com/influxdata/telegraf/issues/7690))
- Fail check-deps when differences are found ([#7694](https://github.com/influxdata/telegraf/issues/7694))
- Add state and readiness to kube_inventory pod metrics ([#7691](https://github.com/influxdata/telegraf/issues/7691))
- update CHANGELOG.md
- procstat performance enhancement ([#7686](https://github.com/influxdata/telegraf/issues/7686))
- Mark unused agent options as deprecated
- Fix processor initialization ([#7693](https://github.com/influxdata/telegraf/issues/7693))
- Update gNMI plugin readme ([#7685](https://github.com/influxdata/telegraf/issues/7685))
- Update changelog
- Remove trailing backslash from tag keys/values ([#7652](https://github.com/influxdata/telegraf/issues/7652))
- Update changelog
- Improve sqlserver input compatibility with older server versions ([#7495](https://github.com/influxdata/telegraf/issues/7495))
- Fix race issue in tick_test.go ([#7663](https://github.com/influxdata/telegraf/issues/7663))
- Flaky shim test ([#7656](https://github.com/influxdata/telegraf/issues/7656))
- link to glob pattern docs ([#7657](https://github.com/influxdata/telegraf/issues/7657))
- Set 1.14.4 release date
- Update changelog
- Update changelog
- Add ability to collect response body as field with http_response ([#7596](https://github.com/influxdata/telegraf/issues/7596))
- Update changelog
- Add timezone configuration to csv data format ([#7619](https://github.com/influxdata/telegraf/issues/7619))
- Change rpm dist packaging type for arm64 to aarch64 ([#7645](https://github.com/influxdata/telegraf/issues/7645))
- Update changelog
- Update changelog
- Update to github.com/shirou/gopsutil v2.20.5 ([#7641](https://github.com/influxdata/telegraf/issues/7641))
- Update changelog
- Fix source field for icinga2 plugin ([#7651](https://github.com/influxdata/telegraf/issues/7651))
- Update changelog
- Add video codec stats to nvidia-smi ([#7646](https://github.com/influxdata/telegraf/issues/7646))
- Update CHANGELOG.md
- fix issue with stream parser blocking when data is in buffer ([#7631](https://github.com/influxdata/telegraf/issues/7631))
- add support for streaming processors ([#7634](https://github.com/influxdata/telegraf/issues/7634))
- Update changelog
- Add tags to snmp_trap input for context name and engine ID ([#7633](https://github.com/influxdata/telegraf/issues/7633))
- Clarify use of multiple mqtt broker servers
- Add SNMPv3 trap support to snmp_trap input plugin ([#7294](https://github.com/influxdata/telegraf/issues/7294))
- Add support for Solus distribution to maintainer scripts ([#7585](https://github.com/influxdata/telegraf/issues/7585))
- Fix typo in queue depth example of diskio plugin ([#7613](https://github.com/influxdata/telegraf/issues/7613))
- Add support for env variables to shim config ([#7603](https://github.com/influxdata/telegraf/issues/7603))
- Update changelog
- Add support for once mode; run processors and aggregators during test ([#7474](https://github.com/influxdata/telegraf/issues/7474))
- Update AGGREGATORS_AND_PROCESSORS.md ([#7599](https://github.com/influxdata/telegraf/issues/7599))
- Add github.com/inabagumi/twitter-telegraf-plugin to list of external plugins
- Fix segmentation violation on connection failed ([#7593](https://github.com/influxdata/telegraf/issues/7593))
- Update changelog
- Add processor to look up service name by port ([#7540](https://github.com/influxdata/telegraf/issues/7540))
- make sure parse error includes offending text ([#7561](https://github.com/influxdata/telegraf/issues/7561))
- Update docs for newrelic output
- Add newrelic output plugin  ([#7019](https://github.com/influxdata/telegraf/issues/7019))
- Update changelog
- Allow collection of HTTP Headers in http_response input ([#7405](https://github.com/influxdata/telegraf/issues/7405))
- Update changelog
- Update to Go 1.14.3 with testing using 1.13.11 ([#7564](https://github.com/influxdata/telegraf/issues/7564))
- Update changelog
- Exclude csv_timestamp_column and csv_measurement_column from fields ([#7572](https://github.com/influxdata/telegraf/issues/7572))
- fix go version check ([#7562](https://github.com/influxdata/telegraf/issues/7562))
- Update changelog
- Fix the typo in `gcc_pu_fraction` to `gc_cpu_fraction` ([#7573](https://github.com/influxdata/telegraf/issues/7573))
- Update changelog
- Fix numeric to bool conversion in converter ([#7579](https://github.com/influxdata/telegraf/issues/7579))
- Add defaults processor to readme/changelog
- Add defaults processor to set default field values ([#7370](https://github.com/influxdata/telegraf/issues/7370))
- Update changelog
- Update changelog
- Add option to disable mongodb cluster status ([#7515](https://github.com/influxdata/telegraf/issues/7515))
- Update changelog
- Fix typos in sqlserver input ([#7524](https://github.com/influxdata/telegraf/issues/7524))
- Use updated clock package to resolve test failures ([#7516](https://github.com/influxdata/telegraf/issues/7516))
- fix randomly failing CI test ([#7514](https://github.com/influxdata/telegraf/issues/7514))
- Update changelog
- Add cluster state integer to mongodb input ([#7489](https://github.com/influxdata/telegraf/issues/7489))
- Update changelog
- Add configurable separator graphite serializer and output ([#7545](https://github.com/influxdata/telegraf/issues/7545))
- Update changelog
- Fix instance name resolution in performance counter query ([#7526](https://github.com/influxdata/telegraf/issues/7526))
- Set 1.14.3 release date
- Update changelog
- Close HTTP2 connections on timeout in influxdb outputs ([#7517](https://github.com/influxdata/telegraf/issues/7517))
- Fix negative value parsing in impi_sensor input ([#7541](https://github.com/influxdata/telegraf/issues/7541))
- Fix assorted spelling mistakes ([#7507](https://github.com/influxdata/telegraf/issues/7507))
- Fix documentation of percent_packet_loss field ([#7510](https://github.com/influxdata/telegraf/issues/7510))
- Update docs for execd plugins ([#7465](https://github.com/influxdata/telegraf/issues/7465))
- Update procstat pid_tag documentation
- Fix spelling errors in comments and documentation ([#7492](https://github.com/influxdata/telegraf/issues/7492))
- Update changelog
- Add truncate_tags setting to wavefront output ([#7503](https://github.com/influxdata/telegraf/issues/7503))
- Update changelog
- Add authentication support to the http_response input plugin ([#7491](https://github.com/influxdata/telegraf/issues/7491))
- Update changelog
- Handle multiple metrics with the same timestamp in dedup processor ([#7439](https://github.com/influxdata/telegraf/issues/7439))
- Update changelog
- Add additional fields to mongodb input ([#7321](https://github.com/influxdata/telegraf/issues/7321))
- Update changelog
- Add integer support to enum processor ([#7483](https://github.com/influxdata/telegraf/issues/7483))
- Fix typo in Windows service description ([#7486](https://github.com/influxdata/telegraf/issues/7486))
- Update changelog
- Add field creation to date processor and integer unix time support ([#7464](https://github.com/influxdata/telegraf/issues/7464))
- Update changelog
- Add cpu query to sqlserver input ([#7359](https://github.com/influxdata/telegraf/issues/7359))
- Update changelog
- Rework plugin tickers to prevent drift and spread write ticks ([#7390](https://github.com/influxdata/telegraf/issues/7390))
- Update changelog
- Update datadog output documentation ([#7467](https://github.com/influxdata/telegraf/issues/7467))
- Use docker log timestamp as metric time ([#7434](https://github.com/influxdata/telegraf/issues/7434))
- fix issue with execd-multiline influx line protocol ([#7463](https://github.com/influxdata/telegraf/issues/7463))
- Add information about HEC JSON format limitations and workaround ([#7459](https://github.com/influxdata/telegraf/issues/7459))
- Rename measurement to sqlserver_volume_space ([#7457](https://github.com/influxdata/telegraf/issues/7457))
- shim improvements for docs, clean quit, and slow readers ([#7452](https://github.com/influxdata/telegraf/issues/7452))
- Update changelog
- Fix gzip support in socket_listener with tcp sockets ([#7446](https://github.com/influxdata/telegraf/issues/7446))
- Update changelog
- Remove debug fields from spunkmetric serializer ([#7455](https://github.com/influxdata/telegraf/issues/7455))
- Fix filepath processor link in changelog ([#7454](https://github.com/influxdata/telegraf/issues/7454))
- Support Go execd plugins with shim ([#7283](https://github.com/influxdata/telegraf/issues/7283))
- Update changelog
- Add filepath processor plugin ([#7418](https://github.com/influxdata/telegraf/issues/7418))
- Add ContentEncoder to socket_writer for datagram sockets ([#7417](https://github.com/influxdata/telegraf/issues/7417))
- Sflow rework ([#7253](https://github.com/influxdata/telegraf/issues/7253))
- Update changelog
- Use same timestamp for all objects in arrays in the json parser ([#7412](https://github.com/influxdata/telegraf/issues/7412))
- Set 1.14.2 release date
- Update changelog
- Allow CR and FF inside of string fields and fix parser panic ([#7427](https://github.com/influxdata/telegraf/issues/7427))
- Fix typo in name of gc_cpu_fraction field ([#7425](https://github.com/influxdata/telegraf/issues/7425))
- Run create database query once per database ([#7333](https://github.com/influxdata/telegraf/issues/7333))
- Ignore fields with NaN or Inf floats in the JSON serializer ([#7426](https://github.com/influxdata/telegraf/issues/7426))
- Fix interfaces with pointers ([#7411](https://github.com/influxdata/telegraf/issues/7411))
- Document distinction between file and tail inputs ([#7353](https://github.com/influxdata/telegraf/issues/7353))
- Update changelog
- Fix shard indices reporting in elasticsearch input ([#7332](https://github.com/influxdata/telegraf/issues/7332))
- Update changelog
- Fix string to int64 conversion for SNMP input ([#7407](https://github.com/influxdata/telegraf/issues/7407))
- Update nvidia-smi README for Windows users ([#7399](https://github.com/influxdata/telegraf/issues/7399))
- Update changelog
- Extract target as a tag for each rule in iptables input ([#7391](https://github.com/influxdata/telegraf/issues/7391))
- Update changelog
- Fix dimension limit on azure_monitor output ([#7336](https://github.com/influxdata/telegraf/issues/7336))
- Update changelog
- Use new higher per request limit for cloudwatch GetMetricData ([#7335](https://github.com/influxdata/telegraf/issues/7335))
- Update changelog
- Add support for MDS and RGW sockets to ceph input ([#6915](https://github.com/influxdata/telegraf/issues/6915))
- Update changelog
- Add option to save retention policy as tag in influxdb_listener ([#7356](https://github.com/influxdata/telegraf/issues/7356))
- Update changelog
- Trim instance tag in the sqlserver performance counters query ([#7351](https://github.com/influxdata/telegraf/issues/7351))
- Update changelog
- Fix vSphere 6.7 missing data issue ([#7233](https://github.com/influxdata/telegraf/issues/7233))
- Update modbus readme
- Update changelog
- Add retry when slave is busy to modbus input ([#7271](https://github.com/influxdata/telegraf/issues/7271))
- fix issue with closing flush signal channel ([#7384](https://github.com/influxdata/telegraf/issues/7384))
- Use the product token for the user agent in more locations ([#7378](https://github.com/influxdata/telegraf/issues/7378))
- Update changelog
- Update changelog
- Update github.com/aws/aws-sdk-go ([#7373](https://github.com/influxdata/telegraf/issues/7373))
- add support for SIGUSR1 to trigger flush ([#7366](https://github.com/influxdata/telegraf/issues/7366))
- add another grok example for custom timestamps ([#7367](https://github.com/influxdata/telegraf/issues/7367))
- Fibaro input: for battery operated devices, add battery level scraping ([#7319](https://github.com/influxdata/telegraf/issues/7319))
- Deprecate logparser input and recommend tail input as replacement ([#7352](https://github.com/influxdata/telegraf/issues/7352))
- Adjust heading level in the filtering examples to allow linking
- Set 1.14.1 release date
- Update changelog
- Add reading bearer token from a file to http input ([#7304](https://github.com/influxdata/telegraf/issues/7304))
- Update changelog
- Fix exclude database and retention policy tags is shared ([#7323](https://github.com/influxdata/telegraf/issues/7323))
- Fix status path when using globs in phpfpm ([#7324](https://github.com/influxdata/telegraf/issues/7324))
- Regenerate telegraf.conf
- Fix error in docs about exclude_retention_policy_tag ([#7311](https://github.com/influxdata/telegraf/issues/7311))
- Update changelog
- Update changelog
- Fix Name field in template processor ([#7258](https://github.com/influxdata/telegraf/issues/7258))
- Deploy telegraf configuration as a "non config" file ([#7250](https://github.com/influxdata/telegraf/issues/7250))
- Fix export timestamp not working for prometheus on v2 ([#7289](https://github.com/influxdata/telegraf/issues/7289))
- Sql Server - Disk Space Measurement ([#7214](https://github.com/influxdata/telegraf/issues/7214))
- Add series cardinality warning to sflow readme ([#7285](https://github.com/influxdata/telegraf/issues/7285))
- Improve documentation for the Metric interface ([#7256](https://github.com/influxdata/telegraf/issues/7256))
- Update permission docs on postfix input ([#7255](https://github.com/influxdata/telegraf/issues/7255))
- Document kapacitor_alert and kapacitor_cluster measurements ([#7278](https://github.com/influxdata/telegraf/issues/7278))
- Update changelog
- Add OPTION RECOMPILE for perf reasons due to temp table ([#7242](https://github.com/influxdata/telegraf/issues/7242))
- Update changelog
- Support multiple templates for graphite serializers ([#7136](https://github.com/influxdata/telegraf/issues/7136))
- Update changelog
- Add possibility to specify measurement per register ([#7231](https://github.com/influxdata/telegraf/issues/7231))
- Add limit to number of undelivered lines to read ahead in tail ([#7210](https://github.com/influxdata/telegraf/issues/7210))
- Add docs for how to handle errors in check-deps script ([#7243](https://github.com/influxdata/telegraf/issues/7243))
- Update changelog
- Add support for 64-bit integer types to modbus input ([#7225](https://github.com/influxdata/telegraf/issues/7225))
- Set 1.14.0 release date
- Update changelog
- Apply ping deadline to dns lookup ([#7140](https://github.com/influxdata/telegraf/issues/7140))
- Update changelog
- Add ability to specify HTTP Headers in http_listener_v2 which will added as tags ([#7223](https://github.com/influxdata/telegraf/issues/7223))
- Fix 'nil' file created by Makefile on Windows ([#7224](https://github.com/influxdata/telegraf/issues/7224))
- Update changelog
- Add additional concurrent transaction information ([#7193](https://github.com/influxdata/telegraf/issues/7193))
- Add commands stats to mongodb input plugin ([#6905](https://github.com/influxdata/telegraf/issues/6905))
- Update changelog
- Fix url encoding of job names in jenkins input plugin ([#7211](https://github.com/influxdata/telegraf/issues/7211))
- Update next_version on master to 1.15.0
- Update etc/telegraf.conf
- Fix datastore_include option in vsphere input readme
- Update github.com/prometheus/client_golang to latest ([#7200](https://github.com/influxdata/telegraf/issues/7200))
- Update etc/telegraf.conf
- Update google.cloud.go to latest ([#7199](https://github.com/influxdata/telegraf/issues/7199))


Thank you for your contributions!





<a name="v1.14.5"></a>
## [v1.14.5] - 2020-06-30

- Telegraf 1.14.5
- Set 1.14.5 release date
- Update changelog
- Allow histograms with no buckets and summary without quantiles ([#7740](https://github.com/influxdata/telegraf/issues/7740))
- Update changelog
- Allow any key usage type on x509 certificate ([#7760](https://github.com/influxdata/telegraf/issues/7760))
- Update changelog
- Update github.com/tidwall/gjson ([#7756](https://github.com/influxdata/telegraf/issues/7756))
- Update changelog
- Return on toml parse errors instead of logging ([#7751](https://github.com/influxdata/telegraf/issues/7751))
- Update changelog
- Skip overs errors in the output of the sensors command ([#7718](https://github.com/influxdata/telegraf/issues/7718))
- Update changelog
- Fix ping exit code handling on non-Linux ([#7658](https://github.com/influxdata/telegraf/issues/7658))
- update CHANGELOG.md
- procstat performance enhancement ([#7686](https://github.com/influxdata/telegraf/issues/7686))


Thank you for your contributions!





<a name="v1.14.4"></a>
## [v1.14.4] - 2020-06-09

- Telegraf 1.14.4
- Set 1.14.4 release date
- Update changelog
- Update CHANGELOG.md
- Update CHANGELOG.md
- fix issue with stream parser blocking when data is in buffer ([#7631](https://github.com/influxdata/telegraf/issues/7631))
- Update changelog
- Fix the typo in `gcc_pu_fraction` to `gc_cpu_fraction` ([#7573](https://github.com/influxdata/telegraf/issues/7573))
- Update changelog
- Fix numeric to bool conversion in converter ([#7579](https://github.com/influxdata/telegraf/issues/7579))
- Update changelog
- Fix instance name resolution in performance counter query ([#7526](https://github.com/influxdata/telegraf/issues/7526))


Thank you for your contributions!





<a name="v1.14.3"></a>
## [v1.14.3] - 2020-05-19

- Telegraf 1.14.3
- Set 1.14.3 release date
- Update changelog
- Close HTTP2 connections on timeout in influxdb outputs ([#7517](https://github.com/influxdata/telegraf/issues/7517))
- Fix negative value parsing in impi_sensor input ([#7541](https://github.com/influxdata/telegraf/issues/7541))
- Update changelog
- Handle multiple metrics with the same timestamp in dedup processor ([#7439](https://github.com/influxdata/telegraf/issues/7439))
- Update changelog
- Use same timestamp for all objects in arrays in the json parser ([#7412](https://github.com/influxdata/telegraf/issues/7412))


Thank you for your contributions!





<a name="v1.14.2"></a>
## [v1.14.2] - 2020-04-28

- Telegraf 1.14.2
- Set 1.14.2 release date
- Update changelog
- Allow CR and FF inside of string fields and fix parser panic ([#7427](https://github.com/influxdata/telegraf/issues/7427))
- Fix typo in name of gc_cpu_fraction field ([#7425](https://github.com/influxdata/telegraf/issues/7425))
- Run create database query once per database ([#7333](https://github.com/influxdata/telegraf/issues/7333))
- Ignore fields with NaN or Inf floats in the JSON serializer ([#7426](https://github.com/influxdata/telegraf/issues/7426))
- Update changelog
- Fix shard indices reporting in elasticsearch input ([#7332](https://github.com/influxdata/telegraf/issues/7332))
- Update changelog
- Fix string to int64 conversion for SNMP input ([#7407](https://github.com/influxdata/telegraf/issues/7407))
- Update nvidia-smi README for Windows users ([#7399](https://github.com/influxdata/telegraf/issues/7399))
- Update changelog
- Fix dimension limit on azure_monitor output ([#7336](https://github.com/influxdata/telegraf/issues/7336))
- Update changelog
- Use new higher per request limit for cloudwatch GetMetricData ([#7335](https://github.com/influxdata/telegraf/issues/7335))
- Update changelog
- Trim instance tag in the sqlserver performance counters query ([#7351](https://github.com/influxdata/telegraf/issues/7351))


Thank you for your contributions!





<a name="v1.14.1"></a>
## [v1.14.1] - 2020-04-14

- Telegraf 1.14.1
- Set 1.14.1 release date
- Update changelog
- Fix exclude database and retention policy tags is shared ([#7323](https://github.com/influxdata/telegraf/issues/7323))
- Fix status path when using globs in phpfpm ([#7324](https://github.com/influxdata/telegraf/issues/7324))
- Regenerate telegraf.conf
- Fix error in docs about exclude_retention_policy_tag ([#7311](https://github.com/influxdata/telegraf/issues/7311))
- Update changelog
- Fix export timestamp not working for prometheus on v2 ([#7289](https://github.com/influxdata/telegraf/issues/7289))
- Update changelog
- Fix Name field in template processor ([#7258](https://github.com/influxdata/telegraf/issues/7258))
- Add series cardinality warning to sflow readme ([#7285](https://github.com/influxdata/telegraf/issues/7285))
- Document kapacitor_alert and kapacitor_cluster measurements ([#7278](https://github.com/influxdata/telegraf/issues/7278))
- Update changelog
- Add OPTION RECOMPILE for perf reasons due to temp table ([#7242](https://github.com/influxdata/telegraf/issues/7242))


Thank you for your contributions!





<a name="v1.14.0"></a>
## [v1.14.0] - 2020-03-26

### Bug Fixes

- **prometheus:** Add support for bearer token to prometheus input plugin
- **Godeps:** Added github.com/opencontainers/runc
- **indent:** For configuration sample
- **import:** Json parser lives outside internal
- Last link on README
- **sample:** Made TOML parser happy again
- **config:** Made sample config consistent.
- **kubernetes:** Only initialize RoundTripper once ([#1951](https://github.com/influxdata/telegraf/issues/1951))
- **vet:** Range var used by goroutine
- **mesos:** TOML annotation

### Features

- **nsq_consumer:** Add input plugin
- **kubernetes:** Add kubernetes input plugin closes [#1774](https://github.com/influxdata/telegraf/issues/1774)
- **whitelist:** Converted black to whitelist
- **timeout:** Use timeout setting


Thank you for your contributions!



Jonathan Chauncey ,
Sergio Jimenez ,### Reverts
- Add CLA check GitHub action ([#6479](https://github.com/influxdata/telegraf/issues/6479))
- Update aerospike-client-go version to latest release ([#4128](https://github.com/influxdata/telegraf/issues/4128))
- Add tengine input plugin ([#4160](https://github.com/influxdata/telegraf/issues/4160))
- Undo Revert "Revert changes since 9b0af4478"
- New Particle Plugin
- bug fixes and refactoring
- Update README.md
- Updated README.md
- Small fixes
- Updated Test JSON
- Updated Test JSON
- New Particle.io Plugin for Telegraf
- Moving cgroup path name to field from tag to reduce cardinality ([#1457](https://github.com/influxdata/telegraf/issues/1457))
- add pgbouncer plugin
- Revert graylog output
- exec plugin: allow using glob pattern in command list

### Pull Requests
- Merge pull request [#2024](https://github.com/influxdata/telegraf/issues/2024) from influxdata/cs2023-single-quote-duration
- Merge pull request [#1847](https://github.com/influxdata/telegraf/issues/1847) from jchauncey/kubernetes-plugin
- Merge pull request [#1768](https://github.com/influxdata/telegraf/issues/1768) from influxdata/dgn-speedup-statsd-parser
- Merge pull request [#1766](https://github.com/influxdata/telegraf/issues/1766) from influxdata/dgn-statsd-parsing-benchmarks
- Merge pull request [#1426](https://github.com/influxdata/telegraf/issues/1426) from influxdata/metrics-panic
- Merge pull request [#1157](https://github.com/influxdata/telegraf/issues/1157) from influxdata/ross-build-updates
- Merge pull request [#896](https://github.com/influxdata/telegraf/issues/896) from jipperinbham/graphite-tag-sanitizer
- Merge pull request [#891](https://github.com/influxdata/telegraf/issues/891) from jipperinbham/librato-serialize-fix
- Merge pull request [#886](https://github.com/influxdata/telegraf/issues/886) from entertainyou/typo
- Merge pull request [#882](https://github.com/influxdata/telegraf/issues/882) from VasuBalakrishnan/master
- Merge pull request [#883](https://github.com/influxdata/telegraf/issues/883) from ljagiello/minor-changelog-fix
- Merge pull request [#875](https://github.com/influxdata/telegraf/issues/875) from Onefootball/feature/link-freebsd-package
- Merge pull request [#858](https://github.com/influxdata/telegraf/issues/858) from LordFPL/patch-1
- Merge pull request [#790](https://github.com/influxdata/telegraf/issues/790) from arthtux/master
- Merge pull request [#764](https://github.com/influxdata/telegraf/issues/764) from arthtux/master
- Merge pull request [#673](https://github.com/influxdata/telegraf/issues/673) from miketonks/f-docker-percentages




<a name="v0.10.1"></a>
## [v0.10.1] - 2016-01-27

- Release 0.10.1
- Fix SNMP unit tests on OSX, improve tag config doc
- Update dependency hashes
- Update changelog
- Additional request header parameters for httpjson plugin
- Update changelog and readme, and small tweaks to github_webhooks
- Merge branch 'ghWebhooks'
- Add sqlserver input plugin
- Fixup some disk usage reporting, make it reflect df
- Fix some inputs panic will lead to the telegraf exit
- Fix naming issue
- Remove internal dependancy
- Added Amazon Linux logic to post-installation script.
- Backporting fixes from the influxdb build script, along with a few improvements: - Added iteration to tar/zip output name (in case of pre-releases) - Switched 32-bit signifier to i386 from 386 - Tweaked upload settings to allow for folder paths in bucket names
- Insert documentation into sample-config on JSON parsing
- RabbitMQ plugin - extra fields:
- Add README.md
- Add snmp input plugin
- Change configuration package to influxdata/config
- add 'gdm restore' to adding a dependency instructions
- Address PR comments and merge conflicts
- Fix merge conflict in all.go
- Change github.com/influxdata to github.com/influxdb where necessary
- Change github.com/influxdata to github.com/influxdata
- Add tests
- Kinesis output shouldn't return an error for no reason
- Implement a per-output fixed size metric buffer
- Gather elasticsearch nodes in goroutines, handle errors
- Changelog update
- Refactor the docker plugin, use go-dockerclient throughout
- Add Cloudwatch output
- Remove go get ./... from the Makefile
- statsd: If parsing a value to int fails, try to float and cast to int
- Push ghwebhooks branch
- Update contributing document
- Update changelog
- Fix issue 524
- Change start implementation
- Filter mount points before stats are collected
- First commit for ghwebhooks service plugin
- Include CPU usage percent with procstat data
- Collection interval random jittering
- Changelog update
- Change default statsd packet size to 1500, make configurable
- Replace plugins by inputs in some strings
- Update Godeps file
- kafka: Add support for using TLS authentication for the kafka output
- Add phusion Passenger plugin
- Add SIGHUP support to reload telegraf config
- changelog bugfix update
- phpfpm plugin: enhance socket gathering and config
- core: print error on output connect fail
- output amqp: Add ssl support
- Tweak config messages for graphite. Update changelog and readme
- Add Graphite output
- Make NSQ plugin compatible with version 0.10.0
- NSQ Plugin
- Add option to disable statsd name conversion
- Update procstat doc
- Update README.md
- Merge pull request [#533](https://github.com/influxdata/telegraf/issues/533) from influxdata/fix-interval-option-v0.10
- interval options should have string value
- Removing old package script, trim Makefile
- Add a quiet mode to telegraf
- Only compile the sensors plugin if the 'sensors' tag is set
- Tweak changelog for sensors plugin, and add a non-linux build file
- Added infor to readme and changelog
- Change build configuration to linux only
- Fixed an unused variable
- Added initial support for gosensors module
- Add response time to httpjson plugin
- Switched to /etc/debian_version for Debian/Ubuntu distribution recognition in post-install.
- Update Godeps and fix changelog 2014->2016
- Note on where to look for plugin information
- Add an interface:"all" tag to the net protocol counters
- Align exec documentation with v0.10 updates
- build.py: Make build script work on both Python2.x and Python3.x
- Ping input doesn't return response time metric when timeout
- internal: FlattenJSON, flatten arrays as well
- Add 0.10.0 blog post link to README
- Fix Telegraf s3 upload and readme links


Thank you for your contributions!


### Pull Requests
- Merge pull request [#533](https://github.com/influxdata/telegraf/issues/533) from influxdata/fix-interval-option-v0.10




<a name="v0.10.0"></a>
## [v0.10.0] - 2016-01-11

- Change 0.3.0 -> 0.10.0
- Update changelog and readme for package updates
- Merge pull request [#497](https://github.com/influxdata/telegraf/issues/497) from influxdata/rm-package-updates
- Removed data directory entries, since Telegraf doesn't need them.
- Added a `build.py` script for compiling and packaging. Added post and pre install scripts to handle installation and upgrades in a cleaner way. Minor fixes to the init script and service unit file.
- 0.3.0: update README and documentation
- add backwards-compatability for 'plugins', remove [inputs] and [outputs] headers
- 0.3.0: update README and documentation
- renaming plugins -> inputs
- 0.3.0 documentation changes and improvements
- Update Makefile and Godeps and various fixups
- 0.3.0 unit tests: agent and prometheus
- 0.3.0 unit tests: internal
- 0.3.0 unit tests: amon, datadog, librato
- 0.3.0 unit tests: influxdb
- 0.3.0 unit tests: rethinkdb, twemproxy, zfs
- 0.3.0 unit tests: statsd, trig, zookeeper
- 0.3.0 unit tests: rabbitmq, redis
- 0.3.0 unit tests: procstat, prometheus, puppetagent
- 0.3.0 unit tests: mysql, nginx, phpfpm, ping, postgres
- 0.3.0 unit tests: mailchimp, memcached, mongodb
- 0.3.0 unit tests: jolokia, kafka_consumer, leofs, lustre2
- 0.3.0 unit tests: exec, httpjson, and haproxy
- 0.3.0 unit tests: disque and elasticsearch
- 0.3.0 unit tests: aerospike, apache, bcache
- 0.3.0 unit tests: system plugins
- Fix httpjson panic for nil request body
- 0.3.0 Removing internal parallelism: twemproxy and rabbitmq
- 0.3.0 Removing internal parallelism: procstat
- 0.3.0 Removing internal parallelism: postgresql
- 0.3.0 Removing internal parallelism: httpjson and exec
- 0.3.0 outputs: riemann
- CHANGELOG update
- 0.3.0 outputs: opentsdb
- 0.3.0 output: librato
- 0.3.0 output: datadog and amon
- 0.3.0: mongodb and jolokia
- 0.3.0: postgresql and phpfpm
- 0.3.0 HAProxy rebase
- 0.3.0: rethinkdb
- 0.3.0: zookeeper and zfs
- backwards compatability for io->diskio change
- 0.3.0: trig and twemproxy
- 0.3.0 redis & rabbitmq
- 0.3.0: prometheus & puppetagent
- 0.3.0: procstat
- 0.3.0: ping, mysql, nginx
- 0.3.0: mailchimp & memcached
- 0.3.0: leofs & lustre2
- 0.3.0 httpjson
- 0.3.0: HAProxy
- Breakout JSON flattening into internal package, exec & elasticsearch aggregation
- Updating aerospike & apache plugins for 0.3.0
- Updating system plugins for 0.3.0
- fix too restrictive .gitignore
- Update circleci badge
- Fix typo in telegraf.conf
- Update 0.3.0 beta links in readme
- Links for the 0.3.0 beta version
- remove Name from influxdb unit test
- Remove 'Name' argument from influxdb plugin for 0.3.0 compatability
- Add influxdb plugin
- add additional stats that were already being collected
- close r.Body, remove network metrics, updated other sections as needed
- Do not rely on external server for amon unit tests
- Use gdm for dependency management
- Remove Godeps/ directory
- Go fmt kinesis output test file
- add amazon kinesis as an output plugin
- Separate pool tag and stat collection.
- Fix single dataset test.
- Add zfs pool stats collection.


Thank you for your contributions!


### Pull Requests
- Merge pull request [#497](https://github.com/influxdata/telegraf/issues/497) from influxdata/rm-package-updates




<a name="v0.2.4"></a>
## [v0.2.4] - 2015-12-08

- Telegraf 0.2.4 version bump
- Implement Glob matching for pass/drop filters
- Add support for pass/drop/tagpass/tagdrop for outputs
- Resolve gopsutil & unit test issues with net proto stats
- Add network protocol stats to the network plugin
- Convert uptime to float64 for backwards compatability.
- Remove  from test and test-short in Makefile
- Mailchimp report plugin
- Update gopsutil godep dependency. Dont use godep go build anymore
- cpu plugin: update LastStats before returning
- memcached plugin. Break out metric parsing into it's own func
- memcached plugin: support unix sockets
- Add optional auth credentials to Jolokia plugin
- io plugin, add an 'unknown' tag when the serial number can't be found
- redis_test.go with instantaneous input/output
- add instantaneous input/output to redis plugin.
- Adding all memcached stats that return a single value
- Create trig plugin
- Don't use panic-happy prometheus client With() function
- Make Prometheus output tests skipped in short mode.
- Update CHANGELOG and README for 0.2.3


Thank you for your contributions!





<a name="v0.2.3"></a>
## [v0.2.3] - 2015-11-30

- Parse statsd lines with multiple metric bits
- Update etc/telegraf.conf file
- Change aerospike plugin server tag to aerospike_host
- Put Agent Config into the config package
- Overhaul config <-> agent coupling. Put config in it's own package.
- Revert much of the newer config file parsing, fix tagdrop/tagpass
- Eliminate merging directory structures
- Change plugin config to be specified as a list
- cmd/telegraf: -configdirectory only includes files ending in .conf
- Add a comment indicating pattern uses pgrep -f
- Use pgrep with a pattern
- cmd/telegraf: -configdirectory only includes files ending in .conf
- GOPATH can have multiple : separated paths in it.
- Skip measurements with NaN fields
- Fix kafka plugin and rename to kafka_consumer
- Riemann output: remove some of the object referencing/dereferencing
- Godep: Add raidman riemann client
- Add riemann output
- Update README for 0.2.2


Thank you for your contributions!





<a name="v0.2.2"></a>
## [v0.2.2] - 2015-11-18

- Dont append to slices in mergeStruct
- Use 'CREATE DATABASE IF NOT EXISTS' syntax


Thank you for your contributions!





<a name="v0.2.1"></a>
## [v0.2.1] - 2015-11-16

- Updating CHANGELOG and README for version 0.2.1
- Update README, CHANGELOG, and unit tests with list output
- FreeBSD compatibility
- Allow users to specify outputs as lists
- CHANGELOG update
- MQTT output unit tests w/ docker container
- Apache plugin unit tests and README
- InfluxDB output: add tests and a README
- Twemproxy go fmt and bug fixups, CHANGELOG, README
- Add plugin for Twemproxy
- Update CHANGELOG with UDP output
- Godep update and dependency resolution
- Use the UDP client for writing to InfluxDB
- phpfpm: add socket fcgi support
- measurement name should have prefix before ShouldPass check
- Fix config file tab indentation
- Fix new error return of client.NewPoint
- Godep update: gopsutil
- Change duration -> internal and implement private gopsutil methods
- Godep update: influxdb
- Godep save: gopsutil
- Revert "redis: support IPv6 addresses with no port"
- redis: support IPv6 addresses with no port
- Amon output
- removed "panic" from zfs plugin
- add ZFS plugin
- Added parameters "Devices" and "SkipSerialNumber to DiskIO plugin.
- Added jolokia README.md
- Test for jolokia plugin
- Add fields value test methods
- Create a JolokiaClient. allowing to inject a stub implementation
- Fixed sampleconfig
- go fmt run over jolokia.go
- Use url.Parse to validate configuration params
- Added Tags as toml field
- Jolokia plugin first commit
- removed "panic" from bcache plugin
- updating Golang crypto
- Change HAProxy plugin tag from host to server
- Suggest running as telegraf user in test mode in README
- Improve the HTTP JSON plugin README with more examples.
- Mongodb should output 2 plugins in test mode
- Completely tab-indent the Makefile
- On a package upgrade, restart telegraf.
- Dont overwrite 'host' tag in redis plugin
- [rabbitmq plugin] Add support for per-queue metrics
- [amqp output] Add ability to specify influxdb database
- add elasticsearch README
- add ValidateTaggedFields func to testutil accumulator
- optinally gather cluster and index health stats
- Prometheus client test refactor
- Add prometheus_client service output module, update prometheus client
- update mongostat from github.com/mongodb/mongo-tools
- Run make in circle, don't build arm and 32-bit
- Execute "long" unit tests using docker containers
- Mongostat diff bug, less equal to less
- Update README & CHANGELOG with docker and NSQ changes
- fixing test for NoError
- use index 0 of server array for nsq test
- updated for new output Write function
- NSQ Output plugin
- Update CHANGELOG with version 0.2.0


Thank you for your contributions!


### Reverts
- redis: support IPv6 addresses with no port




<a name="v0.2.0"></a>
## [v0.2.0] - 2015-10-27

- Making sure telegraf.d directory is created by packages.
- Making the field name matching when merging respect the toml struct tag.
- Update README to version 0.2.0
- Fixup random interval jittering
- add librato output plugin, update datadog plugin to skip non-number metrics
- Change aerospike default config to localhost
- Add httpjson readme
- Rename Tags to TagKeys
- [Fix [#190](https://github.com/influxdata/telegraf/issues/190)] Add httpjson tags support
- add host to metric, replace '_' with '.'
- Use specific mysql version with docker
- Replace opentsb docker image with the official one
- Update kafka reamde; improve intergration tests
- Fix MySQL DSN -> tags parsing
- Support printing output with usage flag too
- Fix for tags in the config not being applied to the agent.
- Do not fail Connect() in influxdb output when db creation fails
- When MongoDB freezes or restarts, do not report negative diffs
- Fix output panic for -test flag
- Update CHANGELOG & README with aerospike plugin
- Add aerospike plugin support
- Update CHANGELOG with new flushing options
- Normalize collection interval to nearest interval
- Tests for LoadDirectory.
- Implementing LoadDirectory.
- Fixing old tests and adding new ones for new code.
- Moving the Duration wrapper to it's own package to break import loops.
- Adding testify/suite to godep.
- Moving away from passing around *ast.Tables.
- Combine BatchPoints with the same RoutingTag to one message in amqp output
- Add support for retrying output writes, using independent threads
- Clean up logging messages and add flusher startup delay
- Add periods to the end of sentences
- add bcache plugin
- Utilizing new client and overhauling Accumulator interface
- Godep update: influxdb
- InfluxDB does not accept uint64, so cast them down to int64
- added keyspace hitrate measurement
- added connections measurement with user tag
- fixed test to check actual value
- PuppetAgent Plugin
- Turn off GOGC for faster build time in CI
- Use Unix() int64 time for comparing timestamps in kafka consumer
- Fix ApplyTemplate change in graphite parser
- godep update: influxdb
- Run go fmt in CI
- Fix Go vet issue, test accumulator should be passed by reference with lock
- Add locking to test accumulator
- Fix typos
- Add phpfpm to readme
- Change config file indentation to 2 spaces
- Fix for init script for other procs with "telegraf"
- Statsd plugin, tags and timings
- wget and install go1.5.1 on machine
- Use graphite parser for templating, godep update to head
- Refactoring gauges to support floats, unit tests
- Statsd: unit tests for gauges, sets, counters
- Statsd listener plugin
- Add recently-added plugins to list
- Issue [#264](https://github.com/influxdata/telegraf/issues/264): Fixes for logrotate config file.
- remove zookeeper declaration
- added measurement prefix
- fixes based on comments
- Zookeeper plugin
- Update CHANGELOG with recent bugfixes
- Fix crash if login/password is incorrect in rabbitmq plugin. Closes [#260](https://github.com/influxdata/telegraf/issues/260)
- Add sample for exec plugin. Fixes [#245](https://github.com/influxdata/telegraf/issues/245)
- Add PHPFPM stat
- add UDP socket counts and rename to 'netstat'.
- add REAME about TCP Connection plugin.
- add NetConnections to the mockPS.
- add tcp connections stat plugin.
- telegraf-agent.toml: Fix example port and use complete examples for mysql plugin
- Merge pull request [#252](https://github.com/influxdata/telegraf/issues/252) from aristanetworks/master
- Dropped SkipInodeUsage option as "drop" achieves the same results. Fixed a bug in restricting Disk reporting to specific mountpoints Added tests for the Disk.Mountpoints option Fixed minor bug in usage of assert for the cpu tests where expected and actual values were swapped.
- Race condition fix: copy BatchPoints into goroutine
- godep update: gopsutil
- Merge remote-tracking branch 'upstream/master'
- Added Mountpoints and SkipInodeUsage options to the Disk plugin to control which mountpoint stats get reported for and to skip inode stats.
- procstat plugin, consolidate PID-getting
- Allow procstat plugin to handle multiple PIDs from pgrep
- Add pid tag to procstat plugin, dont exit on error, only log
- fix typo in sample config and README
- fix plugin registration name
- fix toml struct string
- add readme for procstat plugin
- godep update for procstat
- Monitor process by pidfile or exe name
- Godep update: gopsutil
- add tabs in the apache sampleConfig var
- godep update: gopsutil
- Fix godeps for MQTT output and remove hostname setting
- Change MQTT output topic format to split plugin name.
- update Godep.json
- Add MQTT output.
- Merge remote-tracking branch 'upstream/master'
- CHANGELOG feature updates
- Clean up additional logging and always print basic agent config
- Memory plugin: re-add cached and buffered to memory plugin
- Add more logging to telegraf
- Fix conditional test against useradd so it's compatible with Dash
- Merge remote-tracking branch 'upstream/master'
- Fix packages provides: now new version of package replaces the old one
- AMQP auto reconnect feature
- Fix printf format issue
- Adds command intervals to exec plugin
- Make nginx_test check port in nginx module tags
- Add port tag to nginx plugin
- Update CHANGELOG with ekini's changes and docker plugin
- Add timestamps to points in Kafka/AMQP outputs
- Update godep of go-dockerclient for Label access
- docker plugin: Add docker labels as tags in
- Only run the cpu plugin twice when using -test
- Make redis password config more clear.
- Remove duplicate opentsdb docker images
- Redis: include per-db keyspace info
- Redis plugin, add key metrics and simplify parsing
- Update changelog with info about filtering
- Updating README and CHANGELOG for 0.1.9
- Fixed memory reporting for Linux systems
- Fixed total memory reporting for Darwin systems. hw.memsize is reported as bytes instead of pages.


Thank you for your contributions!


### Pull Requests
- Merge pull request [#252](https://github.com/influxdata/telegraf/issues/252) from aristanetworks/master




<a name="v0.1.9"></a>
## [v0.1.9] - 2015-09-22

- Remove gvm from packaging script
- Update deb/rpm package config, package script
- Add -outputfilter flag, and refactor the filter flag to work for -sample-config
- Select default apache port depending on url scheme
- Add port tag to apache plugin
- Update gopsutil godep dependency
- Memory plugin: use 'available' instead of 'actual_'
- Update new memory unit tests, documentation
- Godep update gopsutil to get darwin mem fix
- Refactor memory stats, remove some, add 'actual_' stats
- Fix CPU unit tests for time_ prefix
- Adding time_ prefix to all CPU time measurements
- Adding a retry to the initial telegraf database connection
- Add shebang to postinstall script (fixes installation on Debian family)
- Fix makefile warning for go1.5
- Remove cpu_usage_busy, this is simply 100-cpu_usage_idle
- Add a CPU collection plugin README
- Update gopsutil dependency to enable 32-bit builds
- Remove non-existent 'stolen' cpu stat, fix measurement names
- Properly vendor the gopsutil dependency
- Delete 'vendored' gopsutil directory
- Check if file exists before running disk usage on it. Not all mounts are normal files.
- Revert godep updates, needs a fix in influxdb repo
- Add amqp/rabbitmq to output list in readme
- Changing AddValues to AddFields and temp disabling adding w time
- Update influxdb godeps for line-protocol precision fix
- mysql plugin: don't emit blank tags
- Catching up on some CHANGELOG updates
- install and init script for el5
- Add HTTP 5xx stats to HAProxy plugin. Closes [#194](https://github.com/influxdata/telegraf/issues/194)
- AMQP routing tag doc & add routing tag for Kafka
- added docker image unit test with OpenTSDB
- AMQP output plugin typo fixes and added README and RoutingTag
- Added amqp output
- Merge pull request [#198](https://github.com/influxdata/telegraf/issues/198) from mced/fix_mem_used_perc
- [fix] mem_used_perc returns percentage of used mem
- add bugfix in CHANGELOG and some notes in pg README
- no longer duplicate ignored columns here
- Makes the test also work across pg versions
- add some comments
- fix some more indentation...
- Add a few notes about the connection strings
- uncomment to skip test in short mode
- Generating metric information dynamically. Makes compatible with postgresql versions < 9.2
- added more UNIT test cases for covering all parts of the code
- added prefix settings of the module and rearrange go test code
- added docker image unit test with OpenTSDB
- fix spaces with gofmt
- added readme as suggested / whished in [#177](https://github.com/influxdata/telegraf/issues/177)
- added opentsdb as sink
- adds opentsdb telnet output plugin
- change/fix expected test result
- code improvements after running tests / compile step
- [fix] mem_used_perc returns percentage of used mem
- Add a server name tag to the RabbitMQ server list
- Fix docker stats to make it work on centos 7.
- darwin net plugin fix, really need to godep vendor gopsutil
- Fix multiple redis server bug, do not cache the TCP connections
- Makefile will now honor GOBIN, if set
- Fix bug in setting the precision before gathering metrics
- Support InfluxDB clusters
- Re-arrange repo files for root dir cleanup
- Bump go version number to 1.5
- README updates for systemd and deb/rpm install
- Update telegraf.service and packaging script for systemd
- Update README plugins list
- Put all ARCH binaries on the README


Thank you for your contributions!


### Pull Requests
- Merge pull request [#198](https://github.com/influxdata/telegraf/issues/198) from mced/fix_mem_used_perc




<a name="v0.1.8"></a>
## [v0.1.8] - 2015-09-04

- Makefile rule for building all linux binaries, and upload all ARCHs
- package.sh script fixes for uploading binaries
- Update package script and readme for 0.1.8
- Ping plugin
- Fix default installed config for consistency
- Write data in UTC by default and use 's' precision
- package.sh: upload raw binaries to S3
- add additional metrics to mysql plugin tests
- add additional MySQL metrics
- README: Say when tagpass/tagdrop are valid from.
- Fixup for g->r change, io.reader was already using 'r'
- Redis plugin internal names consistency fix, g -> r
- Add system uptime metric, string formatted AND in float64
- Apache Plugin
- Rename DEPENDENCY_LICENSES LICENSE_OF_DEPENDENCIES
- Add list of dependency licenses
- Update README with 0.1.7 and make separate CONTRIBUTING file


Thank you for your contributions!





<a name="v0.1.7"></a>
## [v0.1.7] - 2015-08-28

- Only build the docker plugin on linux
- Clean up agent error handling and logging of outputs/plugins
- Kafka output producer, send telegraf metrics to Kafka brokers
- Indent the toml config for readability
- Outputs enhancement to require Description and SampleConfig functions
- Improve build from source instructions
- Merge problem, re-enable non-standard DB names
- makefile: ADVERTISED_HOST needs only be set during docker-compose target
- Fixed memory reporting for Linux systems
- Fixed total memory reporting for Darwin systems. hw.memsize is reported as bytes instead of pages.
- Typo: prec -> perc
- Add MySQL server address tag to all measurements
- memcached: fix when a value contains a space
- Vagrantfile: do a one-way rsync so that binaries don't get shared between VMs and host
- Fixes [#130](https://github.com/influxdata/telegraf/issues/130), document mysql plugin better, README
- Add [#136](https://github.com/influxdata/telegraf/issues/136) to CHANGELOG
- Provide a -usage flag for printing the usage of a single plugin
- Fixes [#128](https://github.com/influxdata/telegraf/issues/128), add system load and swap back to default Telegraf config
- Update CHANGELOG.md
- Update CHANGELOG.md
- add plugin.name to error message
- go fmt remove whitespace
- Log plugin errors in crankParallel and crankSeparate cases. Previously errors weren't logged in these cases.
- Update README to point to url without 'v' prepended to version
- Filter out the 'v' from the version tag, issue [#134](https://github.com/influxdata/telegraf/issues/134)
- Fix for [#129](https://github.com/influxdata/telegraf/issues/129) README typo in the 0.1.6 package name url
- Version= doesnt work on go1.4.2
- README typo fix


Thank you for your contributions!





<a name="v0.1.6"></a>
## [v0.1.6] - 2015-08-24

- Filter out the 'v' from the version tag, issue [#134](https://github.com/influxdata/telegraf/issues/134)
- Version= doesnt work on go1.4.2
- 0.1.6, update changelog, readme, plugins list
- godep update influxdb to 0.9.3-rc1
- fix for [#126](https://github.com/influxdata/telegraf/issues/126), nginx plugin not catching net.SplitHostPort error
- Add a simple integration test at the end of circle-test.sh similar to homebrew test
- Change -X main.Version <n> to -X main.Version=<n> for go1.5
- fix segv on error
- packaging script fix, make_dir_tree is req'd
- Fix for issue [#121](https://github.com/influxdata/telegraf/issues/121), update etc/config.sample.toml
- Modifications to httpjson plugin
- Add httpjson plugin
- Update CHANGELOG with some recent additions
- Merge pull request [#118](https://github.com/influxdata/telegraf/issues/118) from srfraser/diskusage_windows_fix
- Fix issue [#119](https://github.com/influxdata/telegraf/issues/119), remove the _workspace/pkg directory from git tracking
- Get disk usage stats working on windows
- Update README to reflect new release of 0.1.4 & 0.1.5
- Updating the packaging script to assume tag has already been set
- Fix build, testify got removed from godeps somehow
- Telegraf 0.1.5, update InfluxDB client to HEAD


Thank you for your contributions!


### Pull Requests
- Merge pull request [#118](https://github.com/influxdata/telegraf/issues/118) from srfraser/diskusage_windows_fix




<a name="v0.1.4"></a>
## [v0.1.4] - 2015-08-18

- Telegraf 0.1.4, update godep to point to InfluxDB client 0.9.2
- Update Makefile with new build requirements
- Add build function to circle-test.sh, and remove release.sh
- godep: vendor all dependencies & add circle-test.sh
- exec plugin doesn't crash when given null JSON values
- README update to address issue [#113](https://github.com/influxdata/telegraf/issues/113)
- Merge branch 'jipperinbham-datadog-output'
- fix tests, remove debug prints
- fix merge conflicts, update import paths
- add datadog output
- Release 0.1.5, updating CHANGELOG and README
- Put quotes around potentially empty bash variables
- Rebase and fixups for PR [#111](https://github.com/influxdata/telegraf/issues/111), fixes issue [#33](https://github.com/influxdata/telegraf/issues/33)
- Adds cpu busy time and percentages
- Removing DefaultConfig function because there's really no point
- README updates for readability and ease of use
- Allow a PerCPU configuration variable, issue [#108](https://github.com/influxdata/telegraf/issues/108)
- circle.yml: verify that golint violations == 0 for some dirs
- Fix influx.toml and ListTags string printing
- add missing import and Tag marshalling
- Merge pull request [#109](https://github.com/influxdata/telegraf/issues/109) from influxdb/pr-107
- Update changelog with PR [#107](https://github.com/influxdata/telegraf/issues/107), thanks [@jipperinbham](https://github.com/jipperinbham)
- Adding a Close() function to the Output interface and to the agent
- Followup to issue [#77](https://github.com/influxdata/telegraf/issues/77), create configured database name from toml file
- move tags to influxdb struct, update all sample configs
- Print version number on startup, issue [#104](https://github.com/influxdata/telegraf/issues/104)
- Followup to issue [#77](https://github.com/influxdata/telegraf/issues/77), create configured database name from toml file
- Update CHANGELOG with fix for issue [#101](https://github.com/influxdata/telegraf/issues/101)
- Fix for issue [#101](https://github.com/influxdata/telegraf/issues/101), switch back from master branch if building locally
- Update CHANGELOG with PR [#106](https://github.com/influxdata/telegraf/issues/106)
- Merge pull request [#106](https://github.com/influxdata/telegraf/issues/106) from zepouet/master
- Go FMT missing Merge branch 'master' of https://github.com/zepouet/telegraf
- Go FMT missing...
- Revert "PR [#59](https://github.com/influxdata/telegraf/issues/59), implementation of multiple outputs"
- PR [#59](https://github.com/influxdata/telegraf/issues/59), implementation of multiple outputs
- Update changelog with PR [#103](https://github.com/influxdata/telegraf/issues/103)
- Ensure tests pass now that we're passing fstype around
- to filter by filesystem type, we need to pass that up the chain
- tag filtering description added
- Modify ShouldPass so that it checks the tags of a metric, if configured.
- Update Readme with new option filter and add usage chapter with --help
- ShouldPass needs to know the tags being used
- Fix for issue [#77](https://github.com/influxdata/telegraf/issues/77), create telegraf database if not exists
- Automate circleci package process
- Back to regular circle.yml, make and artifact linux binaries
- fix filename for logrotate config
- Log rotation configuration file, and package.sh modifications to add it to deb and rpm
- Massive retro-active changelog update
- README long-line fixing and a couple typos
- Fail and exit telegraf if no plugins are found loaded, issue [#26](https://github.com/influxdata/telegraf/issues/26)
- Add LeoFS plugin
- Revert "Add log rotation to /etc/logrotate.d for deb and rpm packages"
- Using gvm & shell test file to manage circleci go environment
- Remove simplejson dependency in exec plugin
- Fix for issue [#93](https://github.com/influxdata/telegraf/issues/93), just use github path instead of gopkg.in
- Add exec plugin
- Add filtering options to select plugin at startup
- Update changelog with PR [#103](https://github.com/influxdata/telegraf/issues/103)
- Ensure tests pass now that we're passing fstype around
- to filter by filesystem type, we need to pass that up the chain
- tag filtering description added
- Modify ShouldPass so that it checks the tags of a metric, if configured.
- ShouldPass needs to know the tags being used
- Fix for issue [#77](https://github.com/influxdata/telegraf/issues/77), create telegraf database if not exists
- Automate circleci package process
- Back to regular circle.yml, make and artifact linux binaries
- fix filename for logrotate config
- Log rotation configuration file, and package.sh modifications to add it to deb and rpm
- Massive retro-active changelog update
- move tags to influxdb struct, update all sample configs
- README long-line fixing and a couple typos
- Fail and exit telegraf if no plugins are found loaded, issue [#26](https://github.com/influxdata/telegraf/issues/26)
- Add LeoFS plugin
- update config sample, marshal tags from toml
- Merge pull request [#96](https://github.com/influxdata/telegraf/issues/96) from influxdb/revert-87-logrotation
- Revert "Add log rotation to /etc/logrotate.d for deb and rpm packages"
- Merge pull request [#92](https://github.com/influxdata/telegraf/issues/92) from Asana/exec
- Using gvm & shell test file to manage circleci go environment
- Remove simplejson dependency in exec plugin
- Fix for issue [#93](https://github.com/influxdata/telegraf/issues/93), just use github path instead of gopkg.in
- resolve remaining build errors
- resolve go vet issues
- fix issue with var rename
- resolve merge conflicts
- convert influxdb output to multiple outputs
- Add exec plugin
- Marking disque tests 'short', circleci container doesnt appear to support tcp?
- Skip per-cpu unit test when in a circle ci container
- Mark more unit tests as 'integration' tests when they rely on external services/docker
- Merge pull request [#71](https://github.com/influxdata/telegraf/issues/71) from kureikain/haproxy_plugin
- Add Nginx plugin (ngx_http_stub_status_module)
- Adding Disque, Lustre, and memcached to the list of supported plugins
- Merge pull request [#76](https://github.com/influxdata/telegraf/issues/76) from kotopes/redis-port-tag
- Merge branch 'gfloyd-disque-plugin'
- Build & unit test fixup
- Adding Kafka docker container and utilizing it in unit tests
- Verify proper go formatting in circleci job
- go fmt fixes
- Adding circleci build badge
- Fix 'go vet' error, +build comment must be followed by a blank line
- Creating circleci job to just lint and vet code
- Add default log rotation
- Tests for the lustre plugin, initial commit
- Require validation for uint64 as well as int64
- Lustre filesystem plugin (http://lustre.org/)
- Add Lustre 2 plugin
- Fix GetLocalHost testutil function for mac users (boot2docker)
- Build & unit test fixup
- Adding Kafka docker container and utilizing it in unit tests
- Verify proper go formatting in circleci job
- go fmt fixes
- Adding circleci build badge
- Merge pull request [#86](https://github.com/influxdata/telegraf/issues/86) from srfraser/lustre2-plugin
- Fix 'go vet' error, +build comment must be followed by a blank line
- Merge branch 'master' of https://github.com/influxdb/telegraf into lustre2-plugin
- Merge pull request [#87](https://github.com/influxdata/telegraf/issues/87) from srfraser/logrotation
- Creating circleci job to just lint and vet code
- Add default log rotation
- Tests for the lustre plugin, initial commit
- Require validation for uint64 as well as int64
- Lustre filesystem plugin (http://lustre.org/)
- Add Lustre 2 plugin
- Fix GetLocalHost testutil function for mac users (boot2docker)
- Add disque plugin
- Merge pull request [#49](https://github.com/influxdata/telegraf/issues/49) from marcosnils/container_services
- Add haproxy plugin
- add tag "port" to every redis metric
- Merge pull request [#53](https://github.com/influxdata/telegraf/issues/53) from alvaromorales/rethinkdb-fix
- Merge pull request [#54](https://github.com/influxdata/telegraf/issues/54) from jipperinbham/mongodb-plugin
- Merge pull request [#55](https://github.com/influxdata/telegraf/issues/55) from brocaar/elasticsearch_plugin
- Merge pull request [#60](https://github.com/influxdata/telegraf/issues/60) from brocaar/connection_timeout
- Merge pull request [#63](https://github.com/influxdata/telegraf/issues/63) from bewiwi/master
- Merge pull request [#64](https://github.com/influxdata/telegraf/issues/64) from vic3lord/systemd_support
- Merge pull request [#72](https://github.com/influxdata/telegraf/issues/72) from vadimtk/master
- Merge pull request [#73](https://github.com/influxdata/telegraf/issues/73) from ianunruh/plugin/rabbitmq
- Add simple RabbitMQ plugin
- Add TokuDB metrics to MySQL plugin
- systemd unit support
- Fix redis : change ending call with "\r\n"
- Use string for InfluxDB timeout duration config.
- Add connection timeout configuration for InfluxDB.
- Fix typo (tranport > transport).
- fix merge conflicts
- add SSL support, change tag to hostname
- Remove that it only reads indices stats.
- Merge remote-tracking branch 'upstream/master' into elasticsearch_plugin
- Cleanup repeated logic.
- Remove indices filter.
- Cleanup tests.
- Implement breakers stats.
- Implement http stats.
- Implement transport stats.
- Implement fs stats.
- Implement network stats.
- Update README.md
- Update README.md
- Update CHANGELOG.md
- Implement thread-pool stats.
- Merge pull request [#56](https://github.com/influxdata/telegraf/issues/56) from EmilS/plugins/kafka-consumer-readme
- Implement JVM stats.
- Implement process stats.
- Implement os stats.
- Refactor parsing "indices" stats.
- Add node-id and node attributes to tags.
- Add node_name to tags.
- Adds README for Kafka consumer plugin
- Check that API reponse is 200.
- Implement Elasticsearch plugin (indices stats).
- add MongoDB plugin
- Add missing files
- Add DOCKER_HOST support for tests
- Add rethinkdb plugin to all.go.
- Add --no-recreate option to prepare target
- Merge pull request [#50](https://github.com/influxdata/telegraf/issues/50) from jseriff/master
- Merge pull request [#52](https://github.com/influxdata/telegraf/issues/52) from benfb/master
- use influxdb/telegraf instead of influxdb/influxdb in changelog
- update init.sh to use telegraf directories
- Use postgres default configuration
- Remove circle ci implementation due to Golang bug.
- Remove unnecessary circleci configuration as we're using default provided services
- Add cirleci script
- Add docker containers to test services.
- Update README.md for v0.1.3


Thank you for your contributions!


### Reverts
- PR [#59](https://github.com/influxdata/telegraf/issues/59), implementation of multiple outputs
- Add log rotation to /etc/logrotate.d for deb and rpm packages
- Add log rotation to /etc/logrotate.d for deb and rpm packages

### Pull Requests
- Merge pull request [#109](https://github.com/influxdata/telegraf/issues/109) from influxdb/pr-107
- Merge pull request [#106](https://github.com/influxdata/telegraf/issues/106) from zepouet/master
- Merge pull request [#96](https://github.com/influxdata/telegraf/issues/96) from influxdb/revert-87-logrotation
- Merge pull request [#92](https://github.com/influxdata/telegraf/issues/92) from Asana/exec
- Merge pull request [#71](https://github.com/influxdata/telegraf/issues/71) from kureikain/haproxy_plugin
- Merge pull request [#76](https://github.com/influxdata/telegraf/issues/76) from kotopes/redis-port-tag
- Merge pull request [#86](https://github.com/influxdata/telegraf/issues/86) from srfraser/lustre2-plugin
- Merge pull request [#87](https://github.com/influxdata/telegraf/issues/87) from srfraser/logrotation
- Merge pull request [#49](https://github.com/influxdata/telegraf/issues/49) from marcosnils/container_services
- Merge pull request [#53](https://github.com/influxdata/telegraf/issues/53) from alvaromorales/rethinkdb-fix
- Merge pull request [#54](https://github.com/influxdata/telegraf/issues/54) from jipperinbham/mongodb-plugin
- Merge pull request [#55](https://github.com/influxdata/telegraf/issues/55) from brocaar/elasticsearch_plugin
- Merge pull request [#60](https://github.com/influxdata/telegraf/issues/60) from brocaar/connection_timeout
- Merge pull request [#63](https://github.com/influxdata/telegraf/issues/63) from bewiwi/master
- Merge pull request [#64](https://github.com/influxdata/telegraf/issues/64) from vic3lord/systemd_support
- Merge pull request [#72](https://github.com/influxdata/telegraf/issues/72) from vadimtk/master
- Merge pull request [#73](https://github.com/influxdata/telegraf/issues/73) from ianunruh/plugin/rabbitmq
- Merge pull request [#56](https://github.com/influxdata/telegraf/issues/56) from EmilS/plugins/kafka-consumer-readme
- Merge pull request [#50](https://github.com/influxdata/telegraf/issues/50) from jseriff/master
- Merge pull request [#52](https://github.com/influxdata/telegraf/issues/52) from benfb/master




<a name="v0.1.3"></a>
## [v0.1.3] - 2015-07-05

- Telegraf should have its own directories.
- Update CHANGELOG.md
- Merge pull request [#45](https://github.com/influxdata/telegraf/issues/45) from jhofeditz/patch-1
- Merge pull request [#28](https://github.com/influxdata/telegraf/issues/28) from brian-brazil/prometheus-plugin-only
- Merge pull request [#47](https://github.com/influxdata/telegraf/issues/47) from jipperinbham/rethinkdb-plugin
- Merge pull request [#43](https://github.com/influxdata/telegraf/issues/43) from marcosnils/mysql_fix
- Merge pull request [#46](https://github.com/influxdata/telegraf/issues/46) from zepouet/master
- add RethinkDB plugin
- Update README.md
- skip disk tags with no value
- Return error when can't execute stats query
- Fix mysql plugin due to test accumulator refactor
- Merge pull request [#35](https://github.com/influxdata/telegraf/issues/35) from EmilS/plugins/kafka
- Add Kafka Consumer Plugin
- Update CHANGELOG.md
- Update README.md
- Update CHANGELOG.md
- Merge pull request [#32](https://github.com/influxdata/telegraf/issues/32) from tylernisonoff/master
- fixed spelling mistake -- memoory -> memory
- Add Prometheus plugin.
- Improve test infrastructure


Thank you for your contributions!


### Pull Requests
- Merge pull request [#45](https://github.com/influxdata/telegraf/issues/45) from jhofeditz/patch-1
- Merge pull request [#28](https://github.com/influxdata/telegraf/issues/28) from brian-brazil/prometheus-plugin-only
- Merge pull request [#47](https://github.com/influxdata/telegraf/issues/47) from jipperinbham/rethinkdb-plugin
- Merge pull request [#43](https://github.com/influxdata/telegraf/issues/43) from marcosnils/mysql_fix
- Merge pull request [#46](https://github.com/influxdata/telegraf/issues/46) from zepouet/master
- Merge pull request [#35](https://github.com/influxdata/telegraf/issues/35) from EmilS/plugins/kafka
- Merge pull request [#32](https://github.com/influxdata/telegraf/issues/32) from tylernisonoff/master




<a name="v0.1.2"></a>
## [v0.1.2] - 2015-06-23

- Cleanup the URL when one isn't specified
- Fix type error using URL as a string
- Add memcached to the all plugins package
- Merge pull request [#21](https://github.com/influxdata/telegraf/issues/21) from fromYukki/memcached
- Merge pull request [#16](https://github.com/influxdata/telegraf/issues/16) from jipperinbham/redis_auth
- Explore "limit_maxbytes" and "bytes" individually
- redis plugin accepts URI or string, support Redis AUTH
- Merge pull request [#19](https://github.com/influxdata/telegraf/issues/19) from sherifzain/master
- Merge pull request [#20](https://github.com/influxdata/telegraf/issues/20) from nkatsaros/master
- Added: server to tags
- Memcached plugin
- protect accumulator values with a mutex
- Fixed: differentiate stats gathered from multiple redis servers/instances
- Create a CHANGELOG.
- Merge pull request [#13](https://github.com/influxdata/telegraf/issues/13) from influxdb/fix-packaging
- Merge pull request [#12](https://github.com/influxdata/telegraf/issues/12) from influxdb/s3-cleanup
- Merge pull request [#14](https://github.com/influxdata/telegraf/issues/14) from voxxit/voxxit-linux-arm
- Add linux/arm to list of built binaries
- Add Homebrew instructions to README.md
- Un-break the packaging script.
- Clean up descriptions and stop pushing to both S3 buckets.
- Fix typo
- Add supported plugins
- Move plugins details into readme
- Update README.md
- Update README.md


Thank you for your contributions!


### Pull Requests
- Merge pull request [#21](https://github.com/influxdata/telegraf/issues/21) from fromYukki/memcached
- Merge pull request [#16](https://github.com/influxdata/telegraf/issues/16) from jipperinbham/redis_auth
- Merge pull request [#19](https://github.com/influxdata/telegraf/issues/19) from sherifzain/master
- Merge pull request [#20](https://github.com/influxdata/telegraf/issues/20) from nkatsaros/master
- Merge pull request [#13](https://github.com/influxdata/telegraf/issues/13) from influxdb/fix-packaging
- Merge pull request [#12](https://github.com/influxdata/telegraf/issues/12) from influxdb/s3-cleanup
- Merge pull request [#14](https://github.com/influxdata/telegraf/issues/14) from voxxit/voxxit-linux-arm




<a name="v0.1.1"></a>
## [v0.1.1] - 2015-06-18

- Add package.sh script
- Add -pidfile and Commit variable


Thank you for your contributions!





<a name="v0.1.0"></a>
## v0.1.0 - 2015-06-17

- Merge pull request [#9](https://github.com/influxdata/telegraf/issues/9) from influxdb/sample-config
- Move config to `etc`
- Merge pull request [#11](https://github.com/influxdata/telegraf/issues/11) from influxdb/fix-measurement
- Fix `measurement` => `Measurement`
- Merge pull request [#8](https://github.com/influxdata/telegraf/issues/8) from influxdb/name-to-measurement
- Remove telegraph.toml config file
- Explicitly name the config file as an example
- Update plugin registry from name -> measurement
- Update name -> measurement
- A set of fixes to fix the tests
- Add 'AddValuesWithTime' function to accumulator
- issue 5 closed, updating readme
- Merge pull request [#7](https://github.com/influxdata/telegraf/issues/7) from influxdb/beckettsean-patch-3
- Update README.md
- use localhost by default
- Tivan is dead, long live Telegraf. Fixes [#3](https://github.com/influxdata/telegraf/issues/3)
- Add the host tag always. Fixes [#4](https://github.com/influxdata/telegraf/issues/4)
- Regenerate sample config. Fixes [#1](https://github.com/influxdata/telegraf/issues/1)
- Improve sample config
- Merge pull request [#2](https://github.com/influxdata/telegraf/issues/2) from influxdb/beckettsean-patch-2
- clarifying readme
- Clearify some required config parameters
- Actually write the points
- Add docs about how to use the Accumulator
- Fix a couple typos
- Start of a PLUGINS.md
- Start of a README.md
- Add pass, drop, and interval to the plugin options
- Grammar
- Add rule about 'localhost'
- Breakup the system plugin
- Include comment about using test in the sample config
- Add ability to restrict which network interfaces are sampled
- Minor usability fixes to config
- Add ability to generate config from available plugins
- Require plugin declaration in config to use any plugin
- Enforce stat prefixing at the accumulator layer
- Sort the plugins so the order is consintent
- Update for newer API
- Enable pg and mysql by default
- Add mysql plugin
- Gather stats from PG and redis from localhost by default
- Add postgresql plugin
- Add ability to query many redis servers
- Add sample config file
- Use _ as the namespace separator
- Add redis plugin
- Apply any configuration to a plugin
- Namespace the system metrics
- Automatically include a 'host' tag
- Fix all imports
- Add release.sh and Vagrantfile
- Add -version option
- Detect docker is not available gracefully
- Connect on run if not connected
- Remove final cypress remnents
- Provide a test mode to check plugins easily
- Use dockerclient to get containers and info
- Report better errors where system stats can't be gathered
- Report cpu stats using tags
- Report that docker isn't available better in psutils
- Fix a few more imports
- Remove debugging
- Fix a couple imports and a float comparison test
- Add docker stats
- Fix docker stats
- Add VM and Swap stats
- Add disk io stats
- Add NetIO
- Add disk usage stats
- Disable gopsutil tests that don't work on darwin
- Switch plugin API to use an accumulator
- Remove neko entirely
- Vendor psutils and remove neko
- Initial spike
- Initial commit


Thank you for your contributions!


### Pull Requests
- Merge pull request [#9](https://github.com/influxdata/telegraf/issues/9) from influxdb/sample-config
- Merge pull request [#11](https://github.com/influxdata/telegraf/issues/11) from influxdb/fix-measurement
- Merge pull request [#8](https://github.com/influxdata/telegraf/issues/8) from influxdb/name-to-measurement
- Merge pull request [#7](https://github.com/influxdata/telegraf/issues/7) from influxdb/beckettsean-patch-3
- Merge pull request [#2](https://github.com/influxdata/telegraf/issues/2) from influxdb/beckettsean-patch-2


[Unreleased]: https://github.com/influxdata/telegraf/compare/1.22.0...HEAD
[1.22.0]: https://github.com/influxdata/telegraf/compare/v1.21.4...1.22.0
[v1.21.4]: https://github.com/influxdata/telegraf/compare/v1.21.3...v1.21.4
[v1.21.3]: https://github.com/influxdata/telegraf/compare/v1.21.2...v1.21.3
[v1.21.2]: https://github.com/influxdata/telegraf/compare/v1.21.1...v1.21.2
[v1.21.1]: https://github.com/influxdata/telegraf/compare/v1.21.0...v1.21.1
[v1.21.0]: https://github.com/influxdata/telegraf/compare/v1.20.4...v1.21.0
[v1.20.4]: https://github.com/influxdata/telegraf/compare/v1.20.3...v1.20.4
[v1.20.3]: https://github.com/influxdata/telegraf/compare/v1.20.2...v1.20.3
[v1.20.2]: https://github.com/influxdata/telegraf/compare/v1.20.1...v1.20.2
[v1.20.1]: https://github.com/influxdata/telegraf/compare/v1.20.0...v1.20.1
[v1.20.0]: https://github.com/influxdata/telegraf/compare/v1.19.3...v1.20.0
[v1.19.3]: https://github.com/influxdata/telegraf/compare/v1.19.2...v1.19.3
[v1.19.2]: https://github.com/influxdata/telegraf/compare/v1.19.1...v1.19.2
[v1.19.1]: https://github.com/influxdata/telegraf/compare/v1.19.0...v1.19.1
[v1.19.0]: https://github.com/influxdata/telegraf/compare/v1.18.3...v1.19.0
[v1.18.3]: https://github.com/influxdata/telegraf/compare/v1.18.2...v1.18.3
[v1.18.2]: https://github.com/influxdata/telegraf/compare/v1.18.1...v1.18.2
[v1.18.1]: https://github.com/influxdata/telegraf/compare/v1.18.0...v1.18.1
[v1.18.0]: https://github.com/influxdata/telegraf/compare/v1.17.3...v1.18.0
[v1.17.3]: https://github.com/influxdata/telegraf/compare/v1.17.2...v1.17.3
[v1.17.2]: https://github.com/influxdata/telegraf/compare/v1.17.1...v1.17.2
[v1.17.1]: https://github.com/influxdata/telegraf/compare/v1.17.0...v1.17.1
[v1.17.0]: https://github.com/influxdata/telegraf/compare/v1.16.3...v1.17.0
[v1.16.3]: https://github.com/influxdata/telegraf/compare/v1.16.2...v1.16.3
[v1.16.2]: https://github.com/influxdata/telegraf/compare/v1.16.1...v1.16.2
[v1.16.1]: https://github.com/influxdata/telegraf/compare/v1.16.0...v1.16.1
[v1.16.0]: https://github.com/influxdata/telegraf/compare/v1.15.4...v1.16.0
[v1.15.4]: https://github.com/influxdata/telegraf/compare/v1.15.3...v1.15.4
[v1.15.3]: https://github.com/influxdata/telegraf/compare/v1.15.2...v1.15.3
[v1.15.2]: https://github.com/influxdata/telegraf/compare/v1.15.1...v1.15.2
[v1.15.1]: https://github.com/influxdata/telegraf/compare/v1.15.0...v1.15.1
[v1.15.0]: https://github.com/influxdata/telegraf/compare/v1.14.5...v1.15.0
[v1.14.5]: https://github.com/influxdata/telegraf/compare/v1.14.4...v1.14.5
[v1.14.4]: https://github.com/influxdata/telegraf/compare/v1.14.3...v1.14.4
[v1.14.3]: https://github.com/influxdata/telegraf/compare/v1.14.2...v1.14.3
[v1.14.2]: https://github.com/influxdata/telegraf/compare/v1.14.1...v1.14.2
[v1.14.1]: https://github.com/influxdata/telegraf/compare/v1.14.0...v1.14.1
[v1.14.0]: https://github.com/influxdata/telegraf/compare/v0.10.1...v1.14.0
[v0.10.1]: https://github.com/influxdata/telegraf/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/influxdata/telegraf/compare/v0.2.4...v0.10.0
[v0.2.4]: https://github.com/influxdata/telegraf/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/influxdata/telegraf/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/influxdata/telegraf/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/influxdata/telegraf/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/influxdata/telegraf/compare/v0.1.9...v0.2.0
[v0.1.9]: https://github.com/influxdata/telegraf/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/influxdata/telegraf/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/influxdata/telegraf/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/influxdata/telegraf/compare/v0.1.4...v0.1.6
[v0.1.4]: https://github.com/influxdata/telegraf/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/influxdata/telegraf/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/influxdata/telegraf/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/influxdata/telegraf/compare/v0.1.0...v0.1.1
