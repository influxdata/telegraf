//go:generate ../../../tools/readme_config_includer/generator
package kafka

import (
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Number of attempts for fetching log end offsets.
const logEndOffsetAttempts = 2

type Kafka struct {
	Brokers             []string        `toml:"brokers"`
	Cluster             string          `toml:"cluster"`
	GroupsInclude       []string        `toml:"groups_include"`
	GroupsExclude       []string        `toml:"groups_exclude"`
	TopicsInclude       []string        `toml:"topics_include"`
	TopicsExclude       []string        `toml:"topics_exclude"`
	MetricLevels        []string        `toml:"metric_levels"`
	CoordinatorBrokerID int32           `toml:"coordinator_broker_id"`
	Log                 telegraf.Logger `toml:"-"`
	kafka.Config

	config *sarama.Config
	client sarama.Client
	admin  sarama.ClusterAdmin

	filterGroups filter.Filter
	filterTopics filter.Filter

	// Fetch the offsets of all groups sharing a coordinator in one request.
	batchOffsets bool
}

func (*Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Init() error {
	if len(k.Brokers) == 0 {
		return errors.New("brokers must not be empty")
	}

	kafka.SetLogger(k.Log.Level())

	var err error
	k.filterGroups, err = filter.NewIncludeExcludeFilter(k.GroupsInclude, k.GroupsExclude)
	if err != nil {
		return fmt.Errorf("creating group filter failed: %w", err)
	}
	k.filterTopics, err = filter.NewIncludeExcludeFilter(k.TopicsInclude, k.TopicsExclude)
	if err != nil {
		return fmt.Errorf("creating topic filter failed: %w", err)
	}

	if len(k.MetricLevels) == 0 {
		return errors.New("metric_levels must not be empty")
	}
	for _, level := range k.MetricLevels {
		switch level {
		case "partition", "topic", "group":
		default:
			return fmt.Errorf("invalid metric level %q", level)
		}
	}

	cfg := sarama.NewConfig()
	if err := k.SetConfig(cfg, k.Log); err != nil {
		return fmt.Errorf("setting config failed: %w", err)
	}
	// Fetching all partitions of a group needs OffsetFetch v2 and refusing
	// topic auto-creation needs Metadata v4, both introduced with 0.11.0.0.
	// Older brokers would silently return no offsets or recreate topics.
	if !cfg.Version.IsAtLeast(sarama.V0_11_0_0) {
		return fmt.Errorf("kafka version %s is not supported, 0.11.0.0 or greater is required", cfg.Version)
	}
	// Metadata requests for a deleted topic must never recreate it.
	cfg.Metadata.AllowAutoTopicCreation = false
	k.config = cfg
	// Fetching the offsets of several groups at once needs OffsetFetch v8.
	k.batchOffsets = cfg.Version.IsAtLeast(sarama.V3_0_0_0)

	return nil
}

func (k *Kafka) Start(telegraf.Accumulator) error {
	var err error
	k.client, err = sarama.NewClient(k.Brokers, k.config)
	if err != nil {
		return &internal.StartupError{
			Err:   fmt.Errorf("creating client failed: %w", err),
			Retry: errors.Is(err, sarama.ErrOutOfBrokers),
		}
	}

	// With metadata_full disabled sarama fetches no metadata on connect and
	// creating the cluster admin fails as it cannot find the controller, so
	// fetch it once. Later refreshes only cover the topics in use.
	if !k.config.Metadata.Full {
		if err := k.client.RefreshMetadata(); err != nil {
			k.Stop()
			return &internal.StartupError{
				Err:   fmt.Errorf("fetching metadata failed: %w", err),
				Retry: errors.Is(err, sarama.ErrOutOfBrokers),
			}
		}
	}

	k.admin, err = sarama.NewClusterAdminFromClient(k.client)
	if err != nil {
		k.Stop()
		// This only fails if the controller cannot be resolved, e.g. during an
		// election or a rolling restart.
		return &internal.StartupError{
			Err:   fmt.Errorf("creating cluster admin failed: %w", err),
			Retry: true,
		}
	}

	return nil
}

func (k *Kafka) Stop() {
	// Closing the admin also closes the underlying client.
	if k.admin != nil {
		if err := k.admin.Close(); err != nil {
			k.Log.Errorf("Closing cluster admin failed: %v", err)
		}
	} else if k.client != nil {
		if err := k.client.Close(); err != nil {
			k.Log.Errorf("Closing client failed: %v", err)
		}
	}
	k.admin = nil
	k.client = nil
}

func (k *Kafka) Gather(acc telegraf.Accumulator) error {
	// List the consumer groups known to the brokers matching the group filter
	groups, err := k.selectGroups(acc)
	if err != nil {
		acc.AddError(err)
		return nil
	}
	if len(groups) == 0 {
		return nil
	}

	// Get the number of members of each group, only reported on the group level
	var members map[string]int64
	if slices.Contains(k.MetricLevels, "group") {
		members = k.describeGroups(acc, groups)
	}

	// Get the offsets committed by the groups for the topics matching the
	// topic filter
	committed := k.committedOffsets(acc, groups)

	// Collect the partitions to fetch the log end offset for. The log end
	// offset is a property of the partition, so it is fetched once even if
	// several groups consume the same partition.
	logEndNeeded := make(map[topicPartition]bool)
	for _, offsets := range committed {
		for tp := range offsets {
			logEndNeeded[tp] = true
		}
	}

	// Committed offsets outlive their topic, so skip the partitions of topics
	// that no longer exist
	logEndNeeded = k.dropDeletedTopics(acc, logEndNeeded)

	// Get the log end offsets from the partition leaders
	logEnd := k.logEndOffsets(acc, logEndNeeded)

	// Compute the lag of each group and emit the metrics
	for _, group := range groups {
		offsets, ok := committed[group]
		if !ok {
			continue
		}
		k.emit(acc, group, members, offsets, logEnd)
	}

	return nil
}

// selectGroups returns the sorted consumer groups matching the group filter.
func (k *Kafka) selectGroups(acc telegraf.Accumulator) ([]string, error) {
	var brokers []*sarama.Broker
	if k.CoordinatorBrokerID >= 0 {
		// If a coordinator broker is configured, only that broker is asked.
		broker, err := k.coordinatorBroker()
		if err != nil {
			return nil, fmt.Errorf("getting coordinator broker failed: %w", err)
		}
		brokers = []*sarama.Broker{broker}
	} else {
		brokers = k.client.Brokers()
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	selected := make(map[string]bool, len(brokers))
	for _, broker := range brokers {
		wg.Go(func() {
			groups, err := k.listGroups(broker)
			if err != nil {
				acc.AddError(fmt.Errorf("listing groups on broker %d failed: %w", broker.ID(), err))
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for group := range groups {
				if k.filterGroups.Match(group) {
					selected[group] = true
				}
			}
		})
	}
	wg.Wait()

	return slices.Sorted(maps.Keys(selected)), nil
}

// coordinatorBroker returns the broker configured as coordinator_broker_id.
func (k *Kafka) coordinatorBroker() (*sarama.Broker, error) {
	broker, err := k.client.Broker(k.CoordinatorBrokerID)
	if errors.Is(err, sarama.ErrBrokerNotFound) {
		// The broker might have rejoined the cluster since the last metadata
		// refresh, so refresh once before giving up.
		if err := k.client.RefreshMetadata(); err != nil {
			return nil, fmt.Errorf("refreshing metadata failed: %w", err)
		}
		broker, err = k.client.Broker(k.CoordinatorBrokerID)
	}
	if err != nil {
		return nil, fmt.Errorf("finding broker %d failed: %w", k.CoordinatorBrokerID, err)
	}
	return broker, nil
}

// listGroups asks a single broker for the consumer groups it coordinates.
func (k *Kafka) listGroups(broker *sarama.Broker) (map[string]string, error) {
	if err := broker.Open(k.config); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
		return nil, fmt.Errorf("connecting failed: %w", err)
	}

	request := &sarama.ListGroupsRequest{}
	switch {
	case k.config.Version.IsAtLeast(sarama.V3_8_0_0):
		request.Version = 5
	case k.config.Version.IsAtLeast(sarama.V2_6_0_0):
		request.Version = 4
	case k.config.Version.IsAtLeast(sarama.V2_4_0_0):
		request.Version = 3
	case k.config.Version.IsAtLeast(sarama.V2_0_0_0):
		request.Version = 2
	case k.config.Version.IsAtLeast(sarama.V0_11_0_0):
		request.Version = 1
	}

	response, err := broker.ListGroups(request)
	if err != nil {
		// Force a reconnect on the next use, as sarama does itself.
		_ = broker.Close()
		return nil, fmt.Errorf("sending request failed: %w", err)
	}
	if !errors.Is(response.Err, sarama.ErrNoError) {
		return nil, response.Err
	}

	return response.Groups, nil
}

// describeGroups returns the number of members of each group. Groups that
// could not be described are missing from the result.
func (k *Kafka) describeGroups(acc telegraf.Accumulator, groups []string) map[string]int64 {
	descriptions, err := k.admin.DescribeConsumerGroups(groups)
	if err != nil {
		acc.AddError(fmt.Errorf("describing consumer groups failed: %w", err))
		return make(map[string]int64)
	}

	members := make(map[string]int64, len(descriptions))
	for _, description := range descriptions {
		switch {
		case errors.Is(description.Err, sarama.ErrGroupIDNotFound),
			errors.Is(description.Err, sarama.ErrNoError) && description.State == "Dead":
			// Groups using the KIP-848 consumer protocol cannot be described
			// with the classic DescribeGroups API. Brokers report them as an
			// empty "Dead" group up to v5 and as GROUP_ID_NOT_FOUND from v6.
			// Omit the member count instead of reporting zero members.
			k.Log.Debugf("Skipping members of group %q: group cannot be described (state %q, error %v)",
				description.GroupId, description.State, description.Err)
		case errors.Is(description.Err, sarama.ErrNoError):
			members[description.GroupId] = int64(len(description.Members))
		default:
			acc.AddError(fmt.Errorf("describing group %q failed: %w", description.GroupId, description.Err))
		}
	}

	return members
}

// committedOffsets returns the committed offsets of every group for the
// partitions of topics matching the topic filter.
func (k *Kafka) committedOffsets(acc telegraf.Accumulator, groups []string) map[string]map[topicPartition]int64 {
	// Offsets of each group as returned by the brokers, by topic and partition
	results := make(map[string]map[string]map[int32]*sarama.OffsetFetchResponseBlock, len(groups))

	// Fetch the offsets of all groups with one request per coordinator
	var batched bool
	if k.batchOffsets {
		request := make(map[string]map[string][]int32, len(groups))
		for _, group := range groups {
			// A nil partition map fetches all partitions the group committed on.
			request[group] = nil
		}

		responses, err := k.admin.ListConsumerGroupOffsetsBatch(request)
		batched = err == nil
		if batched {
			for _, group := range groups {
				response, ok := responses[group]
				if !ok {
					acc.AddError(fmt.Errorf("batch listing offsets of group %q failed: group missing from response", group))
					continue
				}
				if !errors.Is(response.Err, sarama.ErrNoError) {
					acc.AddError(fmt.Errorf("batch listing offsets of group %q failed: %w", group, response.Err))
					continue
				}
				results[group] = response.Blocks
			}
		}

		// On error, fall back to fetching the groups with individual requests
		if errors.Is(err, sarama.ErrUnsupportedVersion) {
			// The brokers are older than the configured version and do not
			// speak OffsetFetch v8, so stick to one request per group.
			k.Log.Debug("Brokers do not support batched offset fetching, falling back to per-group requests")
			k.batchOffsets = false
		} else if err != nil {
			// Retry group by group to only lose the groups actually affected.
			k.Log.Debugf("Batched offset fetching failed, falling back to per-group requests: %v", err)
		}
	}

	if !batched {
		for _, group := range groups {
			response, err := k.admin.ListConsumerGroupOffsets(group, nil)
			if err != nil {
				acc.AddError(fmt.Errorf("listing offsets of group %q failed: %w", group, err))
				continue
			}
			results[group] = response.Blocks
		}
	}

	// Extract the offsets of the partitions of topics matching the filter
	committed := make(map[string]map[topicPartition]int64, len(results))
	for group, topics := range results {
		offsets := make(map[topicPartition]int64)
		for topic, partitions := range topics {
			if !k.filterTopics.Match(topic) {
				continue
			}
			for partition, block := range partitions {
				if !errors.Is(block.Err, sarama.ErrNoError) {
					k.Log.Debugf("Skipping offset of group %q for %s/%d: %v", group, topic, partition, block.Err)
					continue
				}
				// A negative offset means the group never committed on this
				// partition, so there is nothing to compute a lag from.
				if block.Offset < 0 {
					continue
				}
				offsets[topicPartition{topic: topic, partition: partition}] = block.Offset
			}
		}
		committed[group] = offsets
	}

	return committed
}

// dropDeletedTopics removes the partitions of topics that no longer exist.
func (k *Kafka) dropDeletedTopics(acc telegraf.Accumulator, partitions map[topicPartition]bool) map[topicPartition]bool {
	if len(partitions) == 0 {
		return partitions
	}

	descriptions, err := k.describeTopics(uniqueTopics(partitions))
	if err != nil {
		acc.AddError(fmt.Errorf("describing topics failed: %w", err))
		return partitions
	}

	deleted := make(map[string]bool)
	for _, description := range descriptions {
		if errors.Is(description.Err, sarama.ErrUnknownTopicOrPartition) {
			k.Log.Debugf("Skipping deleted topic %q", description.Name)
			deleted[description.Name] = true
		}
	}
	if len(deleted) == 0 {
		return partitions
	}

	remaining := make(map[topicPartition]bool, len(partitions))
	for tp := range partitions {
		if !deleted[tp.topic] {
			remaining[tp] = true
		}
	}
	return remaining
}

// describeTopics fetches the metadata of the given topics from any broker.
func (k *Kafka) describeTopics(topics []string) ([]*sarama.TopicMetadata, error) {
	brokers := k.client.Brokers()
	if len(brokers) == 0 {
		return nil, errors.New("no broker available")
	}

	request := sarama.NewMetadataRequest(k.config.Version, topics)
	request.AllowAutoTopicCreation = false

	var errs []error
	for _, broker := range brokers {
		if err := broker.Open(k.config); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
			errs = append(errs, fmt.Errorf("connecting to broker %d failed: %w", broker.ID(), err))
			continue
		}
		response, err := broker.GetMetadata(request)
		if err != nil {
			// Force a reconnect on the next use, as sarama does itself.
			_ = broker.Close()
			errs = append(errs, fmt.Errorf("fetching metadata from broker %d failed: %w", broker.ID(), err))
			continue
		}
		return response.Topics, nil
	}

	return nil, errors.Join(errs...)
}

// logEndOffsets resolves the log end offset of the given partitions using one
// ListOffsets request per leader broker.
func (k *Kafka) logEndOffsets(acc telegraf.Accumulator, partitions map[topicPartition]bool) map[topicPartition]int64 {
	result := make(map[topicPartition]int64, len(partitions))
	lastErr := make(map[topicPartition]error)

	pending := partitions
	for attempt := 0; attempt < logEndOffsetAttempts && len(pending) > 0; attempt++ {
		if attempt > 0 {
			// Refresh the cached leaders before retrying, sarama does not do it on its own.
			if err := k.client.RefreshMetadata(uniqueTopics(pending)...); err != nil {
				acc.AddError(fmt.Errorf("refreshing metadata for retrying log end offsets failed: %w", err))
				break
			}
		}

		retry := make(map[topicPartition]bool)
		for _, batch := range k.batchByLeader(pending, retry, lastErr) {
			response, err := batch.broker.GetAvailableOffsets(batch.request)
			if err != nil {
				// Force a reconnect on the next use, as sarama does itself.
				_ = batch.broker.Close()
				for _, tp := range batch.partitions {
					retry[tp] = true
					lastErr[tp] = fmt.Errorf("requesting offsets from broker %d failed: %w", batch.broker.ID(), err)
				}
				continue
			}

			for _, tp := range batch.partitions {
				block := response.GetBlock(tp.topic, tp.partition)
				switch {
				case block == nil:
					retry[tp] = true
					lastErr[tp] = fmt.Errorf("broker %d returned no offset", batch.broker.ID())
				case errors.Is(block.Err, sarama.ErrNoError):
					// ListOffsets v0 returns a list of offsets, later versions a single one.
					result[tp] = block.Offset
					if len(block.Offsets) > 0 {
						result[tp] = block.Offsets[0]
					}
				case errors.Is(block.Err, sarama.ErrUnknownTopicOrPartition):
					// The topic was deleted after dropDeletedTopics ran.
					k.Log.Debugf("Skipping %s/%d: %v", tp.topic, tp.partition, block.Err)
				case isRetriable(block.Err):
					retry[tp] = true
					lastErr[tp] = block.Err
				default:
					acc.AddError(fmt.Errorf("fetching log end offset of %s/%d failed: %w", tp.topic, tp.partition, block.Err))
				}
			}
		}
		pending = retry
	}

	for tp := range pending {
		acc.AddError(fmt.Errorf("giving up on log end offset of %s/%d: %w", tp.topic, tp.partition, lastErr[tp]))
	}

	return result
}

type offsetBatch struct {
	broker     *sarama.Broker
	request    *sarama.OffsetRequest
	partitions []topicPartition
}

// batchByLeader groups the partitions into one offset request per leader broker.
func (k *Kafka) batchByLeader(partitions, retry map[topicPartition]bool, lastErr map[topicPartition]error) map[int32]*offsetBatch {
	batches := make(map[int32]*offsetBatch)
	for tp := range partitions {
		leader, err := k.client.Leader(tp.topic, tp.partition)
		if err != nil {
			if errors.Is(err, sarama.ErrUnknownTopicOrPartition) {
				k.Log.Debugf("Skipping %s/%d: %v", tp.topic, tp.partition, err)
			} else {
				retry[tp] = true
				lastErr[tp] = fmt.Errorf("finding leader failed: %w", err)
			}
			continue
		}

		batch, ok := batches[leader.ID()]
		if !ok {
			batch = &offsetBatch{
				broker:  leader,
				request: sarama.NewOffsetRequest(k.config.Version),
			}
			batches[leader.ID()] = batch
		}
		batch.request.AddBlock(tp.topic, tp.partition, sarama.OffsetNewest, 1)
		batch.partitions = append(batch.partitions, tp)
	}

	return batches
}

// tags adds the cluster tag to the given tags if a cluster name is configured.
func (k *Kafka) tags(tags map[string]string) map[string]string {
	if k.Cluster != "" {
		tags["cluster"] = k.Cluster
	}
	return tags
}

func (k *Kafka) emit(
	acc telegraf.Accumulator,
	group string,
	members map[string]int64,
	committed, logEnd map[topicPartition]int64,
) {
	topicAggregates := make(map[string]*lagAggregate)
	var groupAggregate lagAggregate

	for tp, committedOffset := range committed {
		logEndOffset, ok := logEnd[tp]
		if !ok {
			continue
		}

		// Only negative after a topic was recreated or its log truncated, as
		// offsets are fetched in commit then log-end order.
		lag := max(logEndOffset-committedOffset, 0)

		if slices.Contains(k.MetricLevels, "partition") {
			tags := k.tags(map[string]string{
				"group":     group,
				"topic":     tp.topic,
				"partition": strconv.FormatInt(int64(tp.partition), 10),
			})
			fields := map[string]any{
				"committed_offset": committedOffset,
				"log_end_offset":   logEndOffset,
				"lag":              lag,
			}
			acc.AddGauge("kafka_consumer_group_partition", fields, tags)
		}

		aggregate, ok := topicAggregates[tp.topic]
		if !ok {
			aggregate = &lagAggregate{}
			topicAggregates[tp.topic] = aggregate
		}
		aggregate.add(lag)
		groupAggregate.add(lag)
	}

	if slices.Contains(k.MetricLevels, "topic") {
		for topic, aggregate := range topicAggregates {
			tags := k.tags(map[string]string{
				"group": group,
				"topic": topic,
			})
			fields := map[string]any{
				"lag_sum":    aggregate.sum,
				"lag_max":    aggregate.max,
				"partitions": aggregate.partitions,
			}
			acc.AddGauge("kafka_consumer_group_topic", fields, tags)
		}
	}

	if slices.Contains(k.MetricLevels, "group") && groupAggregate.partitions > 0 {
		tags := k.tags(map[string]string{"group": group})
		fields := map[string]any{
			"lag_sum":    groupAggregate.sum,
			"lag_max":    groupAggregate.max,
			"partitions": groupAggregate.partitions,
			"topics":     int64(len(topicAggregates)),
		}
		if n, ok := members[group]; ok {
			fields["members"] = n
		}
		acc.AddGauge("kafka_consumer_group", fields, tags)
	}
}

// isRetriable reports whether a ListOffsets error is expected to resolve
// after a metadata refresh.
func isRetriable(err sarama.KError) bool {
	switch err {
	case sarama.ErrLeaderNotAvailable,
		sarama.ErrNotLeaderForPartition,
		sarama.ErrReplicaNotAvailable,
		sarama.ErrFencedLeaderEpoch,
		sarama.ErrUnknownLeaderEpoch,
		sarama.ErrOffsetNotAvailable,
		sarama.ErrKafkaStorageError,
		sarama.ErrRequestTimedOut,
		sarama.ErrNetworkException:
		return true
	default:
		return false
	}
}

func uniqueTopics(partitions map[topicPartition]bool) []string {
	seen := make(map[string]bool)
	topics := make([]string, 0, len(partitions))
	for tp := range partitions {
		if seen[tp.topic] {
			continue
		}
		seen[tp.topic] = true
		topics = append(topics, tp.topic)
	}
	return topics
}

func init() {
	inputs.Add("kafka", func() telegraf.Input {
		return &Kafka{
			MetricLevels:        []string{"partition", "topic", "group"},
			CoordinatorBrokerID: -1,
		}
	})
}
