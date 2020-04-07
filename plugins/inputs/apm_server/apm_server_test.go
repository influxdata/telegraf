package apm_server

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
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
		"\"build_sha\":\"bc4d9a286a65b4283c2462404add86a26be61dca\",\"version\":\"7.6.0\"}", string(body))
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
			tags:     map[string]string{"labels_ab_testing": "true", "labels_group": "experimental", "labels_segment": "5", "process_argv_0": "-v", "process_pid": "1234", "process_ppid": "1", "process_title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service_agent_name": "java", "service_agent_version": "1.10.0", "service_environment": "production", "service_framework_name": "spring", "service_framework_version": "5.0.0", "service_language_name": "Java", "service_language_version": "10.0.2", "service_name": "1234_service-12a3", "service_node_configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service_runtime_name": "Java", "service_runtime_version": "10.0.2", "service_version": "4.3.0", "system_architecture": "amd64", "system_configured_hostname": "host1", "system_container_id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system_detected_hostname": "8ec7ceb99074", "system_kubernetes_namespace": "default", "system_kubernetes_node_name": "node-name", "system_kubernetes_pod_name": "instrumented-java-service", "system_kubernetes_pod_uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system_platform": "Linux", "type": "metricset"},
			fields:   map[string]interface{}{"samples_byte_counter_value": 1.0, "samples_dotted.float.gauge_value": 6.12, "samples_double_gauge_value": 3.141592653589793, "samples_float_gauge_value": 9.16, "samples_integer_gauge_value": 42767.0, "samples_long_gauge_value": 3.147483648e+09, "samples_negative.d.o.t.t.e.d_value": -1022.0, "samples_short_counter_value": 227.0, "samples_span.self_time.count_value": 1.0, "samples_span.self_time.sum.us_value": 633.288, "samples_transaction.breakdown.count_value": 12.0, "samples_transaction.duration.count_value": 2.0, "samples_transaction.duration.sum.us_value": 12.0, "samples_transaction.self_time.count_value": 2.0, "samples_transaction.self_time.sum.us_value": 10.0, "span_subtype": "mysql", "span_type": "db", "tags_code": 200.0, "tags_success": true, "transaction_name": "GET/", "transaction_type": "request"},
		},
		{
			name:     "transaction mapping",
			metadata: "metadata.ndjson",
			event:    "transaction.ndjson",
			tags:     map[string]string{"labels_ab_testing": "true", "labels_group": "experimental", "labels_segment": "5", "process_argv_0": "-v", "process_pid": "1234", "process_ppid": "1", "process_title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service_agent_name": "java", "service_agent_version": "1.10.0", "service_environment": "production", "service_framework_name": "spring", "service_framework_version": "5.0.0", "service_language_name": "Java", "service_language_version": "10.0.2", "service_name": "1234_service-12a3", "service_node_configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service_runtime_name": "Java", "service_runtime_version": "10.0.2", "service_version": "4.3.0", "system_architecture": "amd64", "system_configured_hostname": "host1", "system_container_id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system_detected_hostname": "8ec7ceb99074", "system_kubernetes_namespace": "default", "system_kubernetes_node_name": "node-name", "system_kubernetes_pod_name": "instrumented-java-service", "system_kubernetes_pod_uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system_platform": "Linux", "type": "transaction"},
			fields:   map[string]interface{}{"context_custom_(": "notavalidregexandthatisfine", "context_custom_and_objects_foo_0": "bar", "context_custom_and_objects_foo_1": "baz", "context_custom_my_key": 1.0, "context_custom_some_other_value": "foobar", "context_request_body_additional_bar": 123.0, "context_request_body_additional_req": "additionalinformation", "context_request_body_string": "helloworld", "context_request_cookies_c1": "v1", "context_request_cookies_c2": "v2", "context_request_env_GATEWAY_INTERFACE": "CGI/1.1", "context_request_env_SERVER_SOFTWARE": "nginx", "context_request_headers_Elastic-Apm-Traceparent_0": "00-33a0bd4cceff0370a7c57d807032688e-69feaabc5b88d7e8-01", "context_request_headers_content-type": "text/html", "context_request_headers_cookie": "c1=v1,c2=v2", "context_request_headers_user-agent_0": "Mozilla/5.0(Macintosh;IntelMacOSX10_10_5)AppleWebKit/537.36(KHTML,likeGecko)Chrome/51.0.2704.103Safari/537.36", "context_request_headers_user-agent_1": "MozillaChromeEdge", "context_request_http_version": "1.1", "context_request_method": "POST", "context_request_socket_encrypted": true, "context_request_socket_remote_address": "12.53.12.1:8080", "context_request_url_full": "https://www.example.com/p/a/t/h?query=string#hash", "context_request_url_hash": "#hash", "context_request_url_hostname": "www.example.com", "context_request_url_pathname": "/p/a/t/h", "context_request_url_port": "8080", "context_request_url_protocol": "https:", "context_request_url_raw": "/p/a/t/h?query=string#hash", "context_request_url_search": "?query=string", "context_response_decoded_body_size": 401.9, "context_response_encoded_body_size": 356.9, "context_response_finished": true, "context_response_headers_content-type": "application/json", "context_response_headers_sent": true, "context_response_status_code": 200.0, "context_response_transfer_size": 300.0, "context_service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "context_service_agent_version": "1.10.0-SNAPSHOT", "context_service_name": "experimental-java", "context_tags_organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "context_user_email": "foo@mail.com", "context_user_id": "99", "context_user_username": "foo", "duration": 32.592981, "id": "4340a8e0df1906ecbfa9", "name": "ResourceHttpRequestHandler", "parent_id": "abcdefabcdef01234567", "result": "HTTP2xx", "sampled": true, "span_count_dropped": 0.0, "span_count_started": 17.0, "trace_id": "0acd456789abcdef0123456789abcdef", "type": "http"},
		},
		{
			name:     "span mapping",
			metadata: "metadata.ndjson",
			event:    "span.ndjson",
			tags:     map[string]string{"labels_ab_testing": "true", "labels_group": "experimental", "labels_segment": "5", "process_argv_0": "-v", "process_pid": "1234", "process_ppid": "1", "process_title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service_agent_name": "java", "service_agent_version": "1.10.0", "service_environment": "production", "service_framework_name": "spring", "service_framework_version": "5.0.0", "service_language_name": "Java", "service_language_version": "10.0.2", "service_name": "1234_service-12a3", "service_node_configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service_runtime_name": "Java", "service_runtime_version": "10.0.2", "service_version": "4.3.0", "system_architecture": "amd64", "system_configured_hostname": "host1", "system_container_id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system_detected_hostname": "8ec7ceb99074", "system_kubernetes_namespace": "default", "system_kubernetes_node_name": "node-name", "system_kubernetes_pod_name": "instrumented-java-service", "system_kubernetes_pod_uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system_platform": "Linux", "type": "span"},
			fields:   map[string]interface{}{"action": "connect", "context_db_instance": "customers", "context_db_link": "other.db.com", "context_db_statement": "SELECT * FROM product_types WHERE user_id = ?", "context_db_type": "sql", "context_db_user": "postgres", "context_http_method": "GET", "context_http_response_decoded_body_size": 401.0, "context_http_response_encoded_body_size": 356.0, "context_http_response_headers_content-type": "application/json", "context_http_response_status_code": 200.0, "context_http_response_transfer_size": 300.12, "context_http_status_code": 302.0, "context_http_url": "http://localhost:8000", "context_service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "context_service_agent_name": "java", "context_service_agent_version": "1.10.0-SNAPSHOT", "context_service_name": "opbeans-java-1", "duration": 3.781912, "id": "1234567890aaaade", "name": "GET users-authenticated", "parent_id": "abcdef0123456789", "stacktrace_0_filename": "DispatcherServlet.java", "stacktrace_0_lineno": 547.0, "stacktrace_1_abs_path": "/tmp/AbstractView.java", "stacktrace_1_colno": 4.0, "stacktrace_1_context_line": "line3", "stacktrace_1_filename": "AbstractView.java", "stacktrace_1_function": "render", "stacktrace_1_library_frame": true, "stacktrace_1_lineno": 547.0, "stacktrace_1_module": "org.springframework.web.servlet.view", "stacktrace_1_vars_key": "value", "subtype": "http", "sync": true, "trace_id": "abcdef0123456789abcdef9876543210", "transaction_id": "1234567890987654", "type": "external"},
		},
		{
			name:     "error mapping",
			metadata: "metadata.ndjson",
			event:    "error.ndjson",
			tags:     map[string]string{"labels_ab_testing": "true", "labels_group": "experimental", "labels_segment": "5", "process_argv_0": "-v", "process_pid": "1234", "process_ppid": "1", "process_title": "/usr/lib/jvm/java-10-openjdk-amd64/bin/java", "service_agent_ephemeral_id": "e71be9ac-93b0-44b9-a997-5638f6ccfc36", "service_agent_name": "java", "service_agent_version": "1.10.0", "service_environment": "production", "service_framework_name": "spring", "service_framework_version": "5.0.0", "service_language_name": "Java", "service_language_version": "10.0.2", "service_name": "1234_service-12a3", "service_node_configured_name": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "service_runtime_name": "Java", "service_runtime_version": "10.0.2", "service_version": "4.3.0", "system_architecture": "amd64", "system_configured_hostname": "host1", "system_container_id": "8ec7ceb990749e79b37f6dc6cd3628633618d6ce412553a552a0fa6b69419ad4", "system_detected_hostname": "8ec7ceb99074", "system_kubernetes_namespace": "default", "system_kubernetes_node_name": "node-name", "system_kubernetes_pod_name": "instrumented-java-service", "system_kubernetes_pod_uid": "b17f231da0ad128dc6c6c0b2e82f6f303d3893e3", "system_platform": "Linux", "type": "error"},
			fields:   map[string]interface{}{"context_custom_and_objects_foo_0": "bar", "context_custom_and_objects_foo_1": "baz", "context_custom_my_key": 1.0, "context_custom_some_other_value": "foobar", "context_request_body": "HelloWorld", "context_request_cookies_c1": "v1", "context_request_cookies_c2": "v2", "context_request_env_GATEWAY_INTERFACE": "CGI/1.1", "context_request_env_SERVER_SOFTWARE": "nginx", "context_request_headers_Elastic-Apm-Traceparent": "00-8c21b4b556467a0b17ae5da959b5f388-31301f1fb2998121-01", "context_request_headers_Forwarded": "for=192.168.0.1", "context_request_headers_content-length": "0", "context_request_headers_cookie_0": "c1=v1", "context_request_headers_cookie_1": "c2=v2", "context_request_headers_host": "opbeans-java:3000", "context_request_http_version": "1.1", "context_request_method": "POST", "context_request_socket_encrypted": true, "context_request_socket_remote_address": "12.53.12.1", "context_request_url_full": "https://www.example.com/p/a/t/h?query=string#hash", "context_request_url_hash": "#hash", "context_request_url_hostname": "www.example.com", "context_request_url_pathname": "/p/a/t/h", "context_request_url_port": 8080.0, "context_request_url_protocol": "https:", "context_request_url_raw": "/p/a/t/h?query=string#hash", "context_request_url_search": "?query=string", "context_response_finished": true, "context_response_headers_content-type": "application/json", "context_response_headers_sent": true, "context_response_status_code": 200.0, "context_service_framework_name": "Node", "context_service_framework_version": "1", "context_service_language_version": "1.2", "context_service_name": "service1", "context_service_node_configured_name": "node-xyz", "context_tags_organization_uuid": "9f0e9d64-c185-4d21-a6f4-4673ed561ec8", "context_user_email": "user@foo.mail", "context_user_id": 99.0, "context_user_username": "foo", "culprit": "opbeans.controllers.DTInterceptor.preHandle(DTInterceptor.java:73)", "exception_attributes_foo": "bar", "exception_cause_0_cause_0_message": "disk spinning way too fast", "exception_cause_0_cause_0_type": "VeryInternalDbError", "exception_cause_0_cause_1_message": "on top of it,internet doesn't work", "exception_cause_0_cause_1_type": "ConnectionError", "exception_cause_0_message": "something wrong writing a file", "exception_cause_0_type": "InternalDbError", "exception_code": 42.0, "exception_handled": false, "exception_message": "Theusernamerootisunknown", "exception_module": "org.springframework.http.client", "exception_stacktrace_0_abs_path": "/tmp/AbstractPlainSocketImpl.java", "exception_stacktrace_0_colno": 4.0, "exception_stacktrace_0_context_line": "3", "exception_stacktrace_0_filename": "AbstractPlainSocketImpl.java", "exception_stacktrace_0_function": "connect", "exception_stacktrace_0_library_frame": true, "exception_stacktrace_0_lineno": 3.0, "exception_stacktrace_0_module": "java.net", "exception_stacktrace_0_post_context_0": "line4", "exception_stacktrace_0_post_context_1": "line5", "exception_stacktrace_0_pre_context_0": "line1", "exception_stacktrace_0_pre_context_1": "line2", "exception_stacktrace_0_vars_key": "value", "exception_stacktrace_1_filename": "AbstractClientHttpRequest.java", "exception_stacktrace_1_function": "execute", "exception_stacktrace_1_lineno": 102.0, "exception_stacktrace_1_vars_key": "value", "exception_type": "java.net.UnknownHostException", "id": "9876543210abcdeffedcba0123456789", "log_level": "error", "log_logger_name": "http404", "log_message": "Request method 'POST' not supported", "log_param_message": "Request method 'POST' /events/:event not supported", "log_stacktrace_0_abs_path": "/tmp/Socket.java", "log_stacktrace_0_classname": "Request::Socket", "log_stacktrace_0_colno": 4.0, "log_stacktrace_0_context_line": "line3", "log_stacktrace_0_filename": "Socket.java", "log_stacktrace_0_function": "connect", "log_stacktrace_0_library_frame": true, "log_stacktrace_0_lineno": 3.0, "log_stacktrace_0_module": "java.net", "log_stacktrace_0_post_context_0": "line4", "log_stacktrace_0_post_context_1": "line5", "log_stacktrace_0_pre_context_0": "line1", "log_stacktrace_0_pre_context_1": "line2", "log_stacktrace_0_vars_key": "value", "log_stacktrace_1_abs_path": "/tmp/SimpleBufferingClientHttpRequest.java", "log_stacktrace_1_filename": "SimpleBufferingClientHttpRequest.java", "log_stacktrace_1_function": "executeInternal", "log_stacktrace_1_lineno": 102.0, "log_stacktrace_1_vars_key": "value", "parent_id": "9632587410abcdef", "trace_id": "0123456789abcdeffedcba0123456789", "transaction_id": "1234567890987654", "transaction_sampled": true, "transaction_type": "request"},
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
	tags1 := map[string]string{"process_pid": "12345", "process_ppid": "1", "process_title": "/usr/lib/bin/java", "type": "metricset"}
	fields1 := map[string]interface{}{"tags_code": 200.0, "tags_success": true, "transaction_name": "GET/", "transaction_type": "request"}
	tags2 := map[string]string{"process_pid": "54321", "process_ppid": "1", "process_title": "/usr/lib/bin/java", "type": "metricset"}
	fields2 := map[string]interface{}{"tags_code": 200.0, "tags_success": true, "transaction_name": "POST/", "transaction_type": "request"}

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
	tags := map[string]string{"process_pid": "12345", "process_ppid": "1", "process_title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags_code": 200.0, "tags_success": true, "transaction_name": "GET/", "transaction_type": "request"}

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

	tags := map[string]string{"process_pid": "12345", "process_ppid": "1", "process_title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags_code": 200.0, "tags_success": true, "transaction_name": "GET/", "transaction_type": "request"}

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

	tags := map[string]string{"process_pid": "102030", "process_ppid": "1", "process_title": "/usr/lib/bin/java", "type": "metricset"}
	fields := map[string]interface{}{"tags_code": 202.0, "tags_success": true, "transaction_name": "GET/", "transaction_type": "request"}

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
	require.Equal(t, map[string]string{"service_agent_name": "rum-js", "service_agent_version": "5.0.0", "service_language_name": "javascript", "service_name": "DemoRails-RUM", "type": "transaction"}, acc.Metrics[0].Tags)
	require.Equal(t, map[string]interface{}{"context_page_referer": "", "context_page_url": "http://localhost:3000/", "context_response_decoded_body_size": 943.0, "context_response_encoded_body_size": 943.0, "context_response_transfer_size": 1863.0, "duration": 159.20999998343177, "id": "bffcb5c637da7831", "marks_agent_domComplete": 155.0, "marks_agent_domInteractive": 148.0, "marks_agent_firstContentfulPaint": 153.28499997546896, "marks_agent_largestContentfulPaint": 156.28499997546896, "marks_agent_timeToFirstByte": 69.0, "marks_navigationTiming_connectEnd": 6.0, "marks_navigationTiming_connectStart": 6.0, "marks_navigationTiming_domComplete": 155.0, "marks_navigationTiming_domContentLoadedEventEnd": 149.0, "marks_navigationTiming_domContentLoadedEventStart": 148.0, "marks_navigationTiming_domInteractive": 148.0, "marks_navigationTiming_domLoading": 86.0, "marks_navigationTiming_domainLookupEnd": 6.0, "marks_navigationTiming_domainLookupStart": 6.0, "marks_navigationTiming_fetchStart": 0.0, "marks_navigationTiming_loadEventEnd": 156.0, "marks_navigationTiming_loadEventStart": 155.0, "marks_navigationTiming_requestStart": 6.0, "marks_navigationTiming_responseEnd": 70.0, "marks_navigationTiming_responseStart": 69.0, "name": "Unknown", "sampled": true, "span_count_started": 6.0, "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "type": "page-load"}, acc.Metrics[0].Fields)
	require.NotNil(t, acc.Metrics[0].Time)
	require.True(t, acc.Metrics[0].Time.After(now))
	// 2
	require.Equal(t, "apm_server", acc.Metrics[1].Measurement)
	require.Equal(t, map[string]string{"service_agent_name": "rum-js", "service_agent_version": "5.0.0", "service_language_name": "javascript", "service_name": "DemoRails-RUM", "type": "span"}, acc.Metrics[1].Tags)
	require.Equal(t, map[string]interface{}{"duration": 64.0, "id": "665a2d250689f05d", "name": "Requesting and receiving the document", "parent_id": "bffcb5c637da7831", "start": 6.0, "subType": "browser-timing", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "hard-navigation"}, acc.Metrics[1].Fields)
	require.True(t, acc.Metrics[1].Time.After(now))
	// 3
	require.Equal(t, "apm_server", acc.Metrics[2].Measurement)
	require.Equal(t, map[string]string{"service_agent_name": "rum-js", "service_agent_version": "5.0.0", "service_language_name": "javascript", "service_name": "DemoRails-RUM", "type": "span"}, acc.Metrics[2].Tags)
	require.Equal(t, map[string]interface{}{"duration": 62.0, "id": "aae7e219d6b5be7a", "name": "Parsing the document, executing sync. scripts", "parent_id": "bffcb5c637da7831", "start": 86.0, "subType": "browser-timing", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "hard-navigation"}, acc.Metrics[2].Fields)
	require.True(t, acc.Metrics[2].Time.After(now))
	// 4
	require.Equal(t, "apm_server", acc.Metrics[3].Measurement)
	require.Equal(t, map[string]string{"service_agent_name": "rum-js", "service_agent_version": "5.0.0", "service_language_name": "javascript", "service_name": "DemoRails-RUM", "type": "span"}, acc.Metrics[3].Tags)
	require.Equal(t, map[string]interface{}{"context_destination_address": "localhost", "context_destination_port": 3000.0, "context_destination_service_name": "http://localhost:3000", "context_destination_service_resource": "localhost:3000", "context_destination_service_type": "resource", "context_http_response_decoded_body_size": 785.0, "context_http_response_encoded_body_size": 785.0, "context_http_response_transfer_size": 0.0, "context_http_url": "http://localhost:3000/assets/application.debug-8f0ab06df214da85f20badd5140ad9071c25a2186b569d896dbf0f00ebbd5acd.css", "duration": 8.354999998118728, "id": "9e0b3728b470a522", "name": "http://localhost:3000/assets/application.debug-8f0ab06df214da85f20badd5140ad9071c25a2186b569d896dbf0f00ebbd5acd.css", "parent_id": "bffcb5c637da7831", "start": 103.52499998407438, "subType": "link", "trace_id": "3a3bd994744f26c4ca7ae74332e04a76", "transaction_id": "bffcb5c637da7831", "type": "resource"}, acc.Metrics[3].Fields)
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
