package burrow

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
)

// fetch topics: /v2/kafka/(cluster)/topic
func gatherTopicStats(api apiClient, clusterList []string, wg *sync.WaitGroup) {
	defer wg.Done()

	producerChan := make(chan string, len(clusterList))
	doneChan := make(chan bool, len(clusterList))

	for i := 0; i < api.workerCount; i++ {
		go withAPICall(api, producerChan, doneChan, fetchTopic)
	}

	for _, cluster := range clusterList {
		escaped := url.PathEscape(cluster)
		producerChan <- fmt.Sprintf("%s/%s/topic", api.apiPrefix, escaped)
	}

	for i := len(clusterList); i > 0; i-- {
		<-doneChan
	}

	close(producerChan)
}

// fetch topic status: /v2/kafka/(clustername)/topic/(topicname)
func fetchTopic(api apiClient, res apiResponse, uri string) {

	topicList := whitelistSlice(res.Topics, api.limitTopics)

	producerChan := make(chan string, len(topicList))
	doneChan := make(chan bool, len(topicList))

	for i := 0; i < api.workerCount; i++ {
		go withAPICall(api, producerChan, doneChan, publishTopic)
	}

	for _, topic := range topicList {
		escaped := url.PathEscape(topic)
		producerChan <- fmt.Sprintf("%s/%s", uri, escaped)
	}

	for i := len(topicList); i > 0; i-- {
		<-doneChan
	}

	close(producerChan)
}

// publish topic status
func publishTopic(api apiClient, res apiResponse, uri string) {
	for i, offset := range res.Offsets {
		tags := map[string]string{
			"cluster":   res.Request.Cluster,
			"topic":     res.Request.Topic,
			"partition": strconv.Itoa(i),
		}

		api.acc.AddFields(
			"burrow_topic",
			map[string]interface{}{
				"offset": offset,
			},
			tags,
		)
	}
}
