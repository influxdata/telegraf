package kazoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	ErrInvalidPartitionCount    = errors.New("Number of partitions must be larger than 0")
	ErrInvalidReplicationFactor = errors.New("Replication factor must be between 1 and the number of brokers")
	ErrInvalidReplicaCount      = errors.New("All partitions must have the same number of replicas")
	ErrReplicaBrokerOverlap     = errors.New("All replicas for a partition must be on separate brokers")
	ErrInvalidBroker            = errors.New("Replica assigned to invalid broker")
	ErrMissingPartitionID       = errors.New("Partition ids must be sequential starting from 0")
	ErrDuplicatePartitionID     = errors.New("Each partition must have a unique ID")
)

// Topic interacts with Kafka's topic metadata in Zookeeper.
type Topic struct {
	Name string
	kz   *Kazoo
}

// TopicList is a type that implements the sortable interface for a list of Topic instances.
type TopicList []*Topic

// Partition interacts with Kafka's partition metadata in Zookeeper.
type Partition struct {
	topic    *Topic
	ID       int32
	Replicas []int32
}

// PartitionList is a type that implements the sortable interface for a list of Partition instances
type PartitionList []*Partition

// Topics returns a list of all registered Kafka topics.
func (kz *Kazoo) Topics() (TopicList, error) {
	root := fmt.Sprintf("%s/brokers/topics", kz.conf.Chroot)
	children, _, err := kz.conn.Children(root)
	if err != nil {
		return nil, err
	}

	result := make(TopicList, 0, len(children))
	for _, name := range children {
		result = append(result, kz.Topic(name))
	}
	return result, nil
}

// WatchTopics returns a list of all registered Kafka topics, and
// watches that list for changes.
func (kz *Kazoo) WatchTopics() (TopicList, <-chan zk.Event, error) {
	root := fmt.Sprintf("%s/brokers/topics", kz.conf.Chroot)
	children, _, c, err := kz.conn.ChildrenW(root)
	if err != nil {
		return nil, nil, err
	}

	result := make(TopicList, 0, len(children))
	for _, name := range children {
		result = append(result, kz.Topic(name))
	}
	return result, c, nil
}

// Topic returns a Topic instance for a given topic name
func (kz *Kazoo) Topic(topic string) *Topic {
	return &Topic{Name: topic, kz: kz}
}

// Exists returns true if the topic exists on the Kafka cluster.
func (t *Topic) Exists() (bool, error) {
	return t.kz.exists(t.metadataPath())
}

// Partitions returns a list of all partitions for the topic.
func (t *Topic) Partitions() (PartitionList, error) {
	value, _, err := t.kz.conn.Get(t.metadataPath())
	if err != nil {
		return nil, err
	}

	return t.parsePartitions(value)
}

// WatchPartitions returns a list of all partitions for the topic, and watches the topic for changes.
func (t *Topic) WatchPartitions() (PartitionList, <-chan zk.Event, error) {
	value, _, c, err := t.kz.conn.GetW(t.metadataPath())
	if err != nil {
		return nil, nil, err
	}

	list, err := t.parsePartitions(value)
	return list, c, err
}

// Watch watches the topic for changes.
func (t *Topic) Watch() (<-chan zk.Event, error) {
	_, _, c, err := t.kz.conn.GetW(t.metadataPath())
	if err != nil {
		return nil, err
	}

	return c, err
}

type topicMetadata struct {
	Version    int                `json:"version"`
	Partitions map[string][]int32 `json:"partitions"`
}

func (t *Topic) metadataPath() string {
	return fmt.Sprintf("%s/brokers/topics/%s", t.kz.conf.Chroot, t.Name)
}

// parsePartitions parses the JSON representation of the partitions
// that is stored as data on the topic node in Zookeeper.
func (t *Topic) parsePartitions(value []byte) (PartitionList, error) {
	var tm topicMetadata
	if err := json.Unmarshal(value, &tm); err != nil {
		return nil, err
	}

	result := make(PartitionList, len(tm.Partitions))
	for partitionNumber, replicas := range tm.Partitions {
		partitionID, err := strconv.ParseInt(partitionNumber, 10, 32)
		if err != nil {
			return nil, err
		}

		replicaIDs := make([]int32, 0, len(replicas))
		for _, r := range replicas {
			replicaIDs = append(replicaIDs, int32(r))
		}
		result[partitionID] = t.Partition(int32(partitionID), replicaIDs)
	}

	return result, nil
}

// marshalPartitions turns a PartitionList into the JSON representation
// to be stored in Zookeeper.
func (t *Topic) marshalPartitions(partitions PartitionList) ([]byte, error) {
	tm := topicMetadata{Version: 1, Partitions: make(map[string][]int32, len(partitions))}
	for _, part := range partitions {
		tm.Partitions[fmt.Sprintf("%d", part.ID)] = part.Replicas
	}
	return json.Marshal(tm)
}

// generatePartitionAssignments creates a partition list for a topic. The primary replica for
// each partition is assigned in a round-robin fashion starting at a random broker.
// Additional replicas are assigned to subsequent brokers to ensure there is no overlap
func (t *Topic) generatePartitionAssignments(brokers []int32, partitionCount int, replicationFactor int) (PartitionList, error) {
	if partitionCount <= 0 {
		return nil, ErrInvalidPartitionCount
	}
	if replicationFactor <= 0 || len(brokers) < replicationFactor {
		return nil, ErrInvalidReplicationFactor
	}

	result := make(PartitionList, partitionCount)

	brokerCount := len(brokers)
	brokerIdx := rand.Intn(brokerCount)

	for p := 0; p < partitionCount; p++ {
		partition := &Partition{topic: t, ID: int32(p), Replicas: make([]int32, replicationFactor)}

		brokerIndices := rand.Perm(len(brokers))[0:replicationFactor]

		for r := 0; r < replicationFactor; r++ {
			partition.Replicas[r] = brokers[brokerIndices[r]]
		}

		result[p] = partition
		brokerIdx = (brokerIdx + 1) % brokerCount
	}

	return result, nil
}

// validatePartitionAssignments ensures that all partitions are assigned to valid brokers,
// have the same number of replicas, and each replica is assigned to a unique broker
func (t *Topic) validatePartitionAssignments(brokers []int32, assignment PartitionList) error {
	if len(assignment) == 0 {
		return ErrInvalidPartitionCount
	}

	// get the first replica count to compare against. Every partition should have the same.
	var replicaCount int
	for _, part := range assignment {
		replicaCount = len(part.Replicas)
		break
	}
	if replicaCount == 0 {
		return ErrInvalidReplicationFactor
	}

	// ensure all ids are unique and sequential
	maxPartitionID := int32(-1)
	partitionIDmap := make(map[int32]struct{}, len(assignment))

	for _, part := range assignment {
		if part == nil {
			continue
		}
		if maxPartitionID < part.ID {
			maxPartitionID = part.ID
		}
		partitionIDmap[part.ID] = struct{}{}

		// all partitions require the same replica count
		if len(part.Replicas) != replicaCount {
			return ErrInvalidReplicaCount
		}

		rset := make(map[int32]struct{}, replicaCount)
		for _, r := range part.Replicas {
			// replica must be assigned to a valid broker
			found := false
			for _, b := range brokers {
				if r == b {
					found = true
					break
				}
			}
			if !found {
				return ErrInvalidBroker
			}
			rset[r] = struct{}{}
		}
		// broker assignments for a partition must be unique
		if len(rset) != replicaCount {
			return ErrReplicaBrokerOverlap
		}
	}

	// ensure all partitions accounted for
	if int(maxPartitionID) != len(assignment)-1 {
		return ErrMissingPartitionID
	}

	// ensure no duplicate ids
	if len(partitionIDmap) != len(assignment) {
		return ErrDuplicatePartitionID
	}

	return nil
}

// Partition returns a Partition instance for the topic.
func (t *Topic) Partition(id int32, replicas []int32) *Partition {
	return &Partition{ID: id, Replicas: replicas, topic: t}
}

type topicConfig struct {
	Version   int               `json:"version"`
	ConfigMap map[string]string `json:"config"`
}

// getConfigPath returns the zk node path for a topic's config
func (t *Topic) configPath() string {
	return fmt.Sprintf("%s/config/topics/%s", t.kz.conf.Chroot, t.Name)
}

// parseConfig parses the json representation of a topic config
// and returns the configuration values
func (t *Topic) parseConfig(data []byte) (map[string]string, error) {
	var cfg topicConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg.ConfigMap, nil
}

// marshalConfig turns a config map into the json representation
// needed for Zookeeper
func (t *Topic) marshalConfig(data map[string]string) ([]byte, error) {
	cfg := topicConfig{Version: 1, ConfigMap: data}
	if cfg.ConfigMap == nil {
		cfg.ConfigMap = make(map[string]string)
	}
	return json.Marshal(&cfg)
}

// Config returns topic-level configuration settings as a map.
func (t *Topic) Config() (map[string]string, error) {
	value, _, err := t.kz.conn.Get(t.configPath())
	if err != nil {
		return nil, err
	}

	return t.parseConfig(value)
}

// Topic returns the Topic of this partition.
func (p *Partition) Topic() *Topic {
	return p.topic
}

// Key returns a unique identifier for the partition, using the form "topic/partition".
func (p *Partition) Key() string {
	return fmt.Sprintf("%s/%d", p.topic.Name, p.ID)
}

// PreferredReplica returns the preferred replica for this partition.
func (p *Partition) PreferredReplica() int32 {
	if len(p.Replicas) > 0 {
		return p.Replicas[0]
	} else {
		return -1
	}
}

// Leader returns the broker ID of the broker that is currently the leader for the partition.
func (p *Partition) Leader() (int32, error) {
	if state, err := p.state(); err != nil {
		return -1, err
	} else {
		return state.Leader, nil
	}
}

// ISR returns the broker IDs of the current in-sync replica set for the partition
func (p *Partition) ISR() ([]int32, error) {
	if state, err := p.state(); err != nil {
		return nil, err
	} else {
		return state.ISR, nil
	}
}

func (p *Partition) UnderReplicated() (bool, error) {
	if state, err := p.state(); err != nil {
		return false, err
	} else {
		return len(state.ISR) < len(p.Replicas), nil
	}
}

func (p *Partition) UsesPreferredReplica() (bool, error) {
	if state, err := p.state(); err != nil {
		return false, err
	} else {
		return len(state.ISR) > 0 && state.ISR[0] == p.Replicas[0], nil
	}
}

// partitionState represents the partition state as it is stored as JSON
// in Zookeeper on the partition's state node.
type partitionState struct {
	Leader int32   `json:"leader"`
	ISR    []int32 `json:"isr"`
}

// state retrieves and parses the partition State
func (p *Partition) state() (partitionState, error) {
	var state partitionState
	node := fmt.Sprintf("%s/brokers/topics/%s/partitions/%d/state", p.topic.kz.conf.Chroot, p.topic.Name, p.ID)
	value, _, err := p.topic.kz.conn.Get(node)
	if err != nil {
		return state, err
	}

	if err := json.Unmarshal(value, &state); err != nil {
		return state, err
	}

	return state, nil
}

// Find returns the topic with the given name if it exists in the topic list,
// and will return `nil` otherwise.
func (tl TopicList) Find(name string) *Topic {
	for _, topic := range tl {
		if topic.Name == name {
			return topic
		}
	}
	return nil
}

func (tl TopicList) Len() int {
	return len(tl)
}

func (tl TopicList) Less(i, j int) bool {
	return tl[i].Name < tl[j].Name
}

func (tl TopicList) Swap(i, j int) {
	tl[i], tl[j] = tl[j], tl[i]
}

func (pl PartitionList) Len() int {
	return len(pl)
}

func (pl PartitionList) Less(i, j int) bool {
	return pl[i].topic.Name < pl[j].topic.Name || (pl[i].topic.Name == pl[j].topic.Name && pl[i].ID < pl[j].ID)
}

func (pl PartitionList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
