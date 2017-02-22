# ServiceNow Output Plugin
# ==========================

This plugin writes to a specified MID server insatnce via its metrics api.
To use this plugin you need to place the sn folder under the telegraf outputs folder and and the next line to all.go:
	_ "github.com/influxdata/telegraf/plugins/outputs/sn"

To use this output plugin you need to configure your config files, an example of a configuration would be:
# # Configuration for SN server to send metrics to
 [[outputs.sn]]
#   ## prefix for metrics keys
   prefix = "telegraf."
#	## url of the metric api on the MID side
   url = "http://localhost:8081/api/mid/sa/metrics"
   username = "admin"
   password = "admin"

#   ## Debug true - Prints SN communication
   debug = true


To use telgraf with this output plugin use:
telegraf --config telegraf.conf --output-filter sn