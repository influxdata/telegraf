<a name="unreleased"></a>
## [Unreleased]




<a name="1.22.0"></a>
## [1.22.0] - 2022-02-23

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

Thank you for your contributions!



Sven Rebhan ,
Sergey Vilgelm ,
Mya ,
Joshua Powers ,
Sebastian Spaink ,
Laurent Sesquès ,
Vladislav ,
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

Thank you for your contributions!



Sven Rebhan ,
Sergey Vilgelm ,
Sebastian Spaink ,
Laurent Sesquès ,
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



<a name="v1.19.1"></a>
## [v1.19.1] - 2021-07-07



<a name="v1.19.0"></a>
## [v1.19.0] - 2021-06-17

### Bug Fixes

- Beat readme title ([#8938](https://github.com/influxdata/telegraf/issues/8938))
- Verify checksum of Go download in mac script ([#9335](https://github.com/influxdata/telegraf/issues/9335))

Thank you for your contributions!



Russ Savage ,
pierwill ,
### Features

- Add external Big blue button plugin ([#9090](https://github.com/influxdata/telegraf/issues/9090))
- Adding Plex Webhooks external plugin ([#8898](https://github.com/influxdata/telegraf/issues/8898))

Thank you for your contributions!



Russ Savage ,
pierwill ,
LEDUNOIS Simon ,


<a name="v1.18.3"></a>
## [v1.18.3] - 2021-05-21



<a name="v1.18.2"></a>
## [v1.18.2] - 2021-04-30



<a name="v1.18.1"></a>
## [v1.18.1] - 2021-04-07



<a name="v1.18.0"></a>
## [v1.18.0] - 2021-03-17

### Bug Fixes

- Beat readme title ([#8938](https://github.com/influxdata/telegraf/issues/8938))
- reading multiple holding registers in modbus input plugin ([#8628](https://github.com/influxdata/telegraf/issues/8628))
- remove ambiguity on '\v' from line-protocol parser ([#8720](https://github.com/influxdata/telegraf/issues/8720))

Thank you for your contributions!



Russ Savage ,
Antonio Garcia ,
Adrian Thurston ,
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



<a name="v1.17.2"></a>
## [v1.17.2] - 2021-01-28



<a name="v1.17.1"></a>
## [v1.17.1] - 2021-01-27



<a name="v1.17.0"></a>
## [v1.17.0] - 2020-12-18

### Bug Fixes

- **exec:** fix typo in exec readme ([#8265](https://github.com/influxdata/telegraf/issues/8265))
- **ras:** update readme title ([#8266](https://github.com/influxdata/telegraf/issues/8266))

Thank you for your contributions!



Russ Savage ,
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



<a name="v1.16.2"></a>
## [v1.16.2] - 2020-11-13



<a name="v1.16.1"></a>
## [v1.16.1] - 2020-10-28



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

Thank you for your contributions!



Russ Savage ,
Roger Peppe ,
### Features

- add functionality to get values from redis commands ([#8196](https://github.com/influxdata/telegraf/issues/8196))

Thank you for your contributions!



Russ Savage ,
Roger Peppe ,
Yoofi Quansah ,### Reverts
- update influxdb v2 port




<a name="v1.15.4"></a>
## [v1.15.4] - 2020-10-21



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



<a name="v1.15.1"></a>
## [v1.15.1] - 2020-07-22



<a name="v1.15.0"></a>
## [v1.15.0] - 2020-07-22



<a name="v1.14.5"></a>
## [v1.14.5] - 2020-06-30



<a name="v1.14.4"></a>
## [v1.14.4] - 2020-06-09



<a name="v1.14.3"></a>
## [v1.14.3] - 2020-05-19



<a name="v1.14.2"></a>
## [v1.14.2] - 2020-04-28



<a name="v1.14.1"></a>
## [v1.14.1] - 2020-04-14



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

Thank you for your contributions!



Jonathan Chauncey ,
Sergio Jimenez ,
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
### Pull Requests
- Merge pull request [#533](https://github.com/influxdata/telegraf/issues/533) from influxdata/fix-interval-option-v0.10




<a name="v0.10.0"></a>
## [v0.10.0] - 2016-01-11
### Pull Requests
- Merge pull request [#497](https://github.com/influxdata/telegraf/issues/497) from influxdata/rm-package-updates




<a name="v0.2.4"></a>
## [v0.2.4] - 2015-12-08



<a name="v0.2.3"></a>
## [v0.2.3] - 2015-11-30



<a name="v0.2.2"></a>
## [v0.2.2] - 2015-11-18



<a name="v0.2.1"></a>
## [v0.2.1] - 2015-11-16
### Reverts
- redis: support IPv6 addresses with no port




<a name="v0.2.0"></a>
## [v0.2.0] - 2015-10-27
### Pull Requests
- Merge pull request [#252](https://github.com/influxdata/telegraf/issues/252) from aristanetworks/master




<a name="v0.1.9"></a>
## [v0.1.9] - 2015-09-22
### Pull Requests
- Merge pull request [#198](https://github.com/influxdata/telegraf/issues/198) from mced/fix_mem_used_perc




<a name="v0.1.8"></a>
## [v0.1.8] - 2015-09-04



<a name="v0.1.7"></a>
## [v0.1.7] - 2015-08-28



<a name="v0.1.6"></a>
## [v0.1.6] - 2015-08-24
### Pull Requests
- Merge pull request [#118](https://github.com/influxdata/telegraf/issues/118) from srfraser/diskusage_windows_fix




<a name="v0.1.4"></a>
## [v0.1.4] - 2015-08-18
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



<a name="v0.1.0"></a>
## v0.1.0 - 2015-06-17
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
