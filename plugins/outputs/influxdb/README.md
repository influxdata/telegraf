# InfluxDB Output Plugin

This plugin writes to [InfluxDB](https://www.influxdb.com) via HTTP or UDP.

Required parameters:

* `urls`: List of strings, this is for InfluxDB clustering
support. On each flush interval, Telegraf will randomly choose one of the urls
to write to. Each URL should start with either `http://` or `udp://`
* `database`: The name of the database to write to.


