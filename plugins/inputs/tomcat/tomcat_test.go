package tomcat

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

var tomcatStatus8 = `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="/manager/xform.xsl" ?>
<status>
  <jvm>
    <memory free='17909336' total='58195968' max='620756992'/>
    <memorypool name='PS Eden Space' type='Heap memory' usageInit='8912896' usageCommitted='35651584' usageMax='230686720' usageUsed='25591384'/>
    <memorypool name='PS Old Gen' type='Heap memory' usageInit='21495808' usageCommitted='21495808' usageMax='465567744' usageUsed='13663040'/>
    <memorypool name='PS Survivor Space' type='Heap memory' usageInit='1048576' usageCommitted='1048576' usageMax='1048576' usageUsed='1032208'/>
    <memorypool name='Code Cache' type='Non-heap memory' usageInit='2555904' usageCommitted='2555904' usageMax='50331648' usageUsed='1220096'/>
    <memorypool name='PS Perm Gen' type='Non-heap memory' usageInit='22020096' usageCommitted='22020096' usageMax='174063616' usageUsed='17533952'/>
  </jvm>
  <connector name='"ajp-apr-8009"'>
    <threadInfo maxThreads="200" currentThreadCount="0" currentThreadsBusy="0"/>
    <requestInfo maxTime="0" processingTime="0" requestCount="0" errorCount="0" bytesReceived="0" bytesSent="0"/>
    <workers>
    </workers>
  </connector>
  <connector name='"http-apr-8080"'>
    <threadInfo maxThreads="200" currentThreadCount="5" currentThreadsBusy="1"/>
    <requestInfo maxTime="68" processingTime="88" requestCount="2" errorCount="1" bytesReceived="0" bytesSent="9286"/>
    <workers>
      <worker stage="S" requestProcessingTime="4" requestBytesSent="0" requestBytesReceived="0" remoteAddr="127.0.0.1" virtualHost="127.0.0.1" method="GET" currentUri="/manager/status/all" currentQueryString="XML=true" protocol="HTTP/1.1"/>
    </workers>
  </connector>
</status>`

func TestHTTPTomcat8(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, tomcatStatus8)
	}))
	defer ts.Close()

	tc := Tomcat{
		URL:      ts.URL,
		Username: "tomcat",
		Password: "s3cret",
	}

	var acc testutil.Accumulator
	err := tc.Gather(&acc)
	require.NoError(t, err)

	// tomcat_jvm_memory
	jvmMemoryFields := map[string]interface{}{
		"free":  int64(17909336),
		"total": int64(58195968),
		"max":   int64(620756992),
	}
	acc.AssertContainsFields(t, "tomcat_jvm_memory", jvmMemoryFields)

	// tomcat_jvm_memorypool
	jvmMemoryPoolFields := map[string]interface{}{
		"init":      int64(22020096),
		"committed": int64(22020096),
		"max":       int64(174063616),
		"used":      int64(17533952),
	}
	jvmMemoryPoolTags := map[string]string{
		"name": "PS Perm Gen",
		"type": "Non-heap memory",
	}
	acc.AssertContainsTaggedFields(t, "tomcat_jvm_memorypool", jvmMemoryPoolFields, jvmMemoryPoolTags)

	// tomcat_connector
	connectorFields := map[string]interface{}{
		"max_threads":          int64(200),
		"current_thread_count": int64(5),
		"current_threads_busy": int64(1),
		"max_time":             int(68),
		"processing_time":      int(88),
		"request_count":        int(2),
		"error_count":          int(1),
		"bytes_received":       int64(0),
		"bytes_sent":           int64(9286),
	}
	connectorTags := map[string]string{
		"name": "http-apr-8080",
	}
	acc.AssertContainsTaggedFields(t, "tomcat_connector", connectorFields, connectorTags)
}

var tomcatStatus6 = `<?xml version="1.0" encoding="utf-8"?>
<?xml-stylesheet type="text/xsl" href="xform.xsl" ?>
<status>
  <jvm>
    <memory free="1942681600" total="2040070144" max="2040070144"/>
  </jvm>
  <connector name="http-8080">
    <threadInfo maxThreads="150" currentThreadCount="2" currentThreadsBusy="2"/>
    <requestInfo maxTime="1005" processingTime="2465" requestCount="436" errorCount="16" bytesReceived="0" bytesSent="550196"/>
    <workers>
      <worker stage="K" requestProcessingTime="526" requestBytesSent="0" requestBytesReceived="0" remoteAddr="127.0.0.1" virtualHost="?" method="?" currentUri="?" currentQueryString="?" protocol="?"/>
      <worker stage="S" requestProcessingTime="1" requestBytesSent="0" requestBytesReceived="0" remoteAddr="127.0.0.1" virtualHost="127.0.0.1" method="GET" currentUri="/manager/status/all" currentQueryString="XML=true" protocol="HTTP/1.1"/>
    </workers>
  </connector>
</status>`

func TestHTTPTomcat6(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, tomcatStatus6)
	}))
	defer ts.Close()

	tc := Tomcat{
		URL:      ts.URL,
		Username: "tomcat",
		Password: "s3cret",
	}

	var acc testutil.Accumulator
	err := tc.Gather(&acc)
	require.NoError(t, err)

	// tomcat_jvm_memory
	jvmMemoryFields := map[string]interface{}{
		"free":  int64(1942681600),
		"total": int64(2040070144),
		"max":   int64(2040070144),
	}
	acc.AssertContainsFields(t, "tomcat_jvm_memory", jvmMemoryFields)

	// tomcat_connector
	connectorFields := map[string]interface{}{
		"bytes_received":       int64(0),
		"bytes_sent":           int64(550196),
		"current_thread_count": int64(2),
		"current_threads_busy": int64(2),
		"error_count":          int(16),
		"max_threads":          int64(150),
		"max_time":             int(1005),
		"processing_time":      int(2465),
		"request_count":        int(436),
	}
	connectorTags := map[string]string{
		"name": "http-8080",
	}
	acc.AssertContainsTaggedFields(t, "tomcat_connector", connectorFields, connectorTags)
}
