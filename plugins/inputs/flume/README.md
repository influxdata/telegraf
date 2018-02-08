# Flume Input Plugin

The example plugin gathers metrics about Flume.

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
# specify server via a url matching
[[inputs.flume]]
  server = "http://localhost:6666/metrics
```

### Metrics:

The plugin gathers the results of the flume monitor http server and processes json to proper telegraf measurements. Below is a result of a request from flume monitor server.

```json
{
    ...
    "SINK.k1": {
        "BatchCompleteCount": "13918",
        "BatchEmptyCount": "755",
        "BatchUnderflowCount": "766",
        "ConnectionClosedCount": "0",
        "ConnectionCreatedCount": "1",
        "ConnectionFailedCount": "0",
        "EventDrainAttemptCount": "14312319",
        "EventDrainSuccessCount": "14312319",
        "StartTime": "1517214100653",
        "StopTime": "0",
        "Type": "SINK"
    },
    "SINK.k2": {
        "BatchCompleteCount": "13924",
        "BatchEmptyCount": "764",
        "BatchUnderflowCount": "781",
        "ConnectionClosedCount": "0",
        "ConnectionCreatedCount": "1",
        "ConnectionFailedCount": "0",
        "EventDrainAttemptCount": "14324682",
        "EventDrainSuccessCount": "14324682",
        "StartTime": "1517214100653",
        "StopTime": "0",
        "Type": "SINK"
    },
    ...
}
```

- flume
  - tags:
    - component
  - fields:
    - AppendAcceptedCount
    - AppendBatchAcceptedCount
    - AppendBatchReceivedCount
    - AppendReceivedCount
    - BatchCompleteCount
    - BatchEmptyCount
    - BatchUnderflowCount
    - ChannelCapacity
    - ChannelFillPercentage
    - ChannelSize
    - ConnectionClosedCount
    - ConnectionCreatedCount
    - ConnectionFailedCount
    - EventAcceptedCount
    - EventDrainAttemptCount
    - EventDrainSuccessCount
    - EventPutAttemptCount
    - EventPutSuccessCount
    - EventReceivedCount
    - EventTakeAttemptCount
    - EventTakeSuccessCount
    - KafkaCommitTimer
    - KafkaEmptyCount
    - KafkaEventGetTimer
    - OpenConnectionCount
    - StartTime
    - StopTime
    - Type

### Sample Queries:

Get all measurements in the last hour:
```
SELECT * FROM flume WHERE time > now() - 1h GROUP BY component
```

### Example Output:

```
flume,instance=CHANNEL.c1,host=localhost EventPutSuccessCount="29879406",StartTime="1517214100648",ChannelSize="1242405",EventPutAttemptCount="29879406",EventTakeSuccessCount="28637001",ChannelFillPercentage="12.424050000000001",ChannelCapacity="10000000",StopTime="0",EventTakeAttemptCount="28641866",Type="CHANNEL" 1518069650000000000
flume,instance=SINK.k1,host=localhost BatchUnderflowCount="766",ConnectionFailedCount="0",Type="SINK",EventDrainAttemptCount="14312319",BatchEmptyCount="755",StartTime="1517214100653",StopTime="0",ConnectionClosedCount="0",BatchCompleteCount="13918",ConnectionCreatedCount="1",EventDrainSuccessCount="14312319" 1518069650000000000
```
