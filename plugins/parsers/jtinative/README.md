# JTINative Parser

JTINative parser decodes Juniper Networks native telemetry sensors protobuf messages
that are streamed from a Juniper network device.

https://www.juniper.net/documentation/en_US/junos/topics/concept/junos-telemetry-interface-oveview.html

### Configuration

This parser is intended to be used with the socket_listener plugin.

```toml
[[inputs.socket_listener]]
    ## URL to listen on
    service_address = "udp4://:50001"
    ##
    ## Data format to consume.
    ## Each data format has its own unique set of configuration options, read
    ## more about them here:
    ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
    data_format = "jtinative"
    ##
    ## All strings will be converted to tags, be false strings will be fields unless
    ## string is defined as a key in the relative protobuf file.
    jti_str_as_tag = false
    ##
    ## List of paths to convert to tags
    jti_convert_tag = []
    ##
    ## List of paths to convert to fields, this will override Key options in
    ## protobuf message as well jti_str_as_tag configuration option.
    jti_convert_field = []
    ##
    ## By default the measurement name will be the sensor name, using globing
    ## style pattern matching we can define overrides to this behaviour
    ## globs listed within each section are not guaranteed to be ordered However,
    ## each section will be evaluated in the order they appear in the configuration
    ##
    ## Example:
    ##  sensor name would first be compared to interface_telemetry and npu_telemetry
    ##  then would compare against the jnpr_telemetry glob.
    [inputs.socket_listener.jti_measurement_override]
      "*/interface" = "interface_telemetry"
      "*interface*" = "interface_telemetry"
      "*/npu" = "npu_telemetry"
    ##
    [inputs.socket_listener.jti_measurement_override]
      "*" = "jnpr_telemetry"
    ##
    ## Tag override, overrides the name for a given tag or tags that match a glob
    ## glob is matched against a full path as a tag might be referenced at different
    ## layers of a protobuf message hierarchy
    [inputs.socket_listener.jti_tag_override]
      "*.IfName" = "interface"

```

### Router Configuration

Below configuration will export interface sensors to the a Telegraf server on
10.10.10.10:50001.

```
serivces {
    analytics {
        streaming-server test-telemetry-collector {
            remote-address 10.10.10.10;
            remote-port 50001;
        }
        export-profile telemetry-export {
            local-address 192.168.1.1;
            reporting-rate 10;
            format gpb;
            transport udp;
        }
        sensor test-interface-telemetry {
            server-name test-telemetry-collector;
            export-name telemetry-export;
            resource /junos/system/linecard/interface/;
        }
    }
}
```
