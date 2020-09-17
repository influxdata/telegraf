# telemetry

This package provides basic interaction with the New Relic Metrics and Spans
HTTP APIs, automatic batch harvesting on a given schedule, and handling of
errors from the API response.

## Usage

Create a Harvester. It will store your metrics and spans and send this data in
the background.

  ```go
  harvester := telemetry.NewHarvester(
    telemetry.ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
  )
  ```

Record metrics and/or spans.

  ```go
  harvester.RecordMetric(Gauge{
    Name:       "Temperature",
    Attributes: map[string]interface{}{"zip": "zap"},
    Value:      55.62,
    Timestamp:  time.Now(),
  })
  harvester.RecordSpan(Span{
    ID:          "12345",
    TraceID:     "67890",
    Name:        "mySpan",
    Timestamp:   time.Now(),
    Duration:    time.Second,
    ServiceName: "Getting-Started",
    Attributes: map[string]interface{}{
      "color": "purple",
    },
  })
  ```

Data will be sent to New Relic every 5 seconds by default.
