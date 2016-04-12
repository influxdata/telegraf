# Telegraf ipmi plugin

Get bare metal metrics using the command line utility `ipmitool`

see ipmitool(https://sourceforge.net/projects/ipmitool/files/ipmitool/)

The plugin will use the following command to collect remote host sensor stats:

ipmitool -I lan -H 192.168.1.1 -U USERID -P PASSW0RD sdr

## Measurements

- ipmi_sensor:

    * Tags: `name`, `server`, `unit`
    * Fields:
      - status
      - value

## Configuration

```toml
[[inputs.ipmi_sensor]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  servers = ["USERID:PASSW0RD@lan(10.20.2.203)"]
```

## Output

```
> ipmi_sensor,server=10.20.2.203,unit=degrees_c,name=ambient_temp status=1i,value=20 1458488465012559455
> ipmi_sensor,server=10.20.2.203,unit=feet,name=altitude status=1i,value=80 1458488465012688613
> ipmi_sensor,server=10.20.2.203,unit=watts,name=avg_power status=1i,value=220 1458488465012776511
> ipmi_sensor,server=10.20.2.203,unit=volts,name=planar_3.3v status=1i,value=3.28 1458488465012861875
> ipmi_sensor,server=10.20.2.203,unit=volts,name=planar_vbat status=1i,value=3.04 1458488465013072508
> ipmi_sensor,server=10.20.2.203,unit=rpm,name=fan_1a_tach status=1i,value=2610 1458488465013137932
> ipmi_sensor,server=10.20.2.203,unit=rpm,name=fan_1b_tach status=1i,value=1775 1458488465013279896
```
