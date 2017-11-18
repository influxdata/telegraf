package burrow

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
)

// fetch consumer groups: /v2/kafka/(cluster)/consumer
func gatherGroupStats(api apiClient, clusterList []string, wg *sync.WaitGroup) {
	defer wg.Done()

	producerChan := make(chan string, len(clusterList))
	doneChan := make(chan bool, len(clusterList))

	for i := 0; i < api.workerCount; i++ {
		go withAPICall(api, producerChan, doneChan, fetchConsumer)
	}

	for _, cluster := range clusterList {
		escaped := url.PathEscape(cluster)
		producerChan <- fmt.Sprintf("%s/%s/consumer", api.apiPrefix, escaped)
	}

	for i := len(clusterList); i > 0; i-- {
		<-doneChan
	}

	close(producerChan)
}

// fetch consumer status: /v2/kafka/(cluster)/consumer/(group)/status
func fetchConsumer(api apiClient, res apiResponse, uri string) {

	groupList := whitelistSlice(res.Groups, api.limitGroups)

	producerChan := make(chan string, len(groupList))
	doneChan := make(chan bool, len(groupList))

	for i := 0; i < api.workerCount; i++ {
		go withAPICall(api, producerChan, doneChan, publishConsumer)
	}

	for _, group := range groupList {
		escaped := url.PathEscape(group)
		producerChan <- fmt.Sprintf("%s/%s/status", uri, escaped)
	}

	for i := len(groupList); i > 0; i-- {
		<-doneChan
	}

	close(producerChan)
}

// publish consumer status
func publishConsumer(api apiClient, res apiResponse, uri string) {
	for _, partition := range res.Status.Partitions {
		status := remapStatus(partition.Status)

		tags := map[string]string{
			"cluster":   res.Request.Cluster,
			"group":     res.Request.Group,
			"topic":     partition.Topic,
			"partition": strconv.FormatInt(int64(partition.Partition), 10),
		}

		api.acc.AddFields(
			"burrow_consumer",
			map[string]interface{}{
				"start.offset":    partition.Start.Offset,
				"start.lag":       partition.Start.Lag,
				"start.timestamp": partition.Start.Timestamp,
				"end.offset":      partition.End.Offset,
				"end.lag":         partition.End.Lag,
				"end.timestamp":   partition.End.Timestamp,
				"status":          status,
			},
			tags,
		)
	}
}
