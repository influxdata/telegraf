# Enrichment Processor Plugin

The `enrichment` plugin enrichs metrics with external tags based on existing tags as filtering condition.

### Configuration:
```toml
  [[processor.enrichment]]
  ## Enrich with external Tags from an external json file set by EnrichFilePath.
  ##
  ## Conditionnal enrichment based on source tags already added by input plugin
  ## There are 2 levels of filtering. Level1 Source Tag ---> Level2 Source Tag ---> Tags to add
  ## If one level of filtering (default) is used the plugin looks for the wellknown level2
  ## Tag "LEVEL1TAGS" in the json file.
  ## The json file as read periodically every RefreshPeriod minutes. (by default 60m)
  ## See README file for more info about the Json file structure.
  ##
  enrichfilepath = ""
  twolevels = false
  refreshperiod = 60
  ## Filtering input tags
  ## Tags set by input plugin used as filter conditions
  ## Level2TagKey is only required when TwoLevel is set to true
  level1tagkey = ""
  level2tagkey = ""
```
### Tags:

New 'String' Tags might be added by this plugin depending on the level of filtering and only if imput metrics convey the tag(s) that are used as filtering condition. 

### JSON source file

The json file which is selected by the **enrichfilepath** option must have the following structure:

The level1TagKey and level2TagKey are tag's keys extracted from the input metric. The **LEVEL1TAGS** entry is a well-known entry that allows to add some custom tags for all metrics that convey a input tag matching **level1TagKey** at least.

By default as the **twolevels** option is disabled the plugin only checks on one level of tag (level 1). If you set the **twolevels** option, the plugin first checks if **level1TagKey** tag is present in the input metric then it adds the **LEVEL1TAGS** custom tags if present. Then it checks if **level2TagKey** is present in the input metric and if yes adds the level 2 custom tags. 

If no matching found the metric is forwarded without any change. 
```json
{
    "<level1TagKey>": {
        "LEVEL1TAGS": {
            "<yourTagKey>": "<yourTagValue>",
            "<yourTagKey>": "<yourTagValue>",
            ...
        },
        "<level2TagKey>":{
            "<yourTagKey>": "<yourTagValue>",
            "<yourTagKey>": "<yourTagValue>",
            ...
        },
        "<level2TagKey>":{
            ...
        }
    },
    "<level1TagKey>": {
        "LEVEL1TAGS": {
            "<yourTagKey>": "<yourTagValue>",
            "<yourTagKey>": "<yourTagValue>",
            ...
        },
        "<level2TagKey>":{
            "<yourTagKey>": "<yourTagValue>",
            "<yourTagKey>": "<yourTagValue>",
            ...
        },
        "<level2TagKey>":{
            ...
        }
    }
}
```
### Example:

Here, we are using the **jti_openconfig_telemetry** input plugin to retrieve streaming telemetry information coming from Junos devices. 

We request this following sensor path to collect traffic statistics of physical network interfaces of the routers:

```
sensors = ["/interfaces/interface/state/counters"]
```

By default this input plugin add the following tags: 

```
device = <JUNOS-DEVICE-NAME>

   and for the above sensor path it also adds:

/interfaces/interface/@name = <physical network interface name>
```
Without enrichment plugin, a given metric (influx encoding) has this structure: 

*Some tags/fields have been removed to help the reading*

```
/interfaces/interface/state/counters,device=parisR01,host=telegraf01,/interfaces/interface/@name=xe-9/0/12 /interfaces/interface/state/counters/out-unicast-pkts=153259i,/interfaces/interface/state/counters/in-pkts=1229597i 1558535920475228790
```
As observed there are the **device** and **/interfaces/interface/@name** tags set by this input plugin. 

Now we want to add some new tags. Requierements are :

- For each router add a tag **POP** with the POP ID of the router
- For each physical interface add 2 tags. The first one is called **IFDESC** and the second one **CATEGORY**

For that we first create a json file **/var/tmp/mydb.json** like that:

*This file will be updated by an external process every day*

```json
{
    "parisR01": {
        "LEVEL1TAGS": {
            "POP": "PARIS1"
        },
        "xe-9/0/12":{
            "IFDESC": "C;IX;TO_IX_R01;",
            "CATEGORY": "customer"
        },
        "et-1/0/0":{
            "IFDESC": "B;CORE;TO_parisR02;",
            "CATEGORY": "core"
        }
    },
    "parisR02": {
        "LEVEL1TAGS": {
            "POP": "PARIS1"
        },
        "et-1/0/0":{
            "IFDESC": "B;CORE;TO_parisR01;",
            "CATEGORY": "core"
        }
    },
    "rennesR01": {
        "LEVEL1TAGS": {
            "POP": "RENNES"
        }
    }
}
```
Now we add in **telegraf.conf** the enrichment processor with this following configuration:

```toml
  [[processor.enrichment]]
  
  enrichfilepath = "/var/tmp/mydb.json"
  twolevels = true
  # Refresh every 24h 
  refreshperiod = 1440 
  level1tagkey = "device"
  level2tagkey = "/interfaces/interface/@name"
```
After restarting telegraf we can see that our metrics have been enriched with some tags depending on our filtering conditions:

```
/interfaces/interface/state/counters,device=parisR01,host=telegraf01,POP="PARIS1",IFDESC="C;IX;TO_IX_R01;",CATEGORY="customer",/interfaces/interface/@name=xe-9/0/12 /interfaces/interface/state/counters/out-unicast-pkts=198259i,/interfaces/interface/state/counters/in-pkts=1329597i 155853592345228790
```
