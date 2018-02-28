# Flume Input Plugin

The example plugin gathers metrics about Flume.

### Configuration:
```toml
# Read metrics from one client
[[inputs.flume]]
  servers = ["http://localhost:6666/metrics]
```

### Metrics:

The plugin gathers the results of the flume monitor http server and processes json to proper telegraf measurements. Below is a result of a request from flume monitor server.

```json
{
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
}
```

- flume
  - tags:
    - component
    - server
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
flume_CHANNEL,server=http://localhost:8080/ping,host=MacBook-Pro.local,component=CHANNEL.c1 EventPutSuccessCount="29879406",ChannelCapacity="10000000",StopTime="0",Type="CHANNEL",EventPutAttemptCount="29879406",EventTakeSuccessCount="28637001",StartTime="1517214100648",ChannelFillPercentage="12.424050000000001",ChannelSize="1242405",EventTakeAttemptCount="28641866" 1518335280000000000
flume_SINK,server=http://localhost:8080/ping,host=MacBook-Pro.local,component=SINK.k1 StopTime="0",Type="SINK",BatchCompleteCount="13918",BatchEmptyCount="755",ConnectionClosedCount="0",ConnectionCreatedCount="1",ConnectionFailedCount="0",EventDrainAttemptCount="14312319",StartTime="1517214100653",BatchUnderflowCount="766",EventDrainSuccessCount="14312319" 1518335280000000000
```
