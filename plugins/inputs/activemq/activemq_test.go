package activemq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

var activemq_info = `{"request":{"mbean":"org.apache.activemq:brokerName=localhost,type=Broker","type":"read"},"value":{"StatisticsEnabled":true,"TotalConnectionsCount":180,"StompSslURL":"","TransportConnectors":{"openwire":"tcp:\/\/i-6a6d0ghv:61616?maximumConnections=1000&wireFormat.maxFrameSize=104857600","amqp":"amqp:\/\/i-6a6d0ghv:5672?maximumConnections=1000&wireFormat.maxFrameSize=104857600","mqtt":"mqtt:\/\/i-6a6d0ghv:1883?maximumConnections=1000&wireFormat.maxFrameSize=104857600","stomp":"stomp:\/\/i-6a6d0ghv:61613?maximumConnections=1000&wireFormat.maxFrameSize=104857600","ws":"ws:\/\/i-6a6d0ghv:61614?maximumConnections=1000&wireFormat.maxFrameSize=104857600"},"StompURL":"stomp:\/\/i-6a6d0ghv:61613?maximumConnections=1000&wireFormat.maxFrameSize=104857600","TotalProducerCount":0,"CurrentConnectionsCount":9,"TopicProducers":[],"JMSJobScheduler":null,"UptimeMillis":7608145,"TemporaryQueueProducers":[],"TotalDequeueCount":206,"JobSchedulerStorePercentUsage":0,"DurableTopicSubscribers":[],"QueueSubscribers":[{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_31,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_31_1_1,destinationName=four_cache_log_dev,destinationType=Queue,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_35,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_35_1_1,destinationName=four_cache_log,destinationType=Queue,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_1,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_1_1_1,destinationName=git_,destinationType=Queue,endpoint=Consumer,type=Broker"}],"AverageMessageSize":1789,"BrokerVersion":"5.12.1","TemporaryQueues":[],"BrokerName":"localhost","MinMessageSize":1024,"DynamicDestinationProducers":[{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_180,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_180_1_1,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_176,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_176_1_1,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_179,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_179_1_1,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_177,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_177_1_1,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_175,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_175_1_1,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_178,endpoint=dynamicProducer,producerId=ID_i-6a6d0ghv-36349-1480685087423-3_178_1_1,type=Broker"}],"OpenWireURL":"tcp:\/\/i-6a6d0ghv:61616?maximumConnections=1000&wireFormat.maxFrameSize=104857600","TemporaryTopics":[],"JobSchedulerStoreLimit":0,"TotalConsumerCount":3,"MaxMessageSize":4374,"TotalMessageCount":0,"TempPercentUsage":0,"TemporaryQueueSubscribers":[],"MemoryPercentUsage":0,"SslURL":"","InactiveDurableTopicSubscribers":[],"StoreLimit":16292901387,"QueueProducers":[],"VMURL":"vm:\/\/localhost","TemporaryTopicProducers":[],"Topics":[{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.Consumer.Queue.four_cache_log_dev,destinationType=Topic,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.MasterBroker,destinationType=Topic,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.Connection,destinationType=Topic,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.Consumer.Queue.four_cache_log,destinationType=Topic,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.Consumer.Queue.git_,destinationType=Topic,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=ActiveMQ.Advisory.Queue,destinationType=Topic,type=Broker"}],"Uptime":"2 hours 6 minutes","BrokerId":"ID:i-6a6d0ghv-36349-1480685087423-0:1","DataDirectory":"\/root\/apache-activemq-5.12.1\/data","Persistent":true,"TopicSubscribers":[{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_175,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_175_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_35,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_35_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_179,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_179_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_176,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_176_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_177,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_177_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_31,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_31_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_1,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_1_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_180,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_180_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,clientId=ID_i-6a6d0ghv-36349-1480685087423-2_178,consumerId=ID_i-6a6d0ghv-36349-1480685087423-3_178_-1_1,destinationName=ActiveMQ.Advisory.TempQueue_ActiveMQ.Advisory.TempTopic,destinationType=Topic,endpoint=Consumer,type=Broker"}],"MemoryLimit":726571418,"Slave":false,"TotalEnqueueCount":533,"TempLimit":16292732928,"TemporaryTopicSubscribers":[],"StorePercentUsage":0,"Queues":[{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=four_cache_log,destinationType=Queue,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=git_,destinationType=Queue,type=Broker"},{"objectName":"org.apache.activemq:brokerName=localhost,destinationName=four_cache_log_dev,destinationType=Queue,type=Broker"}]},"timestamp":1480692694,"status":200}`

func TestHTTPActivemq(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, activemq_info)
	}))
	defer ts.Close()

	a := Activemq{
		// Fetch it 2 times to catch possible data races.
		Urls: []string{ts.URL, ts.URL},
	}

	var acc testutil.Accumulator
	err := a.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"TotalConnectionsCount":   uint64(180),
		"TotalProducerCount":      uint64(0),
		"CurrentConnectionsCount": uint64(9),
		"TotalDequeueCount":       uint64(206),
		"AverageMessageSize":      float64(1789),
		"MinMessageSize":          float64(1024),
		"TotalConsumerCount":      uint64(3),
		"MaxMessageSize":          float64(4374),
		"TotalMessageCount":       uint64(0),
		"MemoryPercentUsage":      float64(0),
		"TotalEnqueueCount":       uint64(533),
	}
	acc.AssertContainsFields(t, "activemq", fields)
}
