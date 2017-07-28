# Zipkin Plugin

This plugin implements the Zipkin http server to gather trace and timing data needed to troubleshoot latency problems in microservice architectures.



### Configuration:
```toml
[[inputs.zipkin]]
    path = "/api/v1/spans" #Path on which Telegraf listens for spans
    port = 9411 # Port on which Telegraf listens
```

### Tracing:

*This plugin uses Annotations tags and fields to track data from spans

- TRACE : is a set of spans that share a single root span.
Traces are built by collecting all Spans that share a traceId.

- SPAN : is a set of Annotations and BinaryAnnotations that correspond to a particular RPC.

- Annotations : for each annotation & binary annotation of a span a metric is output


#### Annotations: records an occurrence in time at the beginning and end of a request
    - CS (client start) : beginning of span, request is made.
    - SR (server receive): server receives request and will start processing it
      network latency & clock jitters differ it from cs
    - SS (server send) : server is done processing and sends request back to client
      amount of time it took to process request will differ it from sr
    - CR (client receive): end of span, client receives response from server
      RPC is considered complete with this annotation

- TAGS:
      _"id":_               The 64 or 128-bit ID of the trace. Every span in a trace shares this ID.
      _"parent_id":_        An ID associated with a particular child span.  If there is no child span, the parent ID is set to itself.
      _"trace_id":_        The 64 or 128-bit ID of a particular trace. Trace ID High concat Trace ID Low.
      _"name":_             Defines a span
      _"service_name":_     Defines a service
      _"annotation_value":_ Defines each individual annotation
      _"endpoint_host":_    listening port concat with IPV4

-FIELDS
      "annotation_timestamp": Start time of an annotation.  If time is nil we set it to the current UTC time.
      "duration":             The time in microseconds between the end and beginning of a span.

### BINARY ANNOTATIONS:

-TAGS: Contains the same tags as annotations plus these additions

      "key": Acts as a pointer to some address which houses the value
      _"type"_: Given data type

-FIELDS:

      "duration": The time in microseconds between the end and beginning of a span.



### Sample Queries:

- Get All Span Names for Service `my_web_server`
```sql
SHOW TAG VALUES FROM "zipkin" with key="name" WHERE "service_name" = 'my_web_server'```
    - __Description:__  returns a list containing the names of the spans which have annotations with the given `service_name` of `my_web_server`.

- Get All Service Names
    ```sql
    SHOW TAG VALUES FROM "zipkin" WITH KEY = "service_name"```
    - __Description:__  returns a list of all `distinct` endpoint service names.

- Find spans with longest duration
    ```sql
    SELECT max("duration") FROM "zipkin" WHERE "service_name" = 'my_service' AND "name" = 'my_span_name' AND time > now() - 20m GROUP BY "trace_id",time(30s) LIMIT 5
    ```
    - __Description:__  In the last 20 minutes find the top 5 longest span durations for service `my_server` and span name `my_span_name`



### Example Input Trace:

- [Cli microservice with two services Test](https://github.com/openzipkin/zipkin-go-opentracing/tree/master/examples/cli_with_2_services)
- [Test data from distributed trace repo sample json](https://github.com/mattkanwisher/distributedtrace/blob/master/testclient/sample.json)

#### Trace Example
``` {
      "traceId": "bd7a977555f6b982",
      "name": "query",
      "id": "be2d01e33cc78d97",
      "parentId": "ebf33e1a81dc6f71",
      "timestamp": 1458702548786000,
      "duration": 13000,
      "annotations": [
        {
          "endpoint": {
            "serviceName": "zipkin-query",
            "ipv4": "192.168.1.2",
            "port": 9411
          },
          "timestamp": 1458702548786000,
          "value": "cs"
        },
        {
          "endpoint": {
            "serviceName": "zipkin-query",
            "ipv4": "192.168.1.2",
            "port": 9411
          },
          "timestamp": 1458702548799000,
          "value": "cr"
        }
      ],
      "binaryAnnotations": [
        {
          "key": "jdbc.query",
          "value": "select distinct `zipkin_spans`.`trace_id` from `zipkin_spans` join `zipkin_annotations` on (`zipkin_spans`.`trace_id` = `zipkin_annotations`.`trace_id` and `zipkin_spans`.`id` = `zipkin_annotations`.`span_id`) where (`zipkin_annotations`.`endpoint_service_name` = ? and `zipkin_spans`.`start_ts` between ? and ?) order by `zipkin_spans`.`start_ts` desc limit ?",
          "endpoint": {
            "serviceName": "zipkin-query",
            "ipv4": "192.168.1.2",
            "port": 9411
          }
        },
        {
          "key": "sa",
          "value": true,
          "endpoint": {
            "serviceName": "spanstore-jdbc",
            "ipv4": "127.0.0.1",
            "port": 3306
          }
        }
      ]
    },```

### Recommended installation
