# Zipkin Input Plugin

This plugin implements the Zipkin http server to gather trace and timing data
needed to troubleshoot latency problems in microservice architectures.

__Please Note:__ This plugin is experimental; Its data schema may be subject to
change based on its main usage cases and the evolution of the OpenTracing
standard.

> [!IMPORTANT]
> This plugin will create high cardinality data, so please take this into
> account when sending data to your output!

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather data from a Zipkin server including trace and timing data
[[inputs.zipkin]]
  ## URL path for span data
  # path = "/api/v1/spans"

  ## Port on which Telegraf listens
  # port = 9411

  ## Maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## Maximum duration before timing out write of the response
  # write_timeout = "10s"
```

The plugin accepts spans in `JSON` or `thrift` if the `Content-Type` is
`application/json` or `application/x-thrift`, respectively.  If `Content-Type`
is not set, then the plugin assumes it is `JSON` format.

## Tracing

This plugin uses Annotations tags and fields to track data from spans

- `TRACE` is a set of spans that share a single root span. Traces are built by
  collecting all Spans that share a traceId.
- `SPAN` is a set of Annotations and BinaryAnnotations that correspond to a
  particular RPC.
- `Annotations` create a metric for each annotation & binary annotation of a
  span. This records an occurrence in time at the beginning and end of each
  request.

  Annotations may have the following values:
  - `CS` (client start) marks the beginning of the span, a request is made.
  - `SR` (server receive) marks the point in time the server receives the request
    and starts processing it. Network latency & clock jitters distinguish this
    from `CS`.
  - `SS` (server send) marks the point in time the server is finished processing
    and sends a request back to client. The difference to `SR` denotes the
    amount of time it took to process the request.
  - `CR` (client receive) marks the end of the span, with the client receiving
    the response from server. RPC is considered complete with this annotation.

## Metrics

- `duration_ns` the time in nanoseconds between the end and beginning of a span

### Tags

- `id` the 64-bit ID of the span.
- `parent_id` an ID associated with a particular child span. If there is no
  child span, `parent_id` is equal to `id`
- `trace_id` the 64-bit or 128-bit ID of a particular trace. Every span in a
  trace uses this ID.
- `name` defines a span

#### Annotations have these additional tags

- `service_name` defines a service
- `annotation` the value of an annotation
- `endpoint_host` listening IPv4 address and, if present, port

#### Binary Annotations have these additional tag

- `service_name` defines a service
- `annotation` the value of an annotation
- `endpoint_host`  listening IPv4 address and, if present, port
- `annotation_key` label describing the annotation

## Example Output

The Zipkin data

```json
[
    {
        "trace_id": 2505404965370368069,
        "name": "Child",
        "id": 8090652509916334619,
        "parent_id": 22964302721410078,
        "annotations": [],
        "binary_annotations": [
            {
                "key": "lc",
                "value": "dHJpdmlhbA==",
                "annotation_type": "STRING",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            }
        ],
        "timestamp": 1498688360851331,
        "duration": 53106
    },
    {
        "trace_id": 2505404965370368069,
        "name": "Child",
        "id": 103618986556047333,
        "parent_id": 22964302721410078,
        "annotations": [],
        "binary_annotations": [
            {
                "key": "lc",
                "value": "dHJpdmlhbA==",
                "annotation_type": "STRING",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            }
        ],
        "timestamp": 1498688360904552,
        "duration": 50410
    },
    {
        "trace_id": 2505404965370368069,
        "name": "Parent",
        "id": 22964302721410078,
        "annotations": [
            {
                "timestamp": 1498688360851325,
                "value": "Starting child #0",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            },
            {
                "timestamp": 1498688360904545,
                "value": "Starting child #1",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            },
            {
                "timestamp": 1498688360954992,
                "value": "A Log",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            }
        ],
        "binary_annotations": [
            {
                "key": "lc",
                "value": "dHJpdmlhbA==",
                "annotation_type": "STRING",
                "host": {
                    "ipv4": 2130706433,
                    "port": 0,
                    "service_name": "trivial"
                }
            }
        ],
        "timestamp": 1498688360851318,
        "duration": 103680
    }
]
```

generated the following metrics

```text
zipkin,id=7047c59776af8a1b,name=child,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=53106000i 1498688360851331000
zipkin,annotation=trivial,annotation_key=lc,endpoint_host=127.0.0.1,id=7047c59776af8a1b,name=child,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=53106000i 1498688360851331000
zipkin,id=17020eb55a8bfe5,name=child,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=50410000i 1498688360904552000
zipkin,annotation=trivial,annotation_key=lc,endpoint_host=127.0.0.1,id=17020eb55a8bfe5,name=child,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=50410000i 1498688360904552000
zipkin,id=5195e96239641e,name=parent,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=103680000i 1498688360851318000
zipkin,annotation=Starting\ child\ #0,endpoint_host=127.0.0.1,id=5195e96239641e,name=parent,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=103680000i 1498688360851318000
zipkin,annotation=Starting\ child\ #1,endpoint_host=127.0.0.1,id=5195e96239641e,name=parent,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=103680000i 1498688360851318000
zipkin,annotation=A\ Log,endpoint_host=127.0.0.1,id=5195e96239641e,name=parent,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=103680000i 1498688360851318000
zipkin,annotation=trivial,annotation_key=lc,endpoint_host=127.0.0.1,id=5195e96239641e,name=parent,parent_id=5195e96239641e,service_name=trivial,trace_id=22c4fc8ab3669045 duration_ns=103680000i 1498688360851318000
```
