# TELEGRAF COMMAND LINE OPTIONS 

This section is for developers who want a reference to the command line options available in telegraf.

## Usage:

  telegraf [commands|flags]

### The commands & flags are:

 &nbsp; config              print out full sample configuration to stdout
 &nbsp; version             print the version to stdout

 &nbsp; --aggregator-filter <filter>   filter the aggregators to enable, separator is :
 &nbsp; --config <file>                configuration file to load
 &nbsp; --config-directory <directory> directory containing additional *.conf files
 &nbsp; --plugin-directory             directory containing *.so files, this directory will be searched recursively. Any Plugin found will be loaded and namespaced.
 &nbsp; --debug                        turn on debug logging
 &nbsp; --input-filter <filter>        filter the inputs to enable, separator is :
 &nbsp;  --input-list                   print available input plugins.
 &nbsp; --output-filter <filter>       filter the outputs to enable, separator is :
 &nbsp; --output-list                  print available output plugins.
 &nbsp; --pidfile <file>               file to write our pid to
 &nbsp; --pprof-addr <address>         pprof address to listen on, don't activate pprof if empty
 &nbsp; --processor-filter <filter>    filter the processors to enable, separator is :
 &nbsp; --quiet                        run in quiet mode
 &nbsp; --section-filter               filter config sections to output, separator is :
&nbsp; &nbsp;Valid values are 'agent', 'global_tags', 'outputs',
&nbsp;&nbsp;'processors', 'aggregators' and 'inputs'
&nbsp;  --sample-config                print out full sample configuration
&nbsp;  --once                         enable once mode: gather metrics once, write them, and exit
&nbsp;  --test                         enable test mode: gather metrics once and print them
&nbsp;  --test-wait                    wait up to this many seconds for service
                                 inputs to complete in test or once mode
&nbsp;  --usage <plugin>               print usage for a plugin, ie, 'telegraf --usage mysql'
&nbsp;  --version                      display the version and exit

## Examples:

  ### generate a telegraf config file:
  telegraf config > telegraf.conf

  ### generate config with only cpu input & influxdb output plugins defined
  telegraf --input-filter cpu --output-filter influxdb config

  ### run a single telegraf collection, outputting metrics to stdout
  telegraf --config telegraf.conf --test

  ### run telegraf with all plugins defined in config file
  telegraf --config telegraf.conf

  ### run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf --config telegraf.conf --input-filter cpu:mem --output-filter influxdb

  ### run telegraf with pprof
  telegraf --config telegraf.conf --pprof-addr localhost:6060
