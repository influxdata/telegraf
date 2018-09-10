# Telegraf Input Plugin: Fleet

The plugin will gather names of running units from [fleet](https://github.com/coreos/fleet) and the sum total of each running unit. It uses the fleet v1 API to gather data. 

### Configuration:

```toml
# Description
[[inputs.fleet]]
## Works with Fleet HTTP API
## Multiple Hosts from which to read Fleet stats:
	hosts = ["http://localhost:49153/fleet/v1/state"]
```

### Measurements & Fields:

The fields are dynamically generated from the output of the fleet API. Using the ```name``` value.. The values of those fields are the number of containers  with the ```systemdSubState``` value of "running".   
<insert example json output here of both running and not to show what is included and not included>

The unit names will have their instanced id and the @ symbol stripped off.  
For example if you had a unit named ```nginx-1.10.1@35``` the field name would be ```nginx-1.10.1```. 

- fleet
    - ```<dynamic unit name>``` (int)

### Tags:

- All measurements have the following tags:
    - server (name of the host/container telegraf is running on)

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter example -test
* Plugin: fleet, Collection 1
> fleet,host=localhost.local,server=http://fleet.testserver.com:49153/fleet/v1/state some-api=2i,test-application=1i,webapp=1i,nginx=2i,redis=1i 1470615664000000000
```