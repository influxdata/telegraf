# External Plugins

This is a list of plugins that can be compiled outside of Telegraf and used via the `execd` [input](plugins/inputs/execd), [output](plugins/outputs/execd), or [processor](plugins/processors/execd). 
Check out the [external plugin documentation](/docs/EXTERNAL_PLUGINS.md) for more information on writing and contributing a plugin. 

Pull requests welcome.

## Inputs
- [awsalarms](https://github.com/vipinvkmenon/awsalarms) - Simple plugin to gather/monitor alarms generated  in AWS.
- [octoprint](https://github.com/BattleBas/octoprint-telegraf-plugin) - Gather 3d print information from the octoprint API.
- [open-hardware-monitor](https://github.com/marianob85/open_hardware_monitor-telegraf-plugin) - Gather sensors data provided by [Open Hardware Monitor](http://openhardwaremonitor.org)
- [plex](https://github.com/russorat/telegraf-webhooks-plex) - Listens for events from Plex Media Server [Webhooks](https://support.plex.tv/articles/115002267687-webhooks/).
- [rand](https://github.com/ssoroka/rand) - Generate random numbers
- [SMCIPMITool](https://github.com/jhpope/smc_ipmi) - Python script to parse the output of [SMCIPMITool](https://www.supermicro.com/en/solutions/management-software/ipmi-utilities) into [InfluxDB line protocol](https://docs.influxdata.com/influxdb/latest/reference/syntax/line-protocol/). 
- [systemd-timings](https://github.com/pdmorrow/telegraf-execd-systemd-timings) - Gather systemd boot and unit timestamp metrics.
- [twitter](https://github.com/inabagumi/twitter-telegraf-plugin) - Gather account information from Twitter accounts
- [youtube](https://github.com/inabagumi/youtube-telegraf-plugin) - Gather account information from YouTube channels

## Outputs
- [kinesis](https://github.com/morfien101/telegraf-output-kinesis) - Aggregation and compression of metrics to send Amazon Kinesis.

## Processors
 - [geoip](https://github.com/a-bali/telegraf-geoip) - Add GeoIP information to IP addresses.
 
