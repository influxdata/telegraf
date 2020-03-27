package apm_server

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func newTestServer() *APMServer {
	server := &APMServer{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
	}
	_ = internal.SetVersion("0.0.1")
	return server
}

func TestNotMappedPath(t *testing.T) {

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get server information
	resp, err := http.Get(createURL(server, "http", "/not-mapped", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	require.EqualValues(t, 404, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"error\":\"404 page not found\"}", string(body))
}

func TestServerInformation(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	server.buildSHA = "bc4d9a286a65b4283c2462404add86a26be61dca"
	server.buildDate = time.Date(
		2009, 11, 17, 20, 34, 58, 0, time.UTC)

	// get server information
	resp, err := http.Get(createURL(server, "http", "/", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	require.EqualValues(t, 200, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"build_date\":\"2009-11-17T20:34:58Z\","+
		"\"build_sha\":\"bc4d9a286a65b4283c2462404add86a26be61dca\",\"version\":\"0.0.1\"}", string(body))
}

func TestAgentConfiguration(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get agent configuration
	resp, err := http.Get(createURL(server, "http", "/config/v1/agents", "service.name=TEST"))
	require.NoError(t, err)
	require.EqualValues(t, 403, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))
	defer resp.Body.Close()
}

func TestRUMAgentConfiguration(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get RUM agent configuration
	request, err := http.NewRequest("GET", createURL(server, "http", "/config/v1/rum/agents", ""), nil)
	require.NoError(t, err)

	request.Header.Set("Origin", "https://foo.example")
	resp, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	require.EqualValues(t, 403, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))
	require.Equal(t, "https://foo.example", resp.Header.Get("Access-Control-Allow-Origin"))
	defer resp.Body.Close()
}

func TestRUMAgentConfigurationPreflight(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get RUM agent configuration
	request, err := http.NewRequest("OPTIONS", createURL(server, "http", "/config/v1/rum/agents", ""), nil)
	require.NoError(t, err)

	request.Header.Set("Origin", "https://foo.example")
	resp, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)
	require.Equal(t, "https://foo.example", resp.Header.Get("Access-Control-Allow-Origin"))
	require.Equal(t, "POST, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	require.Equal(t, "Content-Type, Content-Encoding, Accept", resp.Header.Get("Access-Control-Allow-Headers"))
	require.Equal(t, "Etag", resp.Header.Get("Access-Control-Expose-Headers"))
	require.Equal(t, "86400", resp.Header.Get("Access-Control-Max-Age"))
	require.Equal(t, "Origin", resp.Header.Get("Vary"))
	require.Equal(t, 0, len(acc.Metrics))
	defer resp.Body.Close()
}

func TestSourceMap(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	sourceMap := `{
  "version" : 3,
  "sources" : ["index.js"],
  "names" : ["resultNum","operator","el","element"],
  "mappings" : "CAAC,WACC,aAyGA,IAAK"
}`

	// post SourceMap
	resp, err := http.Post(createURL(server, "http", "/assets/v1/sourcemaps", "service.name=TEST"), "application/json", bytes.NewBufferString(sourceMap))
	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))
	defer resp.Body.Close()
}

func TestEventsIntakeInvalidHeader(t *testing.T) {

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// post invalid intake
	resp, err := http.Post(createURL(server, "http", "/intake/v2/events", ""), "application/json", bytes.NewBufferString("{}"))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	require.EqualValues(t, 400, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"error\":\"invalid content type: 'application/json'\"}", string(body))
}

func TestEventsIntake(t *testing.T) {
	for _, test := range []struct {
		name     string
		metadata string
		event    string
		tags     map[string]string
		fields   map[string]interface{}
	}{
		{
			name:     "metricset mapping",
			metadata: "metadata.ndjson",
			event:    "metricset.ndjson",
			tags:     map[string]string{"labels.ab_testing": "true", "labels.group": "experimental", "labels.segment": "5", "process.argv.0": "-v", "process.pid": "1234", "process.ppid": "1", "process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service.agent.name": "java", "service.agent.version": "1.10.0", "service.environment": "production", "service.framework.name": "spring", "service.framework.version": "5.0.0", "service.language.name": "Java", "service.language.version": "10.0.2", "service.name": "1234_service-12a3", "service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service.runtime.name": "Java", "service.runtime.version": "10.0.2", "service.version": "4.3.0", "system.architecture": "amd64", "system.configured_hostname": "host1", "system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system.detected_hostname": "8ec7ceb99074", "system.kubernetes.namespace": "default", "system.kubernetes.node.name": "node-name", "system.kubernetes.pod.name": "instrumented-java-service", "system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system.platform": "Linux", "type": "metricset"},
			fields:   map[string]interface{}{"samples.byte_counter.value": 1.0, "samples.dotted.float.gauge.value": 6.12, "samples.double_gauge.value": 3.141592653589793, "samples.float_gauge.value": 9.16, "samples.integer_gauge.value": 42767.0, "samples.long_gauge.value": 3.147483648e+09, "samples.negative.d.o.t.t.e.d.value": -1022.0, "samples.short_counter.value": 227.0, "samples.span.self_time.count.value": 1.0, "samples.span.self_time.sum.us.value": 633.288, "samples.transaction.breakdown.count.value": 12.0, "samples.transaction.duration.count.value": 2.0, "samples.transaction.duration.sum.us.value": 12.0, "samples.transaction.self_time.count.value": 2.0, "samples.transaction.self_time.sum.us.value": 10.0, "span.subtype": "mysql", "span.type": "db", "tags.code": 200.0, "tags.success": true, "transaction.name": "GET/", "transaction.type": "request"},
		},
		{
			name:     "transaction mapping",
			metadata: "metadata.ndjson",
			event:    "transaction.ndjson",
			tags:     map[string]string{"labels.ab_testing": "true", "labels.group": "experimental", "labels.segment": "5", "process.argv.0": "-v", "process.pid": "1234", "process.ppid": "1", "process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service.agent.name": "java", "service.agent.version": "1.10.0", "service.environment": "production", "service.framework.name": "spring", "service.framework.version": "5.0.0", "service.language.name": "Java", "service.language.version": "10.0.2", "service.name": "1234_service-12a3", "service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service.runtime.name": "Java", "service.runtime.version": "10.0.2", "service.version": "4.3.0", "system.architecture": "amd64", "system.configured_hostname": "host1", "system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system.detected_hostname": "8ec7ceb99074", "system.kubernetes.namespace": "default", "system.kubernetes.node.name": "node-name", "system.kubernetes.pod.name": "instrumented-java-service", "system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system.platform": "Linux", "type": "transaction"},
			fields:   map[string]interface{}{"context.custom.(": "notavalidregexandthatisfine", "context.custom.and_objects.foo.0": "bar", "context.custom.and_objects.foo.1": "baz", "context.custom.my_key": 1.0, "context.custom.some_other_value": "foobar", "context.request.body.additional.bar": 123.0, "context.request.body.additional.req": "additionalinformation", "context.request.body.string": "helloworld", "context.request.cookies.c1": "v1", "context.request.cookies.c2": "v2", "context.request.env.GATEWAY_INTERFACE": "CGI/1.1", "context.request.env.SERVER_SOFTWARE": "nginx", "context.request.headers.Elastic-Apm-Traceparent.0": "00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01", "context.request.headers.content-type": "text/html", "context.request.headers.cookie": "c1=v1,c2=v2", "context.request.headers.user-agent.0": "Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36", "context.request.headers.user-agent.1": "MozillaChromeEdge", "context.request.http_version": "1.1", "context.request.method": "POST", "context.request.socket.encrypted": true, "context.request.socket.remote_address": "12.53.12.1:8080", "context.request.url.full": "https://www.example.com/p/a/t/h?query=string#hash", "context.request.url.hash": "#hash", "context.request.url.hostname": "www.example.com", "context.request.url.pathname": "/p/a/t/h", "context.request.url.port": "8080", "context.request.url.protocol": "https:", "context.request.url.raw": "/p/a/t/h?query=string#hash", "context.request.url.search": "?query=string", "context.response.decoded_body_size": 401.9, "context.response.encoded_body_size": 356.9, "context.response.finished": true, "context.response.headers.content-type": "application/json", "context.response.headers_sent": true, "context.response.status_code": 200.0, "context.response.transfer_size": 300.0, "context.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "context.service.agent.version": "1.10.0-SNAPSHOT", "context.service.name": "experimental-java", "context.tags.organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "context.user.email": "foo@mail.com", "context.user.id": "99", "context.user.username": "foo", "duration": 32.592981, "id": "4340a8e0df1906ecbfa9", "name": "ResourceHttpRequestHandler", "parent_id": "abcdefabcdef01234567", "result": "HTTP2xx", "sampled": true, "span_count.dropped": 0.0, "span_count.started": 17.0, "trace_id": "0acd456789abcdef0123456789abcdef", "type": "http"},
		},
		{
			name:     "span mapping",
			metadata: "metadata.ndjson",
			event:    "span.ndjson",
			tags:     map[string]string{"labels.ab_testing": "true", "labels.group": "experimental", "labels.segment": "5", "process.argv.0": "-v", "process.pid": "1234", "process.ppid": "1", "process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service.agent.name": "java", "service.agent.version": "1.10.0", "service.environment": "production", "service.framework.name": "spring", "service.framework.version": "5.0.0", "service.language.name": "Java", "service.language.version": "10.0.2", "service.name": "1234_service-12a3", "service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service.runtime.name": "Java", "service.runtime.version": "10.0.2", "service.version": "4.3.0", "system.architecture": "amd64", "system.configured_hostname": "host1", "system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system.detected_hostname": "8ec7ceb99074", "system.kubernetes.namespace": "default", "system.kubernetes.node.name": "node-name", "system.kubernetes.pod.name": "instrumented-java-service", "system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system.platform": "Linux", "type": "span"},
			fields:   map[string]interface{}{"action": "connect", "context.db.instance": "customers", "context.db.link": "other.db.com", "context.db.statement": "SELECT * FROM product_types WHERE user_id = ?", "context.db.type": "sql", "context.db.user": "postgres", "context.http.method": "GET", "context.http.response.decoded_body_size": 401.0, "context.http.response.encoded_body_size": 356.0, "context.http.response.headers.content-type": "application/json", "context.http.response.status_code": 200.0, "context.http.response.transfer_size": 300.12, "context.http.status_code": 302.0, "context.http.url": "http://localhost:8000", "context.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "context.service.agent.name": "java", "context.service.agent.version": "1.10.0-SNAPSHOT", "context.service.name": "opbeans-java-1", "duration": 3.781912, "id": "1234567890aaaade", "name": "GET users-authenticated", "parent_id": "abcdef0123456789", "stacktrace.0.filename": "DispatcherServlet.java", "stacktrace.0.lineno": 547.0, "stacktrace.1.abs_path": "/tmp/AbstractView.java", "stacktrace.1.colno": 4.0, "stacktrace.1.context_line": "line3", "stacktrace.1.filename": "AbstractView.java", "stacktrace.1.function": "render", "stacktrace.1.library_frame": true, "stacktrace.1.lineno": 547.0, "stacktrace.1.module": "org.springframework.web.servlet.view", "stacktrace.1.vars.key": "value", "subtype": "http", "sync": true, "trace_id": "abcdef0123456789abcdef9876543210", "transaction_id": "1234567890987654", "type": "external"},
		},
		{
			name:     "error mapping",
			metadata: "metadata.ndjson",
			event:    "error.ndjson",
			tags:     map[string]string{"labels.ab_testing": "true", "labels.group": "experimental", "labels.segment": "5", "process.argv.0": "-v", "process.pid": "1234", "process.ppid": "1", "process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service.agent.name": "java", "service.agent.version": "1.10.0", "service.environment": "production", "service.framework.name": "spring", "service.framework.version": "5.0.0", "service.language.name": "Java", "service.language.version": "10.0.2", "service.name": "1234_service-12a3", "service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service.runtime.name": "Java", "service.runtime.version": "10.0.2", "service.version": "4.3.0", "system.architecture": "amd64", "system.configured_hostname": "host1", "system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system.detected_hostname": "8ec7ceb99074", "system.kubernetes.namespace": "default", "system.kubernetes.node.name": "node-name", "system.kubernetes.pod.name": "instrumented-java-service", "system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system.platform": "Linux", "type": "error"},
			fields:   map[string]interface{}{"context.custom.and_objects.foo.0": "bar", "context.custom.and_objects.foo.1": "baz", "context.custom.my_key": 1.0, "context.custom.some_other_value": "foobar", "context.request.body": "HelloWorld", "context.request.cookies.c1": "v1", "context.request.cookies.c2": "v2", "context.request.env.GATEWAY_INTERFACE": "CGI/1.1", "context.request.env.SERVER_SOFTWARE": "nginx", "context.request.headers.Elastic-Apm-Traceparent": "00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01", "context.request.headers.Forwarded": "for=192.168.0.1", "context.request.headers.content-length": "0", "context.request.headers.cookie.0": "c1=v1", "context.request.headers.cookie.1": "c2=v2", "context.request.headers.host": "opbeans-java:3000", "context.request.http_version": "1.1", "context.request.method": "POST", "context.request.socket.encrypted": true, "context.request.socket.remote_address": "12.53.12.1", "context.request.url.full": "https://www.example.com/p/a/t/h?query=string#hash", "context.request.url.hash": "#hash", "context.request.url.hostname": "www.example.com", "context.request.url.pathname": "/p/a/t/h", "context.request.url.port": 8080.0, "context.request.url.protocol": "https:", "context.request.url.raw": "/p/a/t/h?query=string#hash", "context.request.url.search": "?query=string", "context.response.finished": true, "context.response.headers.content-type": "application/json", "context.response.headers_sent": true, "context.response.status_code": 200.0, "context.service.framework.name": "Node", "context.service.framework.version": "1", "context.service.language.version": "1.2", "context.service.name": "service1", "context.service.node.configured_name": "node-xyz", "context.tags.organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "context.user.email": "user@foo.mail", "context.user.id": 99.0, "context.user.username": "foo", "culprit": "opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)", "exception.attributes.foo": "bar", "exception.cause.0.cause.0.message": "disk spinning way too fast", "exception.cause.0.cause.0.type": "VeryInternalDbError", "exception.cause.0.cause.1.message": "on top of it,internet doesn't work", "exception.cause.0.cause.1.type": "ConnectionError", "exception.cause.0.message": "something wrong writing a file", "exception.cause.0.type": "InternalDbError", "exception.code": 42.0, "exception.handled": false, "exception.message": "Theusernamerootisunknown", "exception.module": "org.springframework.http.client", "exception.stacktrace.0.abs_path": "/tmp/AbstractPlainSocketImpl.java", "exception.stacktrace.0.colno": 4.0, "exception.stacktrace.0.context_line": "3", "exception.stacktrace.0.filename": "AbstractPlainSocketImpl.java", "exception.stacktrace.0.function": "connect", "exception.stacktrace.0.library_frame": true, "exception.stacktrace.0.lineno": 3.0, "exception.stacktrace.0.module": "java.net", "exception.stacktrace.0.post_context.0": "line4", "exception.stacktrace.0.post_context.1": "line5", "exception.stacktrace.0.pre_context.0": "line1", "exception.stacktrace.0.pre_context.1": "line2", "exception.stacktrace.0.vars.key": "value", "exception.stacktrace.1.filename": "AbstractClientHttpRequest.java", "exception.stacktrace.1.function": "execute", "exception.stacktrace.1.lineno": 102.0, "exception.stacktrace.1.vars.key": "value", "exception.type": "java.net.UnknownHostException", "id": "9876543210abcdeffedcba0123456789", "log.level": "error", "log.logger_name": "http404", "log.message": "Request method 'POST' not supported", "log.param_message": "Request method 'POST' /events/:event not supported", "log.stacktrace.0.abs_path": "/tmp/Socket.java", "log.stacktrace.0.classname": "Request::Socket", "log.stacktrace.0.colno": 4.0, "log.stacktrace.0.context_line": "line3", "log.stacktrace.0.filename": "Socket.java", "log.stacktrace.0.function": "connect", "log.stacktrace.0.library_frame": true, "log.stacktrace.0.lineno": 3.0, "log.stacktrace.0.module": "java.net", "log.stacktrace.0.post_context.0": "line4", "log.stacktrace.0.post_context.1": "line5", "log.stacktrace.0.pre_context.0": "line1", "log.stacktrace.0.pre_context.1": "line2", "log.stacktrace.0.vars.key": "value", "log.stacktrace.1.abs_path": "/tmp/SimpleBufferingClientHttpRequest.java", "log.stacktrace.1.filename": "SimpleBufferingClientHttpRequest.java", "log.stacktrace.1.function": "executeInternal", "log.stacktrace.1.lineno": 102.0, "log.stacktrace.1.vars.key": "value", "parent_id": "9632587410abcdef", "trace_id": "0123456789abcdeffedcba0123456789", "transaction.sampled": true, "transaction.type": "request", "transaction_id": "1234567890987654"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {

			server := newTestServer()
			acc := &testutil.Accumulator{}
			require.NoError(t, server.Init())
			require.NoError(t, server.Start(acc))
			defer server.Stop()

			metadataFile := fmt.Sprintf("./testdata/%s", test.metadata)
			metadataBytes, _ := ioutil.ReadFile(metadataFile)

			eventFile := fmt.Sprintf("./testdata/%s", test.event)
			eventBytes, _ := ioutil.ReadFile(eventFile)

			buffer := bytes.NewBuffer(metadataBytes)
			buffer.Write(eventBytes)

			resp, err := http.Post(createURL(server, "http", "/intake/v2/events", ""), "application/x-ndjson", buffer)
			require.NoError(t, err)
			require.EqualValues(t, 202, resp.StatusCode)
			require.Equal(t, 1, len(acc.Metrics))
			require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
			require.Equal(t, test.tags, acc.Metrics[0].Tags)
			require.Equal(t, test.fields, acc.Metrics[0].Fields)
			require.Equal(t, "2019-10-21 11:30:44.929001 +0000 UTC", acc.Metrics[0].Time.String())
			defer resp.Body.Close()
		})
	}
}

func TestEventsIntakeMultipleMetadata(t *testing.T) {
	tags1 := map[string]string{"process.pid": "12345", "process.ppid": "1", "process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields1 := map[string]interface{}{"tags.code": 200.0, "tags.success": true, "transaction.name": "GET/", "transaction.type": "request"}
	tags2 := map[string]string{"process.pid": "54321", "process.ppid": "1", "process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields2 := map[string]interface{}{"tags.code": 200.0, "tags.success": true, "transaction.name": "POST/", "transaction.type": "request"}

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	metadataBytes1 := []byte(`{"metadata":{"process":{"pid":12345,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes1 := []byte(`{"metricset":{"tags":{"code":200,"success":true},"transaction":{"type":"request","name":"GET/"}, "timestamp":1571657444929001}}`)

	metadataBytes2 := []byte(`{"metadata":{"process":{"pid":54321,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes2 := []byte(`{"metricset":{"tags":{"code":200,"success":true},"transaction":{"type":"request","name":"POST/"}, "timestamp":1571657440929001}}`)

	buffer := bytes.NewBuffer(metadataBytes1)
	buffer.Write(eventBytes1)
	buffer.Write(metadataBytes2)
	buffer.Write(eventBytes2)

	resp, err := http.Post(createURL(server, "http", "/intake/v2/events", ""), "application/x-ndjson", buffer)
	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 2, len(acc.Metrics))
	require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
	require.Equal(t, tags1, acc.Metrics[0].Tags)
	require.Equal(t, fields1, acc.Metrics[0].Fields)
	require.Equal(t, "2019-10-21 11:30:44.929001 +0000 UTC", acc.Metrics[0].Time.String())
	require.Equal(t, "apm_server", acc.Metrics[1].Measurement)
	require.Equal(t, tags2, acc.Metrics[1].Tags)
	require.Equal(t, fields2, acc.Metrics[1].Fields)
	require.Equal(t, "2019-10-21 11:30:40.929001 +0000 UTC", acc.Metrics[1].Time.String())
	defer resp.Body.Close()
}

func TestEventsIntakeWithoutTimestamp(t *testing.T) {
	now := time.Now().UTC()
	tags := map[string]string{"process.pid": "12345", "process.ppid": "1", "process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags.code": 200.0, "tags.success": true, "transaction.name": "GET/", "transaction.type": "request"}

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	metadataBytes := []byte(`{"metadata":{"process":{"pid":12345,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes := []byte(`{"metricset":{"tags":{"code":200,"success":true},"transaction":{"type":"request","name":"GET/"}}}`)

	buffer := bytes.NewBuffer(metadataBytes)
	buffer.Write(eventBytes)

	resp, err := http.Post(createURL(server, "http", "/intake/v2/events", ""), "application/x-ndjson", buffer)
	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
	require.Equal(t, tags, acc.Metrics[0].Tags)
	require.Equal(t, fields, acc.Metrics[0].Fields)
	require.NotNil(t, acc.Metrics[0].Time)
	require.True(t, acc.Metrics[0].Time.After(now))
	defer resp.Body.Close()
}

func TestEventsIntakeNotValidTimestamp(t *testing.T) {

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	metadataBytes := []byte(`{"metadata":{"process":{"pid":12345,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes := []byte(`{"metricset":{"tags":{"code":200,"success":true},"transaction":{"type":"request","name":"GET/"},"timestamp":"xyz"}}`)

	buffer := bytes.NewBuffer(metadataBytes)
	buffer.Write(eventBytes)

	resp, err := http.Post(createURL(server, "http", "/intake/v2/events", ""), "application/x-ndjson", buffer)
	require.NoError(t, err)
	require.EqualValues(t, 400, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"error\":\"cannot parse timestamp: 'xyz'\"}", string(body))
}

func TestEventsIntakeGzip(t *testing.T) {

	tags := map[string]string{"process.pid": "12345", "process.ppid": "1", "process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags.code": 200.0, "tags.success": true, "transaction.name": "GET/", "transaction.type": "request"}

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	metadataBytes := []byte(`{"metadata":{"process":{"pid":12345,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes := []byte(`{"metricset":{"tags":{"code":200,"success":true},"transaction":{"type":"request","name":"GET/"}, "timestamp":1571657444929001}}`)

	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	_, err := w.Write(metadataBytes)
	require.NoError(t, err)
	_, err = w.Write(eventBytes)
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	request, err := http.NewRequest("POST", createURL(server, "http", "/intake/v2/events", ""), &buffer)
	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/x-ndjson")
	request.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
	require.Equal(t, tags, acc.Metrics[0].Tags)
	require.Equal(t, fields, acc.Metrics[0].Fields)
	require.Equal(t, "2019-10-21 11:30:44.929001 +0000 UTC", acc.Metrics[0].Time.String())

	defer resp.Body.Close()
}

func TestEventsIntakeZlib(t *testing.T) {

	tags := map[string]string{"process.pid": "102030", "process.ppid": "1", "process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags.code": 202.0, "tags.success": true, "transaction.name": "GET/", "transaction.type": "request"}

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	metadataBytes := []byte(`{"metadata":{"process":{"pid":102030,"title":"/usr/lib/bin/java","ppid":1}}}`)
	eventBytes := []byte(`{"metricset":{"tags":{"code":202,"success":true},"transaction":{"type":"request","name":"GET/"}, "timestamp":1571657444929001}}`)

	var buffer bytes.Buffer
	w := zlib.NewWriter(&buffer)
	_, err := w.Write(metadataBytes)
	require.NoError(t, err)
	_, err = w.Write(eventBytes)
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	request, err := http.NewRequest("POST", createURL(server, "http", "/intake/v2/events", ""), &buffer)
	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/x-ndjson")
	request.Header.Set("Content-Encoding", "deflate")
	resp, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
	require.Equal(t, tags, acc.Metrics[0].Tags)
	require.Equal(t, fields, acc.Metrics[0].Fields)
	require.Equal(t, "2019-10-21 11:30:44.929001 +0000 UTC", acc.Metrics[0].Time.String())

	defer resp.Body.Close()
}

func TestRUMEventsIntakePreflight(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get RUM agent configuration
	request, err := http.NewRequest("OPTIONS", createURL(server, "http", "/intake/v2/rum/events", ""), nil)
	require.NoError(t, err)

	request.Header.Set("Origin", "https://foo.example")
	resp, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)
	require.Equal(t, "https://foo.example", resp.Header.Get("Access-Control-Allow-Origin"))
	require.Equal(t, "POST, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	require.Equal(t, "Content-Type, Content-Encoding, Accept", resp.Header.Get("Access-Control-Allow-Headers"))
	require.Equal(t, "Etag", resp.Header.Get("Access-Control-Expose-Headers"))
	require.Equal(t, "86400", resp.Header.Get("Access-Control-Max-Age"))
	require.Equal(t, "Origin", resp.Header.Get("Vary"))
	require.Equal(t, 0, len(acc.Metrics))
	defer resp.Body.Close()
}

func TestRUMEventsIntake(t *testing.T) {

	now := time.Now().UTC()

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	rumBytes, _ := ioutil.ReadFile(fmt.Sprintf("./testdata/rum.ndjson"))
	buffer := bytes.NewBuffer(rumBytes)

	resp, err := http.Post(createURL(server, "http", "/intake/v2/rum/events", ""), "application/x-ndjson", buffer)
	require.NoError(t, err)
	require.EqualValues(t, 202, resp.StatusCode)
	require.Equal(t, 4, len(acc.Metrics))
	// 1
	require.Equal(t, "apm_server", acc.Metrics[0].Measurement)
	require.Equal(t, map[string]string{"service.agent.name": "rum-js", "service.agent.version": "5.0.0", "service.language.name": "javascript", "service.name": "DemoRails-RUM", "type": "transaction"}, acc.Metrics[0].Tags)
	require.Equal(t, map[string]interface{}{"context.page.referer": "", "context.page.url": "http://localhost:3000/", "context.response.decoded_body_size": 943.0, "context.response.encoded_body_size": 943.0, "context.response.transfer_size": 1863.0, "duration": 159.20999998343177, "id": "bffcb5c637da7831", "marks.agent.domComplete": 155.0, "marks.agent.domInteractive": 148.0, "marks.agent.firstContentfulPaint": 153.28499997546896, "marks.agent.largestContentfulPaint": 156.28499997546896, "marks.agent.timeToFirstByte": 69.0, "marks.navigationTiming.connectEnd": 6.0, "marks.navigationTiming.connectStart": 6.0, "marks.navigationTiming.domComplete": 155.0, "marks.navigationTiming.domContentLoadedEventEnd": 149.0, "marks.navigationTiming.domContentLoadedEventStart": 148.0, "marks.navigationTiming.domInteractive": 148.0, "marks.navigationTiming.domLoading": 86.0, "marks.navigationTiming.domainLookupEnd": 6.0, "marks.navigationTiming.domainLookupStart": 6.0, "marks.navigationTiming.fetchStart": 0.0, "marks.navigationTiming.loadEventEnd": 156.0, "marks.navigationTiming.loadEventStart": 155.0, "marks.navigationTiming.requestStart": 6.0, "marks.navigationTiming.responseEnd": 70.0, "marks.navigationTiming.responseStart": 69.0, "name": "Unknown", "sampled": true, "span_count.started": 6.0, "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "type": "page-load"}, acc.Metrics[0].Fields)
	require.NotNil(t, acc.Metrics[0].Time)
	require.True(t, acc.Metrics[0].Time.After(now))
	// 2
	require.Equal(t, "apm_server", acc.Metrics[1].Measurement)
	require.Equal(t, map[string]string{"service.agent.name": "rum-js", "service.agent.version": "5.0.0", "service.language.name": "javascript", "service.name": "DemoRails-RUM", "type": "span"}, acc.Metrics[1].Tags)
	require.Equal(t, map[string]interface{}{"duration": 64.0, "id": "665a2d250689f05d", "name": "Requesting and receiving the document", "parent_id": "bffcb5c637da7831", "start": 6.0, "subType": "browser-timing", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "hard-navigation"}, acc.Metrics[1].Fields)
	require.True(t, acc.Metrics[1].Time.After(now))
	// 3
	require.Equal(t, "apm_server", acc.Metrics[2].Measurement)
	require.Equal(t, map[string]string{"service.agent.name": "rum-js", "service.agent.version": "5.0.0", "service.language.name": "javascript", "service.name": "DemoRails-RUM", "type": "span"}, acc.Metrics[2].Tags)
	require.Equal(t, map[string]interface{}{"duration": 62.0, "id": "aae7e219d6b5be7a", "name": "Parsing the document, executing sync. scripts", "parent_id": "bffcb5c637da7831", "start": 86.0, "subType": "browser-timing", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "hard-navigation"}, acc.Metrics[2].Fields)
	require.True(t, acc.Metrics[2].Time.After(now))
	// 4
	require.Equal(t, "apm_server", acc.Metrics[3].Measurement)
	require.Equal(t, map[string]string{"service.agent.name": "rum-js", "service.agent.version": "5.0.0", "service.language.name": "javascript", "service.name": "DemoRails-RUM", "type": "span"}, acc.Metrics[3].Tags)
	require.Equal(t, map[string]interface{}{"context.destination.address": "localhost", "context.destination.port": 3000.0, "context.destination.service.name": "http://localhost:3000", "context.destination.service.resource": "localhost:3000", "context.destination.service.type": "resource", "context.http.response.decoded_body_size": 785.0, "context.http.response.encoded_body_size": 785.0, "context.http.response.transfer_size": 0.0, "context.http.url": "http://localhost:3000/assets/application.debug-8f0ab06df214da85f20badd5140ad9071c25a2186b569d896dbf0f00ebbd5acd.css", "duration": 8.354999998118728, "id": "9e0b3728b470a522", "name": "http://localhost:3000/assets/application.debug-8f0ab06df214da85f20badd5140ad9071c25a2186b569d896dbf0f00ebbd5acd.css", "parent_id": "bffcb5c637da7831", "start": 103.52499998407438, "subType": "link", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "resource"}, acc.Metrics[3].Fields)
	require.True(t, acc.Metrics[3].Time.After(now))
	defer resp.Body.Close()
}

func createURL(server *APMServer, scheme string, path string, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(server.port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}
