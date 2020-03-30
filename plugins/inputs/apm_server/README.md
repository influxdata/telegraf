# APM Server

APM Server is a input plugin that listens for requests sent by Elastic APM Agents. 
These type of events are supported to transform to metrics:

* [Metadata][datamodel_metadata]
* [Spans][datamodel_spans]
* [Transactions][datamodel_transactions]
* [Metrics][datamodel_metrics]
* [Errors][datamodel_errors]

### Supported APM HTTP endpoints
The [APM server specification][apm_endpoints] exposes endpoints for events intake, sourcemap upload, agent configuration and server information. 

The table below describe how this plugin conforms with them:

| APM Endpoint                                          | Path                                          | Response                                              |
|-------------------------------------------------------|-----------------------------------------------|-------------------------------------------------------|
| [Events intake][endpoint_events_intake]               | `/intake/v2/events`, `/intake/v2/rum/events`  | Serialize Events into LineProtocol. See detail below  |
| [Sourcemap upload][endpoint_sourcemap_upload]         | `/assets/v1/sourcemaps`                       | Accept all request without processing sources         |
| [Agent configuration][endpoint_agent_configuration]   | `/config/v1/agents`, `/config/v1/rum/agents`  | Configuration via APM Server is disabled              |
| [Server information][endpoint_server_information]     | `/`                                           | Returns Telegraf APM Server information               |

### Configuration:

```toml
[[inputs.apm_server]]
  ## Address and port to list APM Agents
  service_address = ":8200"

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"
```

### Metrics

Each incoming event from APM Agent contains two parts: `metadata` and `eventdata`. 
The `metadata` are mapped to LineProtocol's tags and `eventdata` are mapped to LineProtocol's fields.

#### Tags

Each measurement is tagged with the identifiers from `metadata`. Nested objects are represented by `dot` notation.

The example of incoming `metadata`:

```json
{
  "metadata": {
    "process": {
      "pid": 1234,
      "title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java",
      "ppid": 1,
      "argv": [
        "-v"
      ]
    },
    "system": {
      "architecture": "amd64",
      "detected_hostname": "8ec7ceb99074",
      "configured_hostname": "host1",
      "platform": "Linux",
      "container": {
        "id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4"
      },
      "kubernetes": {
        "namespace": "default",
        "pod": {
          "uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3",
          "name": "instrumented-java-service"
        },
        "node": {
          "name": "node-name"
        }
      }
    },
    "service": {
      "name": "1234_service-12a3",
      "version": "4.3.0",
      "node": {
        "configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4"
      },
      "environment": "production",
      "language": {
        "name": "Java",
        "version": "10.0.2"
      },
      "agent": {
        "version": "1.10.0",
        "name": "java",
        "ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36"
      },
      "framework": {
        "name": "spring",
        "version": "5.0.0"
      },
      "runtime": {
        "name": "Java",
        "version": "10.0.2"
      }
    },
    "labels": {
      "group": "experimental",
      "ab_testing": true,
      "segment": 5
    }
  }
}
```

and corresponding LineProtocol looks like:

```
apm_server,labels.ab_testing=true,labels.group=experimental,labels.segment=5,process.argv.0=-v,process.pid=1234,
    process.ppid=1,process.title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service.agent.ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service.agent.name=java,service.agent.version=1.10.0,service.environment=production,service.framework.name=spring,
    service.framework.version=5.0.0,service.language.name=Java,service.language.version=10.0.2,service.name=1234_service-12a3,
    service.node.configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service.runtime.name=Java,
    service.runtime.version=10.0.2,service.version=4.3.0,system.architecture=amd64,system.configured_hostname=host1,
    system.container.id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system.detected_hostname=8ec7ceb99074,
    system.kubernetes.namespace=default,system.kubernetes.node.name=node-name,system.kubernetes.pod.name=instrumented-java-service,
    system.kubernetes.pod.uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system.platform=Linux,
    type=metricset field1=field1value, field2=field2value,... 
    1571657444929001000
```

#### Fields

Each incoming `evendata` are mapped into measurement's fields. Nested objects are represented by `dot` notation. 

There are four types of `eventdata`:
1. Metricset
1. Transaction
1. Span
1. Error

The example of incoming events:

##### Metricset

```json
{
  "metricset": {
    "samples": {
      "transaction.breakdown.count": {
        "value": 12
      },
      "transaction.duration.sum.us": {
        "value": 12
      },
      "transaction.duration.count": {
        "value": 2
      },
      "transaction.self_time.sum.us": {
        "value": 10
      },
      "transaction.self_time.count": {
        "value": 2
      },
      "span.self_time.count": {
        "value": 1
      },
      "span.self_time.sum.us": {
        "value": 633.288
      },
      "byte_counter": {
        "value": 1
      },
      "short_counter": {
        "value": 227
      },
      "integer_gauge": {
        "value": 42767
      },
      "long_gauge": {
        "value": 3147483648
      },
      "float_gauge": {
        "value": 9.16
      },
      "double_gauge": {
        "value": 3.141592653589793
      },
      "dotted.float.gauge": {
        "value": 6.12
      },
      "negative.d.o.t.t.e.d": {
        "value": -1022
      }
    },
    "tags": {
      "code": 200,
      "success": true
    },
    "transaction": {
      "type": "request",
      "name": "GET/"
    },
    "span": {
      "type": "db",
      "subtype": "mysql"
    },
    "timestamp": 1571657444929001
  }
}
```

and corresponding LineProtocol looks like:

```
apm_server,labels.ab_testing=true,labels.group=experimental,labels.segment=5,process.argv.0=-v,process.pid=1234,
    process.ppid=1,process.title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service.agent.ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service.agent.name=java,service.agent.version=1.10.0,service.environment=production,service.framework.name=spring,
    service.framework.version=5.0.0,service.language.name=Java,service.language.version=10.0.2,service.name=1234_service-12a3,
    service.node.configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service.runtime.name=Java,
    service.runtime.version=10.0.2,service.version=4.3.0,system.architecture=amd64,system.configured_hostname=host1,
    system.container.id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system.detected_hostname=8ec7ceb99074,
    system.kubernetes.namespace=default,system.kubernetes.node.name=node-name,system.kubernetes.pod.name=instrumented-java-service,
    system.kubernetes.pod.uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system.platform=Linux,
    type=metricset 
    samples.double_gauge.value=3.141592653589793,samples.long_gauge.value=3147483648,
    transaction.type="request",transaction.name="GET/",tags.success=true,span.subtype="mysql",
    samples.transaction.self_time.sum.us.value=10,samples.negative.d.o.t.t.e.d.value=-1022,
    samples.transaction.duration.sum.us.value=12,span.type="db",samples.transaction.self_time.count.value=2,
    samples.float_gauge.value=9.16,samples.short_counter.value=227,samples.transaction.breakdown.count.value=12,
    tags.code=200,samples.span.self_time.sum.us.value=633.288,samples.span.self_time.count.value=1,
    samples.transaction.duration.count.value=2,samples.dotted.float.gauge.value=6.12,samples.integer_gauge.value=42767,
    samples.byte_counter.value=1 
    1571657444929001000
```

##### Transaction

```json
{
  "transaction": {
    "timestamp": 1571657444929001,
    "name": "ResourceHttpRequestHandler",
    "type": "http",
    "id": "4340a8e0df1906ecbfa9",
    "trace_id": "0acd456789abcdef0123456789abcdef",
    "parent_id": "abcdefabcdef01234567",
    "span_count": {
      "started": 17,
      "dropped": 0
    },
    "duration": 32.592981,
    "result": "HTTP2xx",
    "sampled": true,
    "context": {
      "service": {
        "name": "experimental-java",
        "agent": {
          "version": "1.10.0-SNAPSHOT",
          "ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36"
        }
      },
      "request": {
        "socket": {
          "remote_address": "12.53.12.1:8080",
          "encrypted": true
        },
        "http_version": "1.1",
        "method": "POST",
        "url": {
          "protocol": "https:",
          "full": "https://www.example.com/p/a/t/h?query=string#hash",
          "hostname": "www.example.com",
          "port": "8080",
          "pathname": "/p/a/t/h",
          "search": "?query=string",
          "hash": "#hash",
          "raw": "/p/a/t/h?query=string#hash"
        },
        "headers": {
          "user-agent": [
            "Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36",
            "MozillaChromeEdge"
          ],
          "content-type": "text/html",
          "cookie": "c1=v1,c2=v2",
          "Elastic-Apm-Traceparent": [
            "00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01"
          ]
        },
        "cookies": {
          "c1": "v1",
          "c2": "v2"
        },
        "env": {
          "SERVER_SOFTWARE": "nginx",
          "GATEWAY_INTERFACE": "CGI/1.1"
        },
        "body": {
          "string": "helloworld",
          "additional": {
            "foo": {},
            "bar": 123,
            "req": "additionalinformation"
          }
        }
      },
      "response": {
        "status_code": 200,
        "transfer_size": 300,
        "encoded_body_size": 356.90,
        "decoded_body_size": 401.90,
        "headers": {
          "content-type": "application/json"
        },
        "headers_sent": true,
        "finished": true
      },
      "user": {
        "id": "99",
        "username": "foo",
        "email": "foo@mail.com"
      },
      "tags": {
        "organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8",
        "tag5": null
      },
      "custom": {
        "my_key": 1,
        "some_other_value": "foobar",
        "and_objects": {
          "foo": [
            "bar",
            "baz"
          ]
        },
        "(": "notavalidregexandthatisfine"
      }
    }
  }
}
```

and corresponding LineProtocol looks like:

```
apm_server,labels.ab_testing=true,labels.group=experimental,labels.segment=5,process.argv.0=-v,process.pid=1234,
    process.ppid=1,process.title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service.agent.ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service.agent.name=java,service.agent.version=1.10.0,service.environment=production,service.framework.name=spring,
    service.framework.version=5.0.0,service.language.name=Java,service.language.version=10.0.2,service.name=1234_service-12a3,
    service.node.configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service.runtime.name=Java,
    service.runtime.version=10.0.2,service.version=4.3.0,system.architecture=amd64,system.configured_hostname=host1,
    system.container.id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system.detected_hostname=8ec7ceb99074,
    system.kubernetes.namespace=default,system.kubernetes.node.name=node-name,system.kubernetes.pod.name=instrumented-java-service,
    system.kubernetes.pod.uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system.platform=Linux,
    type=metricset 
    span_count.started=17,context.response.encoded_body_size=356.9,trace_id="0acd456789abcdef0123456789abcdef",
    context.request.cookies.c2="v2",context.request.cookies.c1="v1",context.request.url.raw="/p/a/t/h?query=string#hash",
    context.request.url.hash="#hash",context.request.url.protocol="https:",context.request.headers.user-agent.1="MozillaChromeEdge",
    context.request.env.SERVER_SOFTWARE="nginx",context.request.body.additional.bar=123,context.custom.(="notavalidregexandthatisfine",
    context.request.headers.content-type="text/html",name="ResourceHttpRequestHandler",context.request.http_version="1.1",
    context.response.headers_sent=true,context.request.url.port="8080",duration=32.592981,context.user.email="foo@mail.com",
    context.response.finished=true,context.custom.my_key=1,id="4340a8e0df1906ecbfa9",context.response.transfer_size=300,
    context.service.agent.ephemeral_id="e71be9ac-93b0-44b9-a997-5638f6ccfc36",context.request.socket.encrypted=true,
    context.response.status_code=200,context.tags.organization_uuid="9f0e9d64-c185-4d21-a6f4-4673ed561ec8",
    context.request.env.GATEWAY_INTERFACE="CGI/1.1",context.request.body.additional.req="additionalinformation",
    context.request.headers.user-agent.0="Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36",
    context.response.headers.content-type="application/json",context.request.url.hostname="www.example.com",span_count.dropped=0,
    context.request.socket.remote_address="12.53.12.1:8080",context.request.headers.cookie="c1=v1,c2=v2",sampled=true,
    context.request.url.pathname="/p/a/t/h",context.service.agent.version="1.10.0-SNAPSHOT",context.response.decoded_body_size=401.9,
    context.request.body.string="helloworld",context.custom.and_objects.foo.1="baz",context.request.url.search="?query=string",
    context.custom.some_other_value="foobar",context.service.name="experimental-java",context.request.method="POST",result="HTTP2xx",
    type="http",parent_id="abcdefabcdef01234567",context.user.username="foo",context.user.id="99",
    context.request.headers.Elastic-Apm-Traceparent.0="00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01",
    context.custom.and_objects.foo.0="bar",context.request.url.full="https://www.example.com/p/a/t/h?query=string#hash" 
    1571657444929001000
```

##### Span

```json
{
  "span": {
    "timestamp": 1571657444929001,
    "type": "external",
    "subtype": "http",
    "id": "1234567890aaaade",
    "transaction_id": "1234567890987654",
    "trace_id": "abcdef0123456789abcdef9876543210",
    "parent_id": "abcdef0123456789",
    "action": "connect",
    "sync": true,
    "name": "GET users-authenticated",
    "duration": 3.781912,
    "stacktrace": [
      {
        "filename": "DispatcherServlet.java",
        "lineno": 547
      },
      {
        "function": "render",
        "abs_path": "/tmp/AbstractView.java",
        "filename": "AbstractView.java",
        "lineno": 547,
        "library_frame": true,
        "vars": {
          "key": "value"
        },
        "module": "org.springframework.web.servlet.view",
        "colno": 4,
        "context_line": "line3"
      }
    ],
    "context": {
      "db": {
        "instance": "customers",
        "statement": "SELECT * FROM product_types WHERE user_id = ?",
        "type": "sql",
        "user": "postgres",
        "link": "other.db.com"
      },
      "http": {
        "url": "http://localhost:8000",
        "status_code": 302,
        "method": "GET",
        "response": {
          "status_code": 200,
          "transfer_size": 300.12,
          "encoded_body_size": 356,
          "decoded_body_size": 401,
          "headers": {
            "content-type": "application/json"
          }
        }
      },
      "service": {
        "name": "opbeans-java-1",
        "agent": {
          "version": "1.10.0-SNAPSHOT",
          "name": "java",
          "ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36"
        }
      }
    }
  }
}
```

and corresponding LineProtocol looks like:

```
apm_server,labels.ab_testing=true,labels.group=experimental,labels.segment=5,process.argv.0=-v,process.pid=1234,
    process.ppid=1,process.title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service.agent.ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service.agent.name=java,service.agent.version=1.10.0,service.environment=production,service.framework.name=spring,
    service.framework.version=5.0.0,service.language.name=Java,service.language.version=10.0.2,service.name=1234_service-12a3,
    service.node.configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service.runtime.name=Java,
    service.runtime.version=10.0.2,service.version=4.3.0,system.architecture=amd64,system.configured_hostname=host1,
    system.container.id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system.detected_hostname=8ec7ceb99074,
    system.kubernetes.namespace=default,system.kubernetes.node.name=node-name,system.kubernetes.pod.name=instrumented-java-service,
    system.kubernetes.pod.uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system.platform=Linux,
    type=metricset 
    stacktrace.1.function="render",name="GET users-authenticated",type="external",context.http.response.status_code=200,
    context.http.method="GET",duration=3.781912,transaction_id="1234567890987654",stacktrace.1.lineno=547,
    context.http.response.decoded_body_size=401,context.http.response.transfer_size=300.12,context.http.response.headers.content-type="application/json",
    trace_id="abcdef0123456789abcdef9876543210",stacktrace.1.library_frame=true,context.http.response.encoded_body_size=356,
    stacktrace.1.context_line="line3",subtype="http",stacktrace.1.vars.key="value",context.http.status_code=302,
    context.service.agent.version="1.10.0-SNAPSHOT",stacktrace.0.filename="DispatcherServlet.java",stacktrace.0.lineno=547,
    parent_id="abcdef0123456789",context.service.agent.name="java",stacktrace.1.colno=4,stacktrace.1.module="org.springframework.web.servlet.view",
    context.service.name="opbeans-java-1",context.db.statement="SELECT * FROM product_types WHERE user_id = ?",
    context.db.instance="customers",sync=true,stacktrace.1.filename="AbstractView.java",stacktrace.1.abs_path="/tmp/AbstractView.java",
    context.service.agent.ephemeral_id="e71be9ac-93b0-44b9-a997-5638f6ccfc36",action="connect",context.db.link="other.db.com",
    context.http.url="http://localhost:8000",context.db.user="postgres",context.db.type="sql",id="1234567890aaaade" 
    1571657444929001000
```

##### Error

```json
{
  "error": {
    "id": "9876543210abcdeffedcba0123456789",
    "timestamp": 1571657444929001,
    "trace_id": "0123456789abcdeffedcba0123456789",
    "parent_id": "9632587410abcdef",
    "transaction_id": "1234567890987654",
    "transaction": {
      "sampled": true,
      "type": "request"
    },
    "culprit": "opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)",
    "log": {
      "message": "Request method 'POST' not supported",
      "param_message": "Request method 'POST' /events/:event not supported",
      "logger_name": "http404",
      "level": "error",
      "stacktrace": [
        {
          "abs_path": "/tmp/Socket.java",
          "filename": "Socket.java",
          "classname": "Request::Socket",
          "function": "connect",
          "vars": {
            "key": "value"
          },
          "pre_context": [
            "line1",
            "line2"
          ],
          "context_line": "line3",
          "library_frame": true,
          "lineno": 3,
          "module": "java.net",
          "colno": 4,
          "post_context": [
            "line4",
            "line5"
          ]
        },
        {
          "filename": "SimpleBufferingClientHttpRequest.java",
          "lineno": 102,
          "function": "executeInternal",
          "abs_path": "/tmp/SimpleBufferingClientHttpRequest.java",
          "vars": {
            "key": "value"
          }
        }
      ]
    },
    "exception": {
      "message": "Theusernamerootisunknown",
      "type": "java.net.UnknownHostException",
      "handled": true,
      "module": "org.springframework.http.client",
      "code": 42,
      "handled": false,
      "attributes": {
        "foo": "bar"
      },
      "cause": [
        {
          "type": "InternalDbError",
          "message": "something wrong writing a file",
          "cause": [
            {
              "type": "VeryInternalDbError",
              "message": "disk spinning way too fast"
            },
            {
              "type": "ConnectionError",
              "message": "on top of it,internet doesn't work"
            }
          ]
        }
      ],
      "stacktrace": [
        {
          "abs_path": "/tmp/AbstractPlainSocketImpl.java",
          "filename": "AbstractPlainSocketImpl.java",
          "function": "connect",
          "vars": {
            "key": "value"
          },
          "pre_context": [
            "line1",
            "line2"
          ],
          "context_line": "3",
          "library_frame": true,
          "lineno": 3,
          "module": "java.net",
          "colno": 4,
          "post_context": [
            "line4",
            "line5"
          ]
        },
        {
          "filename": "AbstractClientHttpRequest.java",
          "lineno": 102,
          "function": "execute",
          "vars": {
            "key": "value"
          }
        }
      ]
    },
    "context": {
      "request": {
        "socket": {
          "remote_address": "12.53.12.1",
          "encrypted": true
        },
        "http_version": "1.1",
        "method": "POST",
        "url": {
          "protocol": "https:",
          "full": "https://www.example.com/p/a/t/h?query=string#hash",
          "hostname": "www.example.com",
          "port": 8080,
          "pathname": "/p/a/t/h",
          "search": "?query=string",
          "hash": "#hash",
          "raw": "/p/a/t/h?query=string#hash"
        },
        "headers": {
          "Forwarded": "for=192.168.0.1",
          "host": "opbeans-java:3000",
          "content-length": "0",
          "cookie": [
            "c1=v1",
            "c2=v2"
          ],
          "Elastic-Apm-Traceparent": "00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01"
        },
        "cookies": {
          "c1": "v1",
          "c2": "v2"
        },
        "env": {
          "SERVER_SOFTWARE": "nginx",
          "GATEWAY_INTERFACE": "CGI/1.1"
        },
        "body": "HelloWorld"
      },
      "response": {
        "status_code": 200,
        "headers": {
          "content-type": "application/json"
        },
        "headers_sent": true,
        "finished": true
      },
      "user": {
        "id": 99,
        "username": "foo",
        "email": "user@foo.mail"
      },
      "tags": {
        "organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8"
      },
      "custom": {
        "my_key": 1,
        "some_other_value": "foobar",
        "and_objects": {
          "foo": [
            "bar",
            "baz"
          ]
        }
      },
      "service": {
        "name": "service1",
        "node": {
          "configured_name": "node-xyz"
        },
        "language": {
          "version": "1.2"
        },
        "framework": {
          "version": "1",
          "name": "Node"
        }
      }
    }
  }
}
```

and corresponding LineProtocol looks like:

```
apm_server,labels.ab_testing=true,labels.group=experimental,labels.segment=5,process.argv.0=-v,process.pid=1234,
    process.ppid=1,process.title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service.agent.ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service.agent.name=java,service.agent.version=1.10.0,service.environment=production,service.framework.name=spring,
    service.framework.version=5.0.0,service.language.name=Java,service.language.version=10.0.2,service.name=1234_service-12a3,
    service.node.configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service.runtime.name=Java,
    service.runtime.version=10.0.2,service.version=4.3.0,system.architecture=amd64,system.configured_hostname=host1,
    system.container.id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system.detected_hostname=8ec7ceb99074,
    system.kubernetes.namespace=default,system.kubernetes.node.name=node-name,system.kubernetes.pod.name=instrumented-java-service,
    system.kubernetes.pod.uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system.platform=Linux,
    type=metricset 
    exception.stacktrace.0.filename="AbstractPlainSocketImpl.java",context.request.url.protocol="https:",
    log.logger_name="http404",exception.cause.0.type="InternalDbError",context.service.framework.version="1",
    exception.stacktrace.0.pre_context.0="line1",context.custom.and_objects.foo.0="bar",exception.stacktrace.1.function="execute",
    exception.module="org.springframework.http.client",context.request.socket.remote_address="12.53.12.1",
    context.tags.organization_uuid="9f0e9d64-c185-4d21-a6f4-4673ed561ec8",context.request.env.SERVER_SOFTWARE="nginx",
    context.user.email="user@foo.mail",context.custom.and_objects.foo.1="baz",context.request.headers.host="opbeans-java:3000",
    log.stacktrace.0.post_context.1="line5",context.request.headers.Elastic-Apm-Traceparent="00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01",
    exception.message="Theusernamerootisunknown",context.custom.my_key=1,exception.stacktrace.0.context_line="3",log.stacktrace.1.vars.key="value",
    context.user.id=99,context.service.language.version="1.2",context.request.url.port=8080,exception.cause.0.cause.1.type="ConnectionError",
    culprit="opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)",exception.stacktrace.1.vars.key="value",
    log.stacktrace.1.filename="SimpleBufferingClientHttpRequest.java",log.stacktrace.0.vars.key="value",context.request.headers.cookie.0="c1=v1",
    log.stacktrace.0.filename="Socket.java",context.service.name="service1",transaction_id="1234567890987654",log.level="error",
    exception.stacktrace.0.post_context.0="line4",exception.stacktrace.0.lineno=3,log.stacktrace.0.pre_context.1="line2",
    exception.stacktrace.0.library_frame=true,exception.attributes.foo="bar",context.request.headers.content-length="0",
    context.request.method="POST",log.stacktrace.1.lineno=102,log.stacktrace.1.abs_path="/tmp/SimpleBufferingClientHttpRequest.java",
    exception.cause.0.cause.0.type="VeryInternalDbError",log.stacktrace.0.post_context.0="line4",context.request.url.hash="#hash",
    exception.cause.0.message="something wrong writing a file",log.stacktrace.0.module="java.net",context.request.headers.cookie.1="c2=v2",
    exception.handled=false,context.request.env.GATEWAY_INTERFACE="CGI/1.1",context.request.url.raw="/p/a/t/h?query=string#hash",
    exception.stacktrace.1.lineno=102,context.response.status_code=200,exception.stacktrace.1.filename="AbstractClientHttpRequest.java",
    context.request.url.search="?query=string",context.request.socket.encrypted=true,exception.cause.0.cause.0.message="disk spinning way too fast",
    context.request.body="HelloWorld",context.custom.some_other_value="foobar",log.param_message="Request method 'POST' /events/:event not supported",
    log.stacktrace.0.function="connect",log.stacktrace.0.colno=4,transaction.type="request",log.stacktrace.0.library_frame=true,
    context.response.headers.content-type="application/json",context.request.cookies.c2="v2",log.stacktrace.0.lineno=3,
    exception.stacktrace.0.module="java.net",log.stacktrace.0.pre_context.0="line1",log.stacktrace.0.classname="Request::Socket",
    exception.stacktrace.0.pre_context.1="line2",context.request.http_version="1.1",exception.type="java.net.UnknownHostException",
    context.request.url.pathname="/p/a/t/h",exception.cause.0.cause.1.message="on top of it,internet doesn't work",
    context.request.headers.Forwarded="for=192.168.0.1",transaction.sampled=true,log.message="Request method 'POST' not supported",
    exception.stacktrace.0.vars.key="value",exception.stacktrace.0.colno=4,context.user.username="foo",log.stacktrace.0.context_line="line3",
    exception.stacktrace.0.abs_path="/tmp/AbstractPlainSocketImpl.java",context.response.finished=true,exception.stacktrace.0.post_context.1="line5",
    context.request.url.hostname="www.example.com",log.stacktrace.0.abs_path="/tmp/Socket.java",id="9876543210abcdeffedcba0123456789",
    context.service.node.configured_name="node-xyz",context.service.framework.name="Node",context.request.cookies.c1="v1",parent_id="9632587410abcdef",
    context.response.headers_sent=true,log.stacktrace.1.function="executeInternal",exception.stacktrace.0.function="connect",
    exception.code=42,trace_id="0123456789abcdeffedcba0123456789",context.request.url.full="https://www.example.com/p/a/t/h?query=string#hash" 
    1571657444929001000
```

[datamodel_metadata]: https://www.elastic.co/guide/en/apm/get-started/7.6/metadata.html
[datamodel_spans]: https://www.elastic.co/guide/en/apm/get-started/current/transaction-spans.html
[datamodel_transactions]: https://www.elastic.co/guide/en/apm/get-started/current/transactions.html
[datamodel_metrics]: https://www.elastic.co/guide/en/apm/get-started/current/metrics.html
[datamodel_errors]: https://www.elastic.co/guide/en/apm/get-started/current/errors.html
[apm_endpoints]: https://www.elastic.co/guide/en/apm/server/current/intake-api.html
[endpoint_events_intake]: https://www.elastic.co/guide/en/apm/server/current/events-api.html
[endpoint_sourcemap_upload]: https://www.elastic.co/guide/en/apm/server/current/sourcemap-api.html
[endpoint_agent_configuration]: https://www.elastic.co/guide/en/apm/server/current/agent-configuration-api.html
[endpoint_server_information]: https://www.elastic.co/guide/en/apm/server/current/server-info.html
