package kazoo

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	ErrRunningInstances          = errors.New("Cannot deregister a consumergroup with running instances")
	ErrInstanceAlreadyRegistered = errors.New("Cannot register consumer instance because it already is registered")
	ErrInstanceNotRegistered     = errors.New("Cannot deregister consumer instance because it not registered")
	ErrPartitionClaimedByOther   = errors.New("Cannot claim partition: it is already claimed by another instance")
	ErrPartitionNotClaimed       = errors.New("Cannot release partition: it is not claimed by this instance")
)

// Consumergroup represents a high-level consumer that is registered in Zookeeper,
type Consumergroup struct {
	kz   *Kazoo
	Name string
}

// ConsumergroupInstance represents an instance of a Consumergroup.
type ConsumergroupInstance struct {
	cg *Consumergroup
	ID string
}

type ConsumergroupList []*Consumergroup
type ConsumergroupInstanceList []*ConsumergroupInstance

type Registration struct {
	Pattern      RegPattern     `json:"pattern"`
	Subscription map[string]int `json:"subscription"`
	Timestamp    int64          `json:"timestamp"`
	Version      RegVersion     `json:"version"`
}

type RegPattern string

const (
	RegPatternStatic    RegPattern = "static"
	RegPatternWhiteList RegPattern = "white_list"
	RegPatternBlackList RegPattern = "black_list"
)

type RegVersion int

const (
	RegDefaultVersion RegVersion = 1
)

// Consumergroups returns all the registered consumergroups
func (kz *Kazoo) Consumergroups() (ConsumergroupList, error) {
	root := fmt.Sprintf("%s/consumers", kz.conf.Chroot)
	cgs, _, err := kz.conn.Children(root)
	if err != nil {
		return nil, err
	}

	result := make(ConsumergroupList, 0, len(cgs))
	for _, cg := range cgs {
		result = append(result, kz.Consumergroup(cg))
	}
	return result, nil
}

// Consumergroup instantiates a new consumergroup.
func (kz *Kazoo) Consumergroup(name string) *Consumergroup {
	return &Consumergroup{Name: name, kz: kz}
}

// Exists checks whether the consumergroup has been registered in Zookeeper
func (cg *Consumergroup) Exists() (bool, error) {
	return cg.kz.exists(fmt.Sprintf("%s/consumers/%s", cg.kz.conf.Chroot, cg.Name))
}

// Create registers the consumergroup in zookeeper
func (cg *Consumergroup) Create() error {
	return cg.kz.mkdirRecursive(fmt.Sprintf("%s/consumers/%s", cg.kz.conf.Chroot, cg.Name))
}

// Delete removes the consumergroup from zookeeper
func (cg *Consumergroup) Delete() error {
	if instances, err := cg.Instances(); err != nil {
		return err
	} else if len(instances) > 0 {
		return ErrRunningInstances
	}

	return cg.kz.deleteRecursive(fmt.Sprintf("%s/consumers/%s", cg.kz.conf.Chroot, cg.Name))
}

// Instances returns a map of all running instances inside this consumergroup.
func (cg *Consumergroup) Instances() (ConsumergroupInstanceList, error) {
	root := fmt.Sprintf("%s/consumers/%s/ids", cg.kz.conf.Chroot, cg.Name)
	if exists, err := cg.kz.exists(root); err != nil {
		return nil, err
	} else if exists {
		cgis, _, err := cg.kz.conn.Children(root)
		if err != nil {
			return nil, err
		}

		result := make(ConsumergroupInstanceList, 0, len(cgis))
		for _, cgi := range cgis {
			result = append(result, cg.Instance(cgi))
		}
		return result, nil
	} else {
		result := make(ConsumergroupInstanceList, 0)
		return result, nil
	}
}

// WatchInstances returns a ConsumergroupInstanceList, and a channel that will be closed
// as soon the instance list changes.
func (cg *Consumergroup) WatchInstances() (ConsumergroupInstanceList, <-chan struct{}, error) {
	node := fmt.Sprintf("%s/consumers/%s/ids", cg.kz.conf.Chroot, cg.Name)
	if exists, err := cg.kz.exists(node); err != nil {
		return nil, nil, err
	} else if !exists {
		if err := cg.kz.mkdirRecursive(node); err != nil {
			return nil, nil, err
		}
	}

	cgis, _, c, err := cg.kz.conn.ChildrenW(node)
	if err != nil {
		return nil, nil, err
	}

	result := make(ConsumergroupInstanceList, 0, len(cgis))
	for _, cgi := range cgis {
		result = append(result, cg.Instance(cgi))
	}

	channel := make(chan struct{})
	go func() {
		<-c
		close(channel)
	}()

	return result, channel, nil
}

// NewInstance instantiates a new ConsumergroupInstance inside this consumer group,
// using a newly generated ID.
func (cg *Consumergroup) NewInstance() *ConsumergroupInstance {
	id, err := generateConsumerInstanceID()
	if err != nil {
		panic(err)
	}
	return cg.Instance(id)
}

// Instance instantiates a new ConsumergroupInstance inside this consumer group,
// using an existing ID.
func (cg *Consumergroup) Instance(id string) *ConsumergroupInstance {
	return &ConsumergroupInstance{cg: cg, ID: id}
}

// PartitionOwner returns the ConsumergroupInstance that has claimed the given partition.
// This can be nil if nobody has claime dit yet.
func (cg *Consumergroup) PartitionOwner(topic string, partition int32) (*ConsumergroupInstance, error) {
	node := fmt.Sprintf("%s/consumers/%s/owners/%s/%d", cg.kz.conf.Chroot, cg.Name, topic, partition)
	val, _, err := cg.kz.conn.Get(node)

	// If the node does not exists, nobody has claimed it.
	switch err {
	case nil:
		return &ConsumergroupInstance{cg: cg, ID: string(val)}, nil
	case zk.ErrNoNode:
		return nil, nil
	default:
		return nil, err
	}
}

// Registered checks whether the consumergroup instance is registered in Zookeeper.
func (cgi *ConsumergroupInstance) Registered() (bool, error) {
	node := fmt.Sprintf("%s/consumers/%s/ids/%s", cgi.cg.kz.conf.Chroot, cgi.cg.Name, cgi.ID)
	return cgi.cg.kz.exists(node)
}

// Registered returns current registration of the consumer group instance.
func (cgi *ConsumergroupInstance) Registration() (*Registration, error) {
	node := fmt.Sprintf("%s/consumers/%s/ids/%s", cgi.cg.kz.conf.Chroot, cgi.cg.Name, cgi.ID)
	val, _, err := cgi.cg.kz.conn.Get(node)
	if err != nil {
		return nil, err
	}

	reg := &Registration{}
	if err := json.Unmarshal(val, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// Register registers the consumergroup instance in Zookeeper.
func (cgi *ConsumergroupInstance) Register(topics []string) error {
	if exists, err := cgi.Registered(); err != nil {
		return err
	} else if exists {
		return ErrInstanceAlreadyRegistered
	}

	subscription := make(map[string]int)
	for _, topic := range topics {
		subscription[topic] = 1
	}
	data, err := json.Marshal(&Registration{
		Pattern:      RegPatternStatic,
		Subscription: subscription,
		Timestamp:    time.Now().Unix(),
		Version:      RegDefaultVersion,
	})
	if err != nil {
		return err
	}

	// Create an ephemeral node for the the consumergroup instance.
	node := fmt.Sprintf("%s/consumers/%s/ids/%s", cgi.cg.kz.conf.Chroot, cgi.cg.Name, cgi.ID)
	return cgi.cg.kz.create(node, data, true)
}

// Deregister removes the registration of the instance from zookeeper.
func (cgi *ConsumergroupInstance) Deregister() error {
	node := fmt.Sprintf("%s/consumers/%s/ids/%s", cgi.cg.kz.conf.Chroot, cgi.cg.Name, cgi.ID)
	exists, stat, err := cgi.cg.kz.conn.Exists(node)
	if err != nil {
		return err
	} else if !exists {
		return ErrInstanceNotRegistered
	}

	return cgi.cg.kz.conn.Delete(node, stat.Version)
}

// Claim claims a topic/partition ownership for a consumer ID within a group. If the
// partition is already claimed by another running instance, it will return ErrAlreadyClaimed.
func (cgi *ConsumergroupInstance) ClaimPartition(topic string, partition int32) error {
	root := fmt.Sprintf("%s/consumers/%s/owners/%s", cgi.cg.kz.conf.Chroot, cgi.cg.Name, topic)
	if err := cgi.cg.kz.mkdirRecursive(root); err != nil {
		return err
	}

	// Create an ephemeral node for the partition to claim the partition for this instance
	node := fmt.Sprintf("%s/%d", root, partition)
	err := cgi.cg.kz.create(node, []byte(cgi.ID), true)
	switch err {
	case zk.ErrNodeExists:
		data, _, err := cgi.cg.kz.conn.Get(node)
		if err != nil {
			return err
		}
		if string(data) != cgi.ID {
			// Return a separate error for this, to allow for implementing a retry mechanism.
			return ErrPartitionClaimedByOther
		}
		return nil
	default:
		return err
	}
}

// ReleasePartition releases a claim to a partition.
func (cgi *ConsumergroupInstance) ReleasePartition(topic string, partition int32) error {
	owner, err := cgi.cg.PartitionOwner(topic, partition)
	if err != nil {
		return err
	}
	if owner == nil || owner.ID != cgi.ID {
		return ErrPartitionNotClaimed
	}

	node := fmt.Sprintf("%s/consumers/%s/owners/%s/%d", cgi.cg.kz.conf.Chroot, cgi.cg.Name, topic, partition)
	return cgi.cg.kz.conn.Delete(node, 0)
}

// Topics retrieves the list of topics the consumergroup has claimed ownership of at some point.
func (cg *Consumergroup) Topics() (TopicList, error) {
	root := fmt.Sprintf("%s/consumers/%s/owners", cg.kz.conf.Chroot, cg.Name)
	children, _, err := cg.kz.conn.Children(root)
	if err != nil {
		return nil, err
	}

	result := make(TopicList, 0, len(children))
	for _, name := range children {
		result = append(result, cg.kz.Topic(name))
	}
	return result, nil
}

// CommitOffset commits an offset to a group/topic/partition
func (cg *Consumergroup) CommitOffset(topic string, partition int32, offset int64) error {
	node := fmt.Sprintf("%s/consumers/%s/offsets/%s/%d", cg.kz.conf.Chroot, cg.Name, topic, partition)
	data := []byte(fmt.Sprintf("%d", offset))

	_, stat, err := cg.kz.conn.Get(node)
	switch err {
	case zk.ErrNoNode: // Create a new node
		return cg.kz.create(node, data, false)

	case nil: // Update the existing node
		_, err := cg.kz.conn.Set(node, data, stat.Version)
		return err

	default:
		return err
	}
}

// FetchOffset retrieves an offset to a group/topic/partition
func (cg *Consumergroup) FetchOffset(topic string, partition int32) (int64, error) {
	node := fmt.Sprintf("%s/consumers/%s/offsets/%s/%d", cg.kz.conf.Chroot, cg.Name, topic, partition)
	val, _, err := cg.kz.conn.Get(node)
	if err == zk.ErrNoNode {
		return 0, nil
	} else if err != nil {
		return -1, err
	}
	return strconv.ParseInt(string(val), 10, 64)
}

// generateUUID Generates a UUIDv4.
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// generateConsumerInstanceID generates a consumergroup Instance ID
// that is almost certain to be unique.
func generateConsumerInstanceID() (string, error) {
	uuid, err := generateUUID()
	if err != nil {
		return "", err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", hostname, uuid), nil
}

// Find returns the consumergroup with the given name if it exists in the list.
// Otherwise it will return `nil`.
func (cgl ConsumergroupList) Find(name string) *Consumergroup {
	for _, cg := range cgl {
		if cg.Name == name {
			return cg
		}
	}
	return nil
}

func (cgl ConsumergroupList) Len() int {
	return len(cgl)
}

func (cgl ConsumergroupList) Less(i, j int) bool {
	return cgl[i].Name < cgl[j].Name
}

func (cgl ConsumergroupList) Swap(i, j int) {
	cgl[i], cgl[j] = cgl[j], cgl[i]
}

// Find returns the consumergroup instance with the given ID if it exists in the list.
// Otherwise it will return `nil`.
func (cgil ConsumergroupInstanceList) Find(id string) *ConsumergroupInstance {
	for _, cgi := range cgil {
		if cgi.ID == id {
			return cgi
		}
	}
	return nil
}

func (cgil ConsumergroupInstanceList) Len() int {
	return len(cgil)
}

func (cgil ConsumergroupInstanceList) Less(i, j int) bool {
	return cgil[i].ID < cgil[j].ID
}

func (cgil ConsumergroupInstanceList) Swap(i, j int) {
	cgil[i], cgil[j] = cgil[j], cgil[i]
}
