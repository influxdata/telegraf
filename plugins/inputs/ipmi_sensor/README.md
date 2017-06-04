# Telegraf ipmi plugin

Get bare metal metrics using the command line utility `ipmitool`

see ipmitool(https://sourceforge.net/projects/ipmitool/files/ipmitool/)

If no servers are specified, the plugin will query the local machine sensor stats via the following command:

```
ipmitool sdr
```

When one or more servers are specified, the plugin will use the following command to collect remote host sensor stats:

```
ipmitool -I lan -H SERVER -U USERID -P PASSW0RD sdr
```

## Measurements

- ipmi_sensor:

    * Tags: `name`, `unit`
    * Fields:
      - status
      - value

The `server` tag will be made available when retrieving stats from remote server(s).

## Configuration

```toml
# Read metrics from the bare metal servers via IPMI
[[inputs.ipmi_sensor]]
  ## optionally specify the path to the ipmitool executable
  # path = "/usr/bin/ipmitool"
  #
  ## optionally specify one or more servers via a url matching
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  ## if no servers are specified, local machine sensor stats will be queried
  ##
  # servers = ["USERID:PASSW0RD@lan(192.168.1.1)"]

  ## Recomended: use metric 'interval' that is a multiple of 'timeout' to avoid
  ## gaps or overlap in pulled data
  interval = "30s"

  ## Timeout for the ipmitool command to complete. Default is 20 seconds.
  timeout = "20s"
```

## Output

When retrieving stats from a remote server:
```
> ipmi_sensor,server=10.20.2.203,unit=degrees_c,name=ambient_temp status=1i,value=20 1458488465012559455
> ipmi_sensor,server=10.20.2.203,unit=feet,name=altitude status=1i,value=80 1458488465012688613
> ipmi_sensor,server=10.20.2.203,unit=watts,name=avg_power status=1i,value=220 1458488465012776511
> ipmi_sensor,server=10.20.2.203,unit=volts,name=planar_3.3v status=1i,value=3.28 1458488465012861875
> ipmi_sensor,server=10.20.2.203,unit=volts,name=planar_vbat status=1i,value=3.04 1458488465013072508
> ipmi_sensor,server=10.20.2.203,unit=rpm,name=fan_1a_tach status=1i,value=2610 1458488465013137932
> ipmi_sensor,server=10.20.2.203,unit=rpm,name=fan_1b_tach status=1i,value=1775 1458488465013279896
```

When retrieving stats from the local machine (no server specified):
```
> ipmi_sensor,unit=degrees_c,name=ambient_temp status=1i,value=20 1458488465012559455
> ipmi_sensor,unit=feet,name=altitude status=1i,value=80 1458488465012688613
> ipmi_sensor,unit=watts,name=avg_power status=1i,value=220 1458488465012776511
> ipmi_sensor,unit=volts,name=planar_3.3v status=1i,value=3.28 1458488465012861875
> ipmi_sensor,unit=volts,name=planar_vbat status=1i,value=3.04 1458488465013072508
> ipmi_sensor,unit=rpm,name=fan_1a_tach status=1i,value=2610 1458488465013137932
> ipmi_sensor,unit=rpm,name=fan_1b_tach status=1i,value=1775 1458488465013279896
```
