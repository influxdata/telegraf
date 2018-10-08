# Seq Output Plugin

This plugin writes metrics as structured events to a [Seq](http://www.getseq.net/) instance. Metrics are sent in batches of JSON Lines formatted using Serilog's [compact json event format](https://github.com/serilog/serilog-formatting-compact). Telegraf _field_ and _tag_ collections are serialized with the event and are made available to Seq as top-level properties.

An optional API key can be configured to enable additional tagging and filtering within Seq.

### Configuration

```
[[outputs.seq]]

  ## Seq Instance URL
  seq_instance = "https://localhost:5341" # required

  ## Seq API Key
  seq_api_key = "MYAPIKEY"

  ## Connection timeout.
  # timeout = "5s"
```

### Example outputs
Here is an example metric from the Windows Performance Counters plugin as received by Seq.
```
{
    "@t":"2018-10-07T22:58:30.0000000-07:00",
    "@mt":"Telegraf Measurement {Name} on {Host}",
    "@m":"Telegraf Measurement win_cpu on HP01",
    "@i":"2daeba6e",
    "Fields":{
        "Percent_DPC_Time":0,
        "Percent_Idle_Time":87.126838684082031,
        "Percent_Interrupt_Time":0.10417003184556961,
        "Percent_Privileged_Time":1.1979553699493408,
        "Percent_Processor_Time":10.622114181518555,
        "Percent_User_Time":9.4273872375488281
    },
    "Host":"HP01",
    "Name":"win_cpu",
    "Tags":{
        "host":"HP01",
        "instance":"3",
        "objectname":"Processor"
    }
}

```

### Example queries
In general, you will be constructing Seq queries using the values available in _Fields_ and the dimensions available in _Tags_.

Given the Windows performance counter metric above, you might write a query as such to plot the mean cpu time on a line chart:

```
select mean(Fields.Percent_User_Time) as mean
from stream 
where Name = 'win_cpu'
group by time(30m)
limit 10000
```