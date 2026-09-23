//go:generate ../../../tools/readme_config_includer/generator
package kafka_consumer_lag

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

const (
	levelPartition = "partition"
	levelTopic     = "topic"
	levelGroup     = "group"

	// Number of attempts for fetching log end offsets.
	logEndOffsetAttempts = 2
)

type KafkaConsumerLag struct {
	Brokers             []string        `toml:"brokers"`
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

	emitPartition bool
	emitTopic     bool
	emitGroup     bool

	// Fetch the offsets of all groups sharing a coordinator in one request.
	batchOffsets bool
}

// topicPartition identifies a single partition of a topic.
type topicPartition struct {
	topic     string
	partition int32
}

// lagAggregate accumulates per-partition lag into sums and maxima.
type lagAggregate struct {
	sum        int64
	max        int64
	partitions int64
}

func (a *lagAggregate) add(lag int64) {
	a.sum += lag
	a.partitions++
	a.max = max(a.max, lag)
}

func (*KafkaConsumerLag) SampleConfig() string {
	return sampleConfig
}

func (k *KafkaConsumerLag) Init() error {
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
		case levelPartition:
			k.emitPartition = true
		case levelTopic:
			k.emitTopic = true
		case levelGroup:
			k.emitGroup = true
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

func (k *KafkaConsumerLag) Start(telegraf.Accumulator) error {
	client, err := sarama.NewClient(k.Brokers, k.config)
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
		if err := client.RefreshMetadata(); err != nil {
			_ = client.Close()
			return &internal.StartupError{
				Err:   fmt.Errorf("fetching metadata failed: %w", err),
				Retry: errors.Is(err, sarama.ErrOutOfBrokers),
			}
		}
	}

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		_ = client.Close()
		return &internal.StartupError{Err: fmt.Errorf("creating cluster admin failed: %w", err)}
	}

	k.client = client
	k.admin = admin

	return nil
}

func (k *KafkaConsumerLag) Stop() {
	// Closing the admin also closes the underlying client.
	if k.admin != nil {
		if err := k.admin.Close(); err != nil {
			k.Log.Errorf("Closing cluster admin failed: %v", err)
		}
	}
	k.admin = nil
	k.client = nil
}

func (k *KafkaConsumerLag) Gather(acc telegraf.Accumulator) error {
	groups := k.selectGroups(acc)
	if len(groups) == 0 {
		return nil
	}

	// The member count is only reported on the group level.
	var members map[string]int64
	if k.emitGroup {
		members = k.describeGroups(groups, acc)
	}

	committed, errs := k.committedOffsets(groups)
	for _, err := range errs {
		acc.AddError(err)
	}

	// The log end offset is a property of the partition, so it is fetched
	// once even if several groups consume the same partition.
	logEndNeeded := make(map[topicPartition]struct{})
	for _, offsets := range committed {
		for tp := range offsets {
			logEndNeeded[tp] = struct{}{}
		}
	}

	logEnd, errs := k.logEndOffsets(logEndNeeded)
	for _, err := range errs {
		acc.AddError(err)
	}

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
func (k *KafkaConsumerLag) selectGroups(acc telegraf.Accumulator) []string {
	var brokers []*sarama.Broker
	if k.CoordinatorBrokerID >= 0 {
		// If a coordinator broker is configured, only that broker is asked.
		broker, err := k.coordinatorBroker()
		if err != nil {
			acc.AddError(err)
			return nil
		}
		brokers = []*sarama.Broker{broker}
	} else {
		brokers = k.client.Brokers()
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	selected := make(map[string]struct{})
	for _, broker := range brokers {
		wg.Go(func() {
			groups, err := k.listGroups(broker)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				acc.AddError(err)
				return
			}
			for group := range groups {
				if k.filterGroups.Match(group) {
					selected[group] = struct{}{}
				}
			}
		})
	}
	wg.Wait()

	return slices.Sorted(maps.Keys(selected))
}

// coordinatorBroker returns the broker configured as coordinator_broker_id.
func (k *KafkaConsumerLag) coordinatorBroker() (*sarama.Broker, error) {
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
func (k *KafkaConsumerLag) listGroups(broker *sarama.Broker) (map[string]string, error) {
	if err := broker.Open(k.config); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
		return nil, fmt.Errorf("connecting to broker %d failed: %w", broker.ID(), err)
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
		return nil, fmt.Errorf("listing groups on broker %d failed: %w", broker.ID(), err)
	}
	if !errors.Is(response.Err, sarama.ErrNoError) {
		return nil, fmt.Errorf("listing groups on broker %d failed: %w", broker.ID(), response.Err)
	}

	return response.Groups, nil
}

// describeGroups returns the number of members of each group. Groups that
// could not be described are missing from the result.
func (k *KafkaConsumerLag) describeGroups(groups []string, acc telegraf.Accumulator) map[string]int64 {
	members := make(map[string]int64, len(groups))

	descriptions, err := k.admin.DescribeConsumerGroups(groups)
	if err != nil {
		acc.AddError(fmt.Errorf("describing consumer groups failed: %w", err))
		return members
	}

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

// committedOffsets returns the committed offsets of every group.
func (k *KafkaConsumerLag) committedOffsets(groups []string) (map[string]map[topicPartition]int64, []error) {
	committed := make(map[string]map[topicPartition]int64, len(groups))
	var errs []error

	// Fetch the offsets of all groups with one request per coordinator.
	if k.batchOffsets {
		request := make(map[string]map[string][]int32, len(groups))
		for _, group := range groups {
			// A nil partition map fetches all partitions the group committed on.
			request[group] = nil
		}

		responses, err := k.admin.ListConsumerGroupOffsetsBatch(request)
		switch {
		case err == nil:
			for _, group := range groups {
				response, ok := responses[group]
				if !ok {
					errs = append(errs, fmt.Errorf("listing offsets of group %q failed: group missing from response", group))
					continue
				}
				if !errors.Is(response.Err, sarama.ErrNoError) {
					errs = append(errs, fmt.Errorf("listing offsets of group %q failed: %w", group, response.Err))
					continue
				}
				committed[group] = k.filterOffsets(group, response.Blocks)
			}
			return committed, errs
		case errors.Is(err, sarama.ErrUnsupportedVersion):
			// The brokers are older than the configured version and do not
			// speak OffsetFetch v8, so stick to one request per group.
			k.Log.Debug("Brokers do not support batched offset fetching, falling back to per-group requests")
			k.batchOffsets = false
		default:
			// Retry group by group to only lose the groups actually affected.
			k.Log.Debugf("Batched offset fetching failed, falling back to per-group requests: %v", err)
		}
	}

	for _, group := range groups {
		response, err := k.admin.ListConsumerGroupOffsets(group, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("listing offsets of group %q failed: %w", group, err))
			continue
		}
		committed[group] = k.filterOffsets(group, response.Blocks)
	}

	return committed, errs
}

// filterOffsets extracts the committed offset of every partition the group
// has an offset stored for from an OffsetFetch response, restricted by the
// topic filter.
func (k *KafkaConsumerLag) filterOffsets(group string, blocks map[string]map[int32]*sarama.OffsetFetchResponseBlock) map[topicPartition]int64 {
	offsets := make(map[topicPartition]int64)
	for topic, partitions := range blocks {
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

	return offsets
}

// logEndOffsets resolves the log end offset of the given partitions using one
// ListOffsets request per leader broker.
func (k *KafkaConsumerLag) logEndOffsets(partitions map[topicPartition]struct{}) (map[topicPartition]int64, []error) {
	result := make(map[topicPartition]int64, len(partitions))
	lastErr := make(map[topicPartition]error)
	var errs []error

	// Committed offsets outlive their topic, we emit only the partitions of topics that still exist.
	pending, err := k.dropDeletedTopics(partitions)
	if err != nil {
		errs = append(errs, err)
	}
	for attempt := 0; attempt < logEndOffsetAttempts && len(pending) > 0; attempt++ {
		if attempt > 0 {
			// Refresh the cached leaders before retrying, sarama does not do it on its own.
			if err := k.client.RefreshMetadata(uniqueTopics(pending)...); err != nil {
				errs = append(errs, fmt.Errorf("refreshing metadata failed: %w", err))
				break
			}
		}

		retry := make(map[topicPartition]struct{})
		for _, batch := range k.batchByLeader(pending, retry, lastErr) {
			response, err := batch.broker.GetAvailableOffsets(batch.request)
			if err != nil {
				// Force a reconnect on the next use, as sarama does itself.
				_ = batch.broker.Close()
				for _, tp := range batch.partitions {
					retry[tp] = struct{}{}
					lastErr[tp] = fmt.Errorf("fetching offsets from broker %d failed: %w", batch.broker.ID(), err)
				}
				continue
			}

			for _, tp := range batch.partitions {
				block := response.GetBlock(tp.topic, tp.partition)
				switch {
				case block == nil:
					retry[tp] = struct{}{}
					lastErr[tp] = fmt.Errorf("broker %d returned no offset for %s/%d", batch.broker.ID(), tp.topic, tp.partition)
				case errors.Is(block.Err, sarama.ErrNoError):
					result[tp] = logEndOffsetOf(block)
				case errors.Is(block.Err, sarama.ErrUnknownTopicOrPartition):
					// The topic was deleted after dropDeletedTopics ran.
					k.Log.Debugf("Skipping %s/%d: %v", tp.topic, tp.partition, block.Err)
				case isRetriable(block.Err):
					retry[tp] = struct{}{}
					lastErr[tp] = fmt.Errorf("fetching offset of %s/%d failed: %w", tp.topic, tp.partition, block.Err)
				default:
					errs = append(errs, fmt.Errorf("fetching offset of %s/%d failed: %w", tp.topic, tp.partition, block.Err))
				}
			}
		}
		pending = retry
	}

	for tp := range pending {
		errs = append(errs, lastErr[tp])
	}

	return result, errs
}

// dropDeletedTopics removes the partitions of topics that no longer exist.
func (k *KafkaConsumerLag) dropDeletedTopics(partitions map[topicPartition]struct{}) (map[topicPartition]struct{}, error) {
	if len(partitions) == 0 {
		return partitions, nil
	}

	descriptions, err := k.describeTopics(uniqueTopics(partitions))
	if err != nil {
		return partitions, fmt.Errorf("describing topics failed: %w", err)
	}

	deleted := make(map[string]struct{})
	for _, description := range descriptions {
		if errors.Is(description.Err, sarama.ErrUnknownTopicOrPartition) {
			k.Log.Debugf("Skipping deleted topic %q", description.Name)
			deleted[description.Name] = struct{}{}
		}
	}
	if len(deleted) == 0 {
		return partitions, nil
	}

	remaining := make(map[topicPartition]struct{}, len(partitions))
	for tp := range partitions {
		if _, ok := deleted[tp.topic]; !ok {
			remaining[tp] = struct{}{}
		}
	}
	return remaining, nil
}

// describeTopics fetches the metadata of the given topics from any broker.
func (k *KafkaConsumerLag) describeTopics(topics []string) ([]*sarama.TopicMetadata, error) {
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

type offsetBatch struct {
	broker     *sarama.Broker
	request    *sarama.OffsetRequest
	partitions []topicPartition
}

// batchByLeader groups the partitions into one offset request per leader broker.
func (k *KafkaConsumerLag) batchByLeader(
	partitions, retry map[topicPartition]struct{},
	lastErr map[topicPartition]error,
) map[int32]*offsetBatch {
	batches := make(map[int32]*offsetBatch)
	for tp := range partitions {
		leader, err := k.client.Leader(tp.topic, tp.partition)
		if errors.Is(err, sarama.ErrUnknownTopicOrPartition) {
			k.Log.Debugf("Skipping %s/%d: %v", tp.topic, tp.partition, err)
			continue
		}

		if err != nil {
			retry[tp] = struct{}{}
			lastErr[tp] = fmt.Errorf("finding leader of %s/%d failed: %w", tp.topic, tp.partition, err)
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

func (k *KafkaConsumerLag) emit(
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

		if k.emitPartition {
			tags := map[string]string{
				"group":     group,
				"topic":     tp.topic,
				"partition": strconv.FormatInt(int64(tp.partition), 10),
			}
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

	if k.emitTopic {
		for topic, aggregate := range topicAggregates {
			tags := map[string]string{
				"group": group,
				"topic": topic,
			}
			fields := map[string]any{
				"lag_sum":    aggregate.sum,
				"lag_max":    aggregate.max,
				"partitions": aggregate.partitions,
			}
			acc.AddGauge("kafka_consumer_group_topic", fields, tags)
		}
	}

	if k.emitGroup && groupAggregate.partitions > 0 {
		tags := map[string]string{"group": group}
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

// logEndOffsetOf extracts the offset of a ListOffsets response block for
// both the v0 (offset list) and v1+ (single offset) response formats.
func logEndOffsetOf(block *sarama.OffsetResponseBlock) int64 {
	if len(block.Offsets) > 0 {
		return block.Offsets[0]
	}
	return block.Offset
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

func uniqueTopics(partitions map[topicPartition]struct{}) []string {
	seen := make(map[string]struct{})
	topics := make([]string, 0, len(partitions))
	for tp := range partitions {
		if _, ok := seen[tp.topic]; ok {
			continue
		}
		seen[tp.topic] = struct{}{}
		topics = append(topics, tp.topic)
	}
	return topics
}

func init() {
	inputs.Add("kafka_consumer_lag", func() telegraf.Input {
		return &KafkaConsumerLag{
			MetricLevels:        []string{levelPartition, levelTopic, levelGroup},
			CoordinatorBrokerID: -1,
		}
	})
}
