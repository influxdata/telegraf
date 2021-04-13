package activemq

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGatherQueuesMetrics(t *testing.T) {
	s := `<queues>
<queue name="sandra">
<stats size="0" consumerCount="0" enqueueCount="0" dequeueCount="0"/>
<feed>
<atom>queueBrowse/sandra?view=rss&amp;feedType=atom_1.0</atom>
<rss>queueBrowse/sandra?view=rss&amp;feedType=rss_2.0</rss>
</feed>
</queue>
<queue name="Test">
<stats size="0" consumerCount="0" enqueueCount="0" dequeueCount="0"/>
<feed>
<atom>queueBrowse/Test?view=rss&amp;feedType=atom_1.0</atom>
<rss>queueBrowse/Test?view=rss&amp;feedType=rss_2.0</rss>
</feed>
</queue>
</queues>`

	queues := Queues{}

	require.NoError(t, xml.Unmarshal([]byte(s), &queues))

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["name"] = "Test"
	tags["source"] = "localhost"
	tags["port"] = "8161"

	records["size"] = 0
	records["consumer_count"] = 0
	records["enqueue_count"] = 0
	records["dequeue_count"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)
	activeMQ.Server = "localhost"
	activeMQ.Port = 8161
	require.NoError(t, activeMQ.Init())

	activeMQ.GatherQueuesMetrics(&acc, queues)
	acc.AssertContainsTaggedFields(t, "activemq_queues", records, tags)
}

func TestGatherTopicsMetrics(t *testing.T) {
	s := `<topics>
<topic name="ActiveMQ.Advisory.MasterBroker ">
<stats size="0" consumerCount="0" enqueueCount="1" dequeueCount="0"/>
</topic>
<topic name="AAA ">
<stats size="0" consumerCount="1" enqueueCount="0" dequeueCount="0"/>
</topic>
<topic name="ActiveMQ.Advisory.Topic ">
<stats size="0" consumerCount="0" enqueueCount="1" dequeueCount="0"/>
</topic>
<topic name="ActiveMQ.Advisory.Queue ">
<stats size="0" consumerCount="0" enqueueCount="2" dequeueCount="0"/>
</topic>
<topic name="AAAA ">
<stats size="0" consumerCount="0" enqueueCount="0" dequeueCount="0"/>
</topic>
</topics>`

	topics := Topics{}

	require.NoError(t, xml.Unmarshal([]byte(s), &topics))

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["name"] = "ActiveMQ.Advisory.MasterBroker "
	tags["source"] = "localhost"
	tags["port"] = "8161"

	records["size"] = 0
	records["consumer_count"] = 0
	records["enqueue_count"] = 1
	records["dequeue_count"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)
	activeMQ.Server = "localhost"
	activeMQ.Port = 8161
	require.NoError(t, activeMQ.Init())

	activeMQ.GatherTopicsMetrics(&acc, topics)
	acc.AssertContainsTaggedFields(t, "activemq_topics", records, tags)
}

func TestGatherSubscribersMetrics(t *testing.T) {
	s := `<subscribers>
<subscriber clientId="AAA" subscriptionName="AAA" connectionId="NOTSET" destinationName="AAA" selector="AA" active="no">
<stats pendingQueueSize="0" dispatchedQueueSize="0" dispatchedCounter="0" enqueueCounter="0" dequeueCounter="0"/>
</subscriber>
</subscribers>`

	subscribers := Subscribers{}

	require.NoError(t, xml.Unmarshal([]byte(s), &subscribers))

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["client_id"] = "AAA"
	tags["subscription_name"] = "AAA"
	tags["connection_id"] = "NOTSET"
	tags["destination_name"] = "AAA"
	tags["selector"] = "AA"
	tags["active"] = "no"
	tags["source"] = "localhost"
	tags["port"] = "8161"

	records["pending_queue_size"] = 0
	records["dispatched_queue_size"] = 0
	records["dispatched_counter"] = 0
	records["enqueue_counter"] = 0
	records["dequeue_counter"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)
	activeMQ.Server = "localhost"
	activeMQ.Port = 8161
	require.NoError(t, activeMQ.Init())

	activeMQ.GatherSubscribersMetrics(&acc, subscribers)
	acc.AssertContainsTaggedFields(t, "activemq_subscribers", records, tags)
}

func TestURLs(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/xml/queues.jsp":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("<queues></queues>"))
			require.NoError(t, err)
		case "/admin/xml/topics.jsp":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("<topics></topics>"))
			require.NoError(t, err)
		case "/admin/xml/subscribers.jsp":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("<subscribers></subscribers>"))
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
			t.Fatalf("unexpected path: " + r.URL.Path)
		}
	})

	plugin := ActiveMQ{
		URL:      "http://" + ts.Listener.Addr().String(),
		Webadmin: "admin",
	}
	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Gather(&acc)
	require.NoError(t, err)

	require.Len(t, acc.GetTelegrafMetrics(), 0)
}
