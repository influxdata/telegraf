# Telegraf Plugin: Burrow

Collect Kafka topics and consumers status from [Burrow](https://github.com/linkedin/Burrow) HTTP Endpoints.

### Configuration:

```
[[inputs.burrow]]
  ## Burrow HTTP endpoint urls.
  urls = ["http://burrow-service.com:8000"]
  ## Clusters to fetch data. Default to fetch all.
  #clusters = []
  ## Topics to monitor. Default to monitor all.
  #topics = []
  ## Groups to monitor. Default to monitor all.
  #groups = []
```

### Measurements & Fields:

- Measurement
    - burrow_topic
    - burrow_consumer

### Tags:

- `burrow_topic` has the following tags:
    - cluster
    - topic
    - partition

- `burrow_consumer` has the following tags:
    - cluster
    - group
    - topic
    - partition
    - status

