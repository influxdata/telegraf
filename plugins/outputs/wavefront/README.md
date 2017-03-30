# Wavefront Output Plugin

This plugin writes to a [Wavefront](https://www.wavefront.com) proxy, in Wavefront data format over TCP.


## Wavefront Data format

The expected input for Wavefront is specified in the following way:

```
<metric> <value> [<timestamp>] <source|host>=<soureTagValue> [tagk1=tagv1 ...tagkN=tagvN]
```

More information about the Wavefront data format is available [here](https://community.wavefront.com/docs/DOC-1031)


By default, to ease Metrics browsing in the Wavefront UI, metrics are grouped by converting any `_` characters to `.` in the final name.
This behavior can be altered by changing the `metric_separator` and/or the `convert_paths` settings.  
Most illegal characters in the metric name are automatically converted to `-`.  
The `use_regex` setting can be used to ensure all illegal characters are properly handled, but can lead to performance degradation.

## Configuration:

```toml
# Configuration for Wavefront output 
[[outputs.wavefront]]
  ## prefix for metrics keys
  prefix = "my.specific.prefix."

  ## DNS name of the wavefront proxy server
  host = "wavefront.example.com"

  ## Port that the Wavefront proxy server listens on
  port = 2878

  ## wether to use "value" for name of simple fields
  simple_fields = false

  ## character to use between metric and field name.  defaults to . (dot)
  metric_separator = "."

  ## Convert metric name paths to use metricSeperator character
  ## When true (default) will convert all _ (underscore) chartacters in final metric name
  convert_paths = true

  ## Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower
  use_regex = false

  ## point tags to use as the source name for Wavefront (if none found, host will be used)
  source_override = ["hostname", "snmp_host", "node_host"]

  ## Print additional debug information requires debug = true at the agent level
  debug_all = false
```

Parameters:

	Prefix          string
	Host            string
	Port            int
	SimpleFields    bool
	MetricSeparator string
	ConvertPaths    bool
	UseRegex    	bool
	SourceOverride  string
	DebugAll        bool

* `prefix`: String to use as a prefix for all sent metrics.
* `host`: Name of Wavefront proxy server
* `port`: Port that Wavefront proxy server is configured for `pushListenerPorts`
* `simple_fields`: if false (default) metric field names called `value` are converted to empty strings
* `metric_separator`: character to use to separate metric and field names. (default is `_`)
* `convert_paths`: if true (default) will convert all `_` in metric and field names to `metric_seperator`
* `use_regex`: if true (default is false) will use regex to ensure all illegal characters are converted to `-`.  Regex is much slower than the default mode which will catch most illegal characters.  Use with caution.
* `source_override`: ordered list of point tags to use as the source name for Wavefront. Once a match is found, that tag is used as the source for that point.  If no tags are found the host tag will be used.
* `debug_all`: Will output additional debug information.  Requires `debug = true` to be configured at the agent level


##

The Wavefront proxy interface can be simulated with this reader:

```
// wavefront_proxy_mock.go
package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "localhost:2878")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			defer c.Close()
			io.Copy(os.Stdout, c)
		}(conn)
	}
}

```

## Allowed values for metrics

Wavefront allows `integers` and `floats` as input values
