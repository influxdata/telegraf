# chronyc Input Plugin

Get standard chrony metrics, requires chronyc executable. Compatible from chrony 2.4 onwards, tested with 3.2.

Produces many metrics from output of each chronyc command. When you want to get overview of chrony performance, try:

	chronyc_commands = ["tracking"]

For more in-depth metrics, you may want to explore other commands. Some of the commands require root privileges. 
When you run telegraf from unprivileged user, set:

    use_sudo = true

In this case, you need to configure sudo. For example:

	Defaults:telegraf !syslog
	telegraf ALL=(root) NOPASSWD:/usr/bin/chronyc

### Measurements:

The plugin produce measurements under common name **chronyc**.

### Tags:

All measurements have tags:

- command (chronyc command, from which the value has been taken)

Some of points, where appropriate, tagged with:

- clockId (remote or local clock identifier)
- clockIdHex (the same in hex form, from ntpdata)

### Example Output:

```
$ telegraf --config telegraf.conf --input-filter chronyc --test
> chronyc,clockId=chrony,command=tracking,host=freak freqResidual=-0,freqSkew=0.003,frequency=-17.886,lastOffset=-0.000000111,leapStatus="Normal",refId="PPS",refIdHex="50505300",refTime=1543064066.6897483,rmsOffset=0.000000115,rootDelay=0.000001,rootDispersion=0.000014628,stratum=1i,systemTimeOffset=0.000000424,updateInterval=16 1543064081000000000
```




