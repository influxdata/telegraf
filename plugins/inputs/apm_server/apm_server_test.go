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
	resp, err := http.Get(createURL(server, "http", "/config/v1/rum/agents", ""))
	require.NoError(t, err)
	require.EqualValues(t, 403, resp.StatusCode)
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
			tags:     map[string]string{"metadata.labels.ab_testing": "true", "metadata.labels.group": "experimental", "metadata.labels.segment": "5", "metadata.process.argv.0": "-v", "metadata.process.pid": "1234", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "metadata.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "metadata.service.agent.name": "java", "metadata.service.agent.version": "1.10.0", "metadata.service.environment": "production", "metadata.service.framework.name": "spring", "metadata.service.framework.version": "5.0.0", "metadata.service.language.name": "Java", "metadata.service.language.version": "10.0.2", "metadata.service.name": "1234_service-12a3", "metadata.service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.service.runtime.name": "Java", "metadata.service.runtime.version": "10.0.2", "metadata.service.version": "4.3.0", "metadata.system.architecture": "amd64", "metadata.system.configured_hostname": "host1", "metadata.system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.system.detected_hostname": "8ec7ceb99074", "metadata.system.kubernetes.namespace": "default", "metadata.system.kubernetes.node.name": "node-name", "metadata.system.kubernetes.pod.name": "instrumented-java-service", "metadata.system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "metadata.system.platform": "Linux", "type": "metricset"},
			fields:   map[string]interface{}{"metricset.samples.byte_counter.value": 1.0, "metricset.samples.dotted.float.gauge.value": 6.12, "metricset.samples.double_gauge.value": 3.141592653589793, "metricset.samples.float_gauge.value": 9.16, "metricset.samples.integer_gauge.value": 42767.0, "metricset.samples.long_gauge.value": 3.147483648e+09, "metricset.samples.negative.d.o.t.t.e.d.value": -1022.0, "metricset.samples.short_counter.value": 227.0, "metricset.samples.span.self_time.count.value": 1.0, "metricset.samples.span.self_time.sum.us.value": 633.288, "metricset.samples.transaction.breakdown.count.value": 12.0, "metricset.samples.transaction.duration.count.value": 2.0, "metricset.samples.transaction.duration.sum.us.value": 12.0, "metricset.samples.transaction.self_time.count.value": 2.0, "metricset.samples.transaction.self_time.sum.us.value": 10.0, "metricset.span.subtype": "mysql", "metricset.span.type": "db", "metricset.tags.code": 200.0, "metricset.tags.success": true, "metricset.transaction.name": "GET/", "metricset.transaction.type": "request"},
		},
		{
			name:     "transaction mapping",
			metadata: "metadata.ndjson",
			event:    "transaction.ndjson",
			tags:     map[string]string{"metadata.labels.ab_testing": "true", "metadata.labels.group": "experimental", "metadata.labels.segment": "5", "metadata.process.argv.0": "-v", "metadata.process.pid": "1234", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "metadata.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "metadata.service.agent.name": "java", "metadata.service.agent.version": "1.10.0", "metadata.service.environment": "production", "metadata.service.framework.name": "spring", "metadata.service.framework.version": "5.0.0", "metadata.service.language.name": "Java", "metadata.service.language.version": "10.0.2", "metadata.service.name": "1234_service-12a3", "metadata.service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.service.runtime.name": "Java", "metadata.service.runtime.version": "10.0.2", "metadata.service.version": "4.3.0", "metadata.system.architecture": "amd64", "metadata.system.configured_hostname": "host1", "metadata.system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.system.detected_hostname": "8ec7ceb99074", "metadata.system.kubernetes.namespace": "default", "metadata.system.kubernetes.node.name": "node-name", "metadata.system.kubernetes.pod.name": "instrumented-java-service", "metadata.system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "metadata.system.platform": "Linux", "type": "transaction"},
			fields:   map[string]interface{}{"transaction.context.custom.(": "notavalidregexandthatisfine", "transaction.context.custom.and_objects.foo.0": "bar", "transaction.context.custom.and_objects.foo.1": "baz", "transaction.context.custom.my_key": 1.0, "transaction.context.custom.some_other_value": "foobar", "transaction.context.request.body.additional.bar": 123.0, "transaction.context.request.body.additional.req": "additionalinformation", "transaction.context.request.body.string": "helloworld", "transaction.context.request.cookies.c1": "v1", "transaction.context.request.cookies.c2": "v2", "transaction.context.request.env.GATEWAY_INTERFACE": "CGI/1.1", "transaction.context.request.env.SERVER_SOFTWARE": "nginx", "transaction.context.request.headers.Elastic-Apm-Traceparent.0": "00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01", "transaction.context.request.headers.content-type": "text/html", "transaction.context.request.headers.cookie": "c1=v1,c2=v2", "transaction.context.request.headers.user-agent.0": "Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36", "transaction.context.request.headers.user-agent.1": "MozillaChromeEdge", "transaction.context.request.http_version": "1.1", "transaction.context.request.method": "POST", "transaction.context.request.socket.encrypted": true, "transaction.context.request.socket.remote_address": "12.53.12.1:8080", "transaction.context.request.url.full": "https://www.example.com/p/a/t/h?query=string#hash", "transaction.context.request.url.hash": "#hash", "transaction.context.request.url.hostname": "www.example.com", "transaction.context.request.url.pathname": "/p/a/t/h", "transaction.context.request.url.port": "8080", "transaction.context.request.url.protocol": "https:", "transaction.context.request.url.raw": "/p/a/t/h?query=string#hash", "transaction.context.request.url.search": "?query=string", "transaction.context.response.decoded_body_size": 401.9, "transaction.context.response.encoded_body_size": 356.9, "transaction.context.response.finished": true, "transaction.context.response.headers.content-type": "application/json", "transaction.context.response.headers_sent": true, "transaction.context.response.status_code": 200.0, "transaction.context.response.transfer_size": 300.0, "transaction.context.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "transaction.context.service.agent.version": "1.10.0-SNAPSHOT", "transaction.context.service.name": "experimental-java", "transaction.context.tags.organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "transaction.context.user.email": "foo@mail.com", "transaction.context.user.id": "99", "transaction.context.user.username": "foo", "transaction.duration": 32.592981, "transaction.id": "4340a8e0df1906ecbfa9", "transaction.name": "ResourceHttpRequestHandler", "transaction.parent_id": "abcdefabcdef01234567", "transaction.result": "HTTP2xx", "transaction.sampled": true, "transaction.span_count.dropped": 0.0, "transaction.span_count.started": 17.0, "transaction.trace_id": "0acd456789abcdef0123456789abcdef", "transaction.type": "http"},
		},
		{
			name:     "span mapping",
			metadata: "metadata.ndjson",
			event:    "span.ndjson",
			tags:     map[string]string{"metadata.labels.ab_testing": "true", "metadata.labels.group": "experimental", "metadata.labels.segment": "5", "metadata.process.argv.0": "-v", "metadata.process.pid": "1234", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "metadata.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "metadata.service.agent.name": "java", "metadata.service.agent.version": "1.10.0", "metadata.service.environment": "production", "metadata.service.framework.name": "spring", "metadata.service.framework.version": "5.0.0", "metadata.service.language.name": "Java", "metadata.service.language.version": "10.0.2", "metadata.service.name": "1234_service-12a3", "metadata.service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.service.runtime.name": "Java", "metadata.service.runtime.version": "10.0.2", "metadata.service.version": "4.3.0", "metadata.system.architecture": "amd64", "metadata.system.configured_hostname": "host1", "metadata.system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.system.detected_hostname": "8ec7ceb99074", "metadata.system.kubernetes.namespace": "default", "metadata.system.kubernetes.node.name": "node-name", "metadata.system.kubernetes.pod.name": "instrumented-java-service", "metadata.system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "metadata.system.platform": "Linux", "type": "span"},
			fields:   map[string]interface{}{"span.action": "connect", "span.context.db.instance": "customers", "span.context.db.link": "other.db.com", "span.context.db.statement": "SELECT * FROM product_types WHERE user_id = ?", "span.context.db.type": "sql", "span.context.db.user": "postgres", "span.context.http.method": "GET", "span.context.http.response.decoded_body_size": 401.0, "span.context.http.response.encoded_body_size": 356.0, "span.context.http.response.headers.content-type": "application/json", "span.context.http.response.status_code": 200.0, "span.context.http.response.transfer_size": 300.12, "span.context.http.status_code": 302.0, "span.context.http.url": "http://localhost:8000", "span.context.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "span.context.service.agent.name": "java", "span.context.service.agent.version": "1.10.0-SNAPSHOT", "span.context.service.name": "opbeans-java-1", "span.duration": 3.781912, "span.id": "1234567890aaaade", "span.name": "GET users-authenticated", "span.parent_id": "abcdef0123456789", "span.stacktrace.0.filename": "DispatcherServlet.java", "span.stacktrace.0.lineno": 547.0, "span.stacktrace.1.abs_path": "/tmp/AbstractView.java", "span.stacktrace.1.colno": 4.0, "span.stacktrace.1.context_line": "line3", "span.stacktrace.1.filename": "AbstractView.java", "span.stacktrace.1.function": "render", "span.stacktrace.1.library_frame": true, "span.stacktrace.1.lineno": 547.0, "span.stacktrace.1.module": "org.springframework.web.servlet.view", "span.stacktrace.1.vars.key": "value", "span.subtype": "http", "span.sync": true, "span.trace_id": "abcdef0123456789abcdef9876543210", "span.transaction_id": "1234567890987654", "span.type": "external"},
		},
		{
			name:     "error mapping",
			metadata: "metadata.ndjson",
			event:    "error.ndjson",
			tags:     map[string]string{"metadata.labels.ab_testing": "true", "metadata.labels.group": "experimental", "metadata.labels.segment": "5", "metadata.process.argv.0": "-v", "metadata.process.pid": "1234", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "metadata.service.agent.ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "metadata.service.agent.name": "java", "metadata.service.agent.version": "1.10.0", "metadata.service.environment": "production", "metadata.service.framework.name": "spring", "metadata.service.framework.version": "5.0.0", "metadata.service.language.name": "Java", "metadata.service.language.version": "10.0.2", "metadata.service.name": "1234_service-12a3", "metadata.service.node.configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.service.runtime.name": "Java", "metadata.service.runtime.version": "10.0.2", "metadata.service.version": "4.3.0", "metadata.system.architecture": "amd64", "metadata.system.configured_hostname": "host1", "metadata.system.container.id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "metadata.system.detected_hostname": "8ec7ceb99074", "metadata.system.kubernetes.namespace": "default", "metadata.system.kubernetes.node.name": "node-name", "metadata.system.kubernetes.pod.name": "instrumented-java-service", "metadata.system.kubernetes.pod.uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "metadata.system.platform": "Linux", "type": "error"},
			fields:   map[string]interface{}{"error.context.custom.and_objects.foo.0": "bar", "error.context.custom.and_objects.foo.1": "baz", "error.context.custom.my_key": 1.0, "error.context.custom.some_other_value": "foobar", "error.context.request.body": "HelloWorld", "error.context.request.cookies.c1": "v1", "error.context.request.cookies.c2": "v2", "error.context.request.env.GATEWAY_INTERFACE": "CGI/1.1", "error.context.request.env.SERVER_SOFTWARE": "nginx", "error.context.request.headers.Elastic-Apm-Traceparent": "00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01", "error.context.request.headers.Forwarded": "for=192.168.0.1", "error.context.request.headers.content-length": "0", "error.context.request.headers.cookie.0": "c1=v1", "error.context.request.headers.cookie.1": "c2=v2", "error.context.request.headers.host": "opbeans-java:3000", "error.context.request.http_version": "1.1", "error.context.request.method": "POST", "error.context.request.socket.encrypted": true, "error.context.request.socket.remote_address": "12.53.12.1", "error.context.request.url.full": "https://www.example.com/p/a/t/h?query=string#hash", "error.context.request.url.hash": "#hash", "error.context.request.url.hostname": "www.example.com", "error.context.request.url.pathname": "/p/a/t/h", "error.context.request.url.port": 8080.0, "error.context.request.url.protocol": "https:", "error.context.request.url.raw": "/p/a/t/h?query=string#hash", "error.context.request.url.search": "?query=string", "error.context.response.finished": true, "error.context.response.headers.content-type": "application/json", "error.context.response.headers_sent": true, "error.context.response.status_code": 200.0, "error.context.service.framework.name": "Node", "error.context.service.framework.version": "1", "error.context.service.language.version": "1.2", "error.context.service.name": "service1", "error.context.service.node.configured_name": "node-xyz", "error.context.tags.organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "error.context.user.email": "user@foo.mail", "error.context.user.id": 99.0, "error.context.user.username": "foo", "error.culprit": "opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)", "error.exception.attributes.foo": "bar", "error.exception.cause.0.cause.0.message": "disk spinning way too fast", "error.exception.cause.0.cause.0.type": "VeryInternalDbError", "error.exception.cause.0.cause.1.message": "on top of it,internet doesn't work", "error.exception.cause.0.cause.1.type": "ConnectionError", "error.exception.cause.0.message": "something wrong writing a file", "error.exception.cause.0.type": "InternalDbError", "error.exception.code": 42.0, "error.exception.handled": false, "error.exception.message": "Theusernamerootisunknown", "error.exception.module": "org.springframework.http.client", "error.exception.stacktrace.0.abs_path": "/tmp/AbstractPlainSocketImpl.java", "error.exception.stacktrace.0.colno": 4.0, "error.exception.stacktrace.0.context_line": "3", "error.exception.stacktrace.0.filename": "AbstractPlainSocketImpl.java", "error.exception.stacktrace.0.function": "connect", "error.exception.stacktrace.0.library_frame": true, "error.exception.stacktrace.0.lineno": 3.0, "error.exception.stacktrace.0.module": "java.net", "error.exception.stacktrace.0.post_context.0": "line4", "error.exception.stacktrace.0.post_context.1": "line5", "error.exception.stacktrace.0.pre_context.0": "line1", "error.exception.stacktrace.0.pre_context.1": "line2", "error.exception.stacktrace.0.vars.key": "value", "error.exception.stacktrace.1.filename": "AbstractClientHttpRequest.java", "error.exception.stacktrace.1.function": "execute", "error.exception.stacktrace.1.lineno": 102.0, "error.exception.stacktrace.1.vars.key": "value", "error.exception.type": "java.net.UnknownHostException", "error.id": "9876543210abcdeffedcba0123456789", "error.log.level": "error", "error.log.logger_name": "http404", "error.log.message": "Request method 'POST' not supported", "error.log.param_message": "Request method 'POST' /events/:event not supported", "error.log.stacktrace.0.abs_path": "/tmp/Socket.java", "error.log.stacktrace.0.classname": "Request::Socket", "error.log.stacktrace.0.colno": 4.0, "error.log.stacktrace.0.context_line": "line3", "error.log.stacktrace.0.filename": "Socket.java", "error.log.stacktrace.0.function": "connect", "error.log.stacktrace.0.library_frame": true, "error.log.stacktrace.0.lineno": 3.0, "error.log.stacktrace.0.module": "java.net", "error.log.stacktrace.0.post_context.0": "line4", "error.log.stacktrace.0.post_context.1": "line5", "error.log.stacktrace.0.pre_context.0": "line1", "error.log.stacktrace.0.pre_context.1": "line2", "error.log.stacktrace.0.vars.key": "value", "error.log.stacktrace.1.abs_path": "/tmp/SimpleBufferingClientHttpRequest.java", "error.log.stacktrace.1.filename": "SimpleBufferingClientHttpRequest.java", "error.log.stacktrace.1.function": "executeInternal", "error.log.stacktrace.1.lineno": 102.0, "error.log.stacktrace.1.vars.key": "value", "error.parent_id": "9632587410abcdef", "error.trace_id": "0123456789abcdeffedcba0123456789", "error.transaction.sampled": true, "error.transaction.type": "request", "error.transaction_id": "1234567890987654"},
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
	tags1 := map[string]string{"metadata.process.pid": "12345", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields1 := map[string]interface{}{"metricset.tags.code": 200.0, "metricset.tags.success": true, "metricset.transaction.name": "GET/", "metricset.transaction.type": "request"}
	tags2 := map[string]string{"metadata.process.pid": "54321", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields2 := map[string]interface{}{"metricset.tags.code": 200.0, "metricset.tags.success": true, "metricset.transaction.name": "POST/", "metricset.transaction.type": "request"}

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
	tags := map[string]string{"metadata.process.pid": "12345", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"metricset.tags.code": 200.0, "metricset.tags.success": true, "metricset.transaction.name": "GET/", "metricset.transaction.type": "request"}

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

	tags := map[string]string{"metadata.process.pid": "12345", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"metricset.tags.code": 200.0, "metricset.tags.success": true, "metricset.transaction.name": "GET/", "metricset.transaction.type": "request"}

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

	tags := map[string]string{"metadata.process.pid": "102030", "metadata.process.ppid": "1", "metadata.process.title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"metricset.tags.code": 202.0, "metricset.tags.success": true, "metricset.transaction.name": "GET/", "metricset.transaction.type": "request"}

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

func createURL(server *APMServer, scheme string, path string, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(server.port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}
