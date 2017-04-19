package activemq

import (
	"encoding/xml"
	"testing"

	"github.com/influxdata/telegraf/testutil"
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

	xml.Unmarshal([]byte(s), &queues)

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["name"] = "Test"

	records["size"] = 0
	records["consumer_count"] = 0
	records["enqueue_count"] = 0
	records["dequeue_count"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)

	activeMQ.GatherQueuesMetrics(&acc, queues)
	acc.AssertContainsTaggedFields(t, "queues_metrics", records, tags)
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

	xml.Unmarshal([]byte(s), &topics)

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["name"] = "ActiveMQ.Advisory.MasterBroker "

	records["size"] = 0
	records["consumer_count"] = 0
	records["enqueue_count"] = 1
	records["dequeue_count"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)

	activeMQ.GatherTopicsMetrics(&acc, topics)
	acc.AssertContainsTaggedFields(t, "topics_metrics", records, tags)
}

func TestGatherSubscribersMetrics(t *testing.T) {

	s := `<subscribers>
<subscriber clientId="AAA" subscriptionName="AAA" connectionId="NOTSET" destinationName="AAA" selector="AA" active="no">
<stats pendingQueueSize="0" dispatchedQueueSize="0" dispatchedCounter="0" enqueueCounter="0" dequeueCounter="0"/>
</subscriber>
</subscribers>`

	subscribers := Subscribers{}

	xml.Unmarshal([]byte(s), &subscribers)

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["client_id"] = "AAA"
	tags["subscription_name"] = "AAA"
	tags["connection_id"] = "NOTSET"
	tags["destination_name"] = "AAA"
	tags["selector"] = "AA"
	tags["active"] = "no"

	records["pending_queue_size"] = 0
	records["dispatched_queue_size"] = 0
	records["dispatched_counter"] = 0
	records["enqueue_counter"] = 0
	records["dequeue_counter"] = 0

	var acc testutil.Accumulator

	activeMQ := new(ActiveMQ)

	activeMQ.GatherSubscribersMetrics(&acc, subscribers)
	acc.AssertContainsTaggedFields(t, "subscribers_metrics", records, tags)
}
