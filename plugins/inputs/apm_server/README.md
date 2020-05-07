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

  ## exclude event types
  exclude_events = ["span"]
  ## exclude fields matching following patterns
  exclude_fields = ["exception_stacktrace_*", "log_stacktrace_*"]
  ## store selected fields as tags 
  tag_keys =[ "context_request_method", "result"]

```

### Metrics

Each incoming event from APM Agent contains two parts: `metadata` and `eventdata`. 
The `metadata` are mapped to LineProtocol's tags and `eventdata` are mapped to LineProtocol's fields.

#### Tags

Each measurement is tagged with the identifiers from `metadata`. Nested objects are represented by `_` notation.

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
apm_server,labels_ab_testing=true,labels_group=experimental,labels_segment=5,process_argv_0=-v,process_pid=1234,
    process_ppid=1,process_title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service_agent_ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service_agent_name=java,service_agent_version=1.10.0,service_environment=production,service_framework_name=spring,
    service_framework_version=5.0.0,service_language_name=Java,service_language_version=10.0.2,service_name=1234_service-12a3,
    service_node_configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service_runtime_name=Java,
    service_runtime_version=10.0.2,service_version=4.3.0,system_architecture=amd64,system_configured_hostname=host1,
    system_container_id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system_detected_hostname=8ec7ceb99074,
    system_kubernetes_namespace=default,system_kubernetes_node_name=node-name,system_kubernetes_pod_name=instrumented-java-service,
    system_kubernetes_pod_uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system_platform=Linux,
    type=metricset 
    field1=field1value, field2=field2value,... 
    1571657444929001000
```

#### Fields

Each incoming `evendata` are mapped into measurement's fields. Nested objects are represented by `_` notation. 

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
apm_server,labels_ab_testing=true,labels_group=experimental,labels_segment=5,process_argv_0=-v,process_pid=1234,
    process_ppid=1,process_title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service_agent_ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service_agent_name=java,service_agent_version=1.10.0,service_environment=production,service_framework_name=spring,
    service_framework_version=5.0.0,service_language_name=Java,service_language_version=10.0.2,service_name=1234_service-12a3,
    service_node_configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service_runtime_name=Java,
    service_runtime_version=10.0.2,service_version=4.3.0,system_architecture=amd64,system_configured_hostname=host1,
    system_container_id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system_detected_hostname=8ec7ceb99074,
    system_kubernetes_namespace=default,system_kubernetes_node_name=node-name,system_kubernetes_pod_name=instrumented-java-service,
    system_kubernetes_pod_uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system_platform=Linux,
    type=metricset 
    samples_transaction.duration.sum.us=12,samples_dotted.float.gauge=6.12,samples_span.self_time.sum.us=633.288,
    span_type="db",tags_code=200,tags_success=true,samples_negative.d.o.t.t.e.d=-1022,samples_transaction.self_time.sum.us=10,
    samples_transaction.self_time.count=2,samples_byte_counter=1,samples_transaction.duration.count=2,
    samples_long_gauge=3147483648,span_subtype="mysql",samples_float_gauge=9.16,samples_transaction.breakdown.count=12,
    samples_double_gauge=3.141592653589793,transaction_name="GET/",samples_short_counter=227,
    samples_span.self_time.count=1,samples_integer_gauge=42767,transaction_type="request" 
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
apm_server,labels_ab_testing=true,labels_group=experimental,labels_segment=5,process_argv_0=-v,process_pid=1234,
    process_ppid=1,process_title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service_agent_ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service_agent_name=java,service_agent_version=1.10.0,service_environment=production,service_framework_name=spring,service_framework_version=5.0.0,
    service_language_name=Java,service_language_version=10.0.2,service_name=1234_service-12a3,service_node_configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,
    service_runtime_name=Java,service_runtime_version=10.0.2,service_version=4.3.0,system_architecture=amd64,system_configured_hostname=host1,
    system_container_id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system_detected_hostname=8ec7ceb99074,system_kubernetes_namespace=default,
    system_kubernetes_node_name=node-name,system_kubernetes_pod_name=instrumented-java-service,system_kubernetes_pod_uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system_platform=Linux,
    type=transaction 
    parent_id="abcdefabcdef01234567",context_custom_my_key=1,context_user_email="foo@mail.com",context_request_env_SERVER_SOFTWARE="nginx",
    context_custom_some_other="foobar",context_request_cookies_c1="v1",context_request_cookies_c2="v2",context_request_socket_encrypted=true,
    context_request_url_pathname="/p/a/t/h",id="4340a8e0df1906ecbfa9",context_request_http_version="1.1",context_request_url_full="https://www.example.com/p/a/t/h?query=string#hash",
    context_user_username="foo",context_request_url_protocol="https:",context_request_body_additional_req="additionalinformation",result="HTTP2xx",
    context_request_url_port="8080",context_custom_and_objects_foo_1="baz",context_request_headers_content-type="text/html",
    context_user_id="99",context_service_name="experimental-java",type="http",context_custom_and_objects_foo_0="bar",
    context_request_body_string="helloworld",context_request_url_hostname="www.example.com",context_response_headers_content-type="application/json",
    context_request_url_hash="#hash",context_request_headers_Elastic-Apm-Traceparent_0="00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01",
    name="ResourceHttpRequestHandler",context_service_agent_ephemeral_id="e71be9ac-93b0-44b9-a997-5638f6ccfc36",
    context_request_url_search="?query=string",context_request_headers_user-agent_1="MozillaChromeEdge",context_response_decoded_body_size=401.9,
    context_request_method="POST",context_response_finished=true,context_request_headers_cookie="c1=v1,c2=v2",context_request_url_raw="/p/a/t/h?query=string#hash",
    sampled=true,context_request_env_GATEWAY_INTERFACE="CGI/1.1",span_count_dropped=0,context_request_socket_remote_address="12.53.12.1:8080",
    span_count_started=17,context_response_transfer_size=300,context_request_body_additional_bar=123,duration=32.592981,
    context_request_headers_user-agent_0="Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36",
    context_tags_organization_uuid="9f0e9d64-c185-4d21-a6f4-4673ed561ec8",context_service_agent_version="1.10.0-SNAPSHOT",
    context_custom_(="notavalidregexandthatisfine",trace_id="0acd456789abcdef0123456789abcdef",context_response_status_code=200,
    context_response_encoded_body_size=356.9,context_response_headers_sent=true 
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
apm_server,labels_ab_testing=true,labels_group=experimental,labels_segment=5,process_argv_0=-v,process_pid=1234,
    process_ppid=1,process_title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service_agent_ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service_agent_name=java,service_agent_version=1.10.0,service_environment=production,service_framework_name=spring,
    service_framework_version=5.0.0,service_language_name=Java,service_language_version=10.0.2,service_name=1234_service-12a3,
    service_node_configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,service_runtime_name=Java,
    service_runtime_version=10.0.2,service_version=4.3.0,system_architecture=amd64,system_configured_hostname=host1,
    system_container_id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,system_detected_hostname=8ec7ceb99074,
    system_kubernetes_namespace=default,system_kubernetes_node_name=node-name,system_kubernetes_pod_name=instrumented-java-service,
    system_kubernetes_pod_uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system_platform=Linux,
    type=span 
    stacktrace_1_module="org.springframework.web.servlet.view",trace_id="abcdef0123456789abcdef9876543210",
    context_http_response_headers_content-type="application/json",context_service_agent_name="java",stacktrace_1_function="render",
    sync=true,context_http_response_encoded_body_size=356,stacktrace_0_lineno=547,context_http_method="GET",
    context_http_response_status_code=200,stacktrace_1_abs_path="/tmp/AbstractView.java",stacktrace_1_library_frame=true,
    context_db_link="other.db.com",context_db_type="sql",stacktrace_1_filename="AbstractView.java",subtype="http",
    context_http_response_transfer_size=300.12,context_service_name="opbeans-java-1",stacktrace_1_colno=4,
    context_db_statement="SELECT * FROM product_types WHERE user_id = ?",type="external",name="GET users-authenticated",
    context_db_instance="customers",id="1234567890aaaade",parent_id="abcdef0123456789",stacktrace_1_context_line="line3",
    context_http_url="http://localhost:8000",context_service_agent_version="1.10.0-SNAPSHOT",action="connect",
    stacktrace_1_vars_key="value",context_db_user="postgres",transaction_id="1234567890987654",duration=3.781912,
    context_http_status_code=302,stacktrace_1_lineno=547,context_service_agent_ephemeral_id="e71be9ac-93b0-44b9-a997-5638f6ccfc36",
    context_http_response_decoded_body_size=401,stacktrace_0_filename="DispatcherServlet.java" 
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
apm_server,labels_ab_testing=true,labels_group=experimental,labels_segment=5,process_argv_0=-v,process_pid=1234,
    process_ppid=1,process_title=/usr/lib/jvm/java-10-openjdk-amd64/bin/java,service_agent_ephemeral_id=e71be9ac-93b0-44b9-a997-5638f6ccfc36,
    service_agent_name=java,service_agent_version=1.10.0,service_environment=production,service_framework_name=spring,
    service_framework_version=5.0.0,service_language_name=Java,service_language_version=10.0.2,service_name=1234_service-12a3,
    service_node_configured_name=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,
    service_runtime_name=Java,service_runtime_version=10.0.2,service_version=4.3.0,system_architecture=amd64,
    system_configured_hostname=host1,system_container_id=8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4,
    system_detected_hostname=8ec7ceb99074,system_kubernetes_namespace=default,system_kubernetes_node_name=node-name,
    system_kubernetes_pod_name=instrumented-java-service,system_kubernetes_pod_uid=b17f231da0ad128dc6c6c0b2e82f6f303d3893e3,system_platform=Linux,
    type=error 
    log_stacktrace_0_context_line="line3",exception_cause_0_cause_0_type="VeryInternalDbError",context_request_url_raw="/p/a/t/h?query=string#hash",
    context_request_headers_Forwarded="for=192.168.0.1",context_response_headers_content-type="application/json",
    context_request_cookies_c1="v1",id="9876543210abcdeffedcba0123456789",exception_cause_0_cause_0_message="disk spinning way too fast",
    exception_stacktrace_0_library_frame=true,exception_stacktrace_0_function="connect",context_custom_and_objects_foo_1="baz",
    context_custom_and_objects_foo_0="bar",exception_stacktrace_0_pre_context_0="line1",exception_stacktrace_0_vars_key="value",
    exception_code=42,exception_stacktrace_1_filename="AbstractClientHttpRequest.java",log_stacktrace_0_colno=4,exception_stacktrace_0_post_context_1="line5",
    context_request_headers_host="opbeans-java:3000",exception_stacktrace_0_context_line="3",context_request_headers_cookie_0="c1=v1",
    exception_stacktrace_0_pre_context_1="line2",log_stacktrace_1_vars_key="value",context_response_status_code=200,context_service_framework_name="Node",
    log_stacktrace_0_lineno=3,exception_stacktrace_0_colno=4,transaction_sampled=true,log_stacktrace_0_vars_key="value",context_request_env_SERVER_SOFTWARE="nginx",
    context_user_id=99,context_request_env_GATEWAY_INTERFACE="CGI/1.1",log_stacktrace_0_post_context_0="line4",context_service_framework_version="1",
    context_custom_my_key=1,context_request_url_port=8080,trace_id="0123456789abcdeffedcba0123456789",log_stacktrace_1_function="executeInternal",
    context_custom_some_other="foobar",context_request_headers_Elastic-Apm-Traceparent="00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01",
    log_stacktrace_0_pre_context_1="line2",culprit="opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)",exception_handled=false,
    exception_cause_0_cause_1_message="on top of it,internet doesn't work",log_logger_name="http404",context_request_url_protocol="https:",
    context_request_body="HelloWorld",exception_stacktrace_0_filename="AbstractPlainSocketImpl.java",transaction_type="request",
    context_request_cookies_c2="v2",exception_cause_0_cause_1_type="ConnectionError",context_service_node_configured_name="node-xyz",
    context_request_http_version="1.1",context_request_url_pathname="/p/a/t/h",context_service_language_version="1.2",
    transaction_id="1234567890987654",exception_cause_0_message="something wrong writing a file",context_request_socket_remote_address="12.53.12.1",
    context_response_headers_sent=true,log_stacktrace_0_filename="Socket.java",context_user_username="foo",context_request_socket_encrypted=true,
    exception_attributes_foo="bar",log_stacktrace_0_pre_context_0="line1",log_message="Request method 'POST' not supported",
    log_stacktrace_0_abs_path="/tmp/Socket.java",exception_stacktrace_0_module="java.net",log_stacktrace_0_post_context_1="line5",
    exception_module="org.springframework.http.client",log_param_message="Request method 'POST' /events/:event not supported",
    log_stacktrace_0_classname="Request::Socket",parent_id="9632587410abcdef",context_request_url_hostname="www.example.com",
    log_level="error",context_tags_organization_uuid="9f0e9d64-c185-4d21-a6f4-4673ed561ec8",exception_stacktrace_0_lineno=3,
    exception_stacktrace_1_lineno=102,log_stacktrace_1_abs_path="/tmp/SimpleBufferingClientHttpRequest.java",
    context_request_url_full="https://www.example.com/p/a/t/h?query=string#hash",context_service_name="service1",
    context_response_finished=true,context_request_method="POST",context_request_headers_cookie_1="c2=v2",
    log_stacktrace_0_function="connect",exception_message="Theusernamerootisunknown",context_request_url_search="?query=string",
    exception_cause_0_type="InternalDbError",log_stacktrace_1_lineno=102,context_request_headers_content-length="0",
    log_stacktrace_0_library_frame=true,exception_stacktrace_1_vars_key="value",exception_stacktrace_0_post_context_0="line4",
    context_request_url_hash="#hash",context_user_email="user@foo.mail",exception_type="java.net.UnknownHostException",
    log_stacktrace_1_filename="SimpleBufferingClientHttpRequest.java",exception_stacktrace_0_abs_path="/tmp/AbstractPlainSocketImpl.java",
    log_stacktrace_0_module="java.net",exception_stacktrace_1_function="execute" 
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
