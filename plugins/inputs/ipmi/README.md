# Telegraf ipmi plugin

Get bare metal metrics using the command line utility `ipmitool`

see ipmitool(https://sourceforge.net/projects/ipmitool/files/ipmitool/)

The plugin will use the following command to collect remote host sensor stats:

ipmitool -I lan -H 192.168.1.1 -U USERID -P PASSW0RD sdr

## Measurements

- ipmi_sensor:

    * Tags: `server`,`host`
    * Fields:
      - status
      - value
	
## Configuration

```toml
[[inputs.ipmi]]
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  servers = ["USERID:PASSW0RD@lan(10.20.2.203)"]
```

## Output

> ipmi_sensor,host=10.20.2.203,inst=Ambient\ Temp status=1i,value=20 1458488465012559455
> ipmi_sensor,host=10.20.2.203,inst=Altitude status=1i,value=80 1458488465012688613
> ipmi_sensor,host=10.20.2.203,inst=Avg\ Power status=1i,value=220 1458488465012776511
> ipmi_sensor,host=10.20.2.203,inst=Planar\ 3.3V status=1i,value=3.28 1458488465012861875
> ipmi_sensor,host=10.20.2.203,inst=Planar\ 5V status=1i,value=4.9 1458488465012944188
> ipmi_sensor,host=10.20.2.203,inst=Planar\ 12V status=1i,value=12.04 1458488465013008485
> ipmi_sensor,host=10.20.2.203,inst=Planar\ VBAT status=1i,value=3.04 1458488465013072508
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 1A\ Tach status=1i,value=2610 1458488465013137932
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 1B\ Tach status=1i,value=1775 1458488465013279896
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 2A\ Tach status=1i,value=1972 1458488465013358177
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 2B\ Tach status=1i,value=1275 1458488465013434023
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 3A\ Tach status=1i,value=2929 1458488465013514567
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 3B\ Tach status=1i,value=2125 1458488465013582616
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 1 status=1i,value=0 1458488465013643746
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 2 status=1i,value=0 1458488465013714887
> ipmi_sensor,host=10.20.2.203,inst=Fan\ 3 status=1i,value=0 1458488465013861854

