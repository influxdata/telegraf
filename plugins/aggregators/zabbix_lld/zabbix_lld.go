package zabbixlld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

const (
	LLDName = "lld"
)

// ZabbixLLD stores the config of the aggregator and values ready to be, or already, sent
type ZabbixLLD struct {
	// cache stores the signature of measurements to avoid processing twice the same
	// This signature is generated with the measurement name, tag and tag values
	cache map[uint64]bool

	// receivedData stores metrics that should generate LLD in the last period
	receivedData map[HostLLD]LLDTags

	// previousReceivedData stores the receivedData of the previous period
	previousReceivedData map[HostLLD]LLDTags

	// pushCount maintains the number of pushes to known where we reach the ResetPeriod
	pushCount int

	// ResetPeriod after this number of pushes, all data is considered new.
	// The idea behind this parameter is to resend known LLDs with low freq in case
	// previous sent was not processed by Zabbix
	ResetPeriod int `toml:"reset_period"`
}

// LLDTags stores different combinations of tags seen for a particular host and LLD
type LLDTags []map[string]string

// HostLLD used as a key to identify different LLDs in the same, or different, hosts
type HostLLD struct {
	Hostname string
	LLDKey   string
}

// NewZabbixLLD return a *ZabbixLLD object initializated
func NewZabbixLLD() *ZabbixLLD {
	zl := &ZabbixLLD{}
	zl.ResetPeriod = 10
	zl.cache = make(map[uint64]bool)
	zl.receivedData = make(map[HostLLD]LLDTags)
	zl.previousReceivedData = make(map[HostLLD]LLDTags)
	return zl
}

var sampleConfig = `
  ## Time between sending LLD traps
  period = "60s"
	## Numer of executions after all LLDs are sent again
	reset_period = 10
`

// SampleConfig return an example of the config for this aggregator
func (zl *ZabbixLLD) SampleConfig() string {
	return sampleConfig
}

// Description return a one line description of this aggregator
func (zl *ZabbixLLD) Description() string {
	return "Send Zabbix lld info about metrics"
}

// Add inyect a new metric in this aggregator
func (zl *ZabbixLLD) Add(in telegraf.Metric) {
	// HashID generates an id based on the measurement, tags keys and tags values
	id := in.HashID()
	if _, ok := zl.cache[id]; !ok {
		// hit an uncached metric, store in the lld with metrics of the same kind
		zl.cache[id] = true

		host, exists := in.Tags()["host"]
		if !exists {
			log.Printf("W! Metric without host tag. Skipped. Metric: %v", in)
			return
		}
		in.RemoveTag("host")

		// This aggregator is only interested in metrics with tags (excluding "host" tag)
		if len(in.TagList()) == 0 {
			return
		}

		lldKey, err := generateLLDKey(in.Name(), in.Tags())
		if err != nil {
			log.Printf("W! Generating LLD Hash: %v", err)
			return
		}

		hostLLDKey := HostLLD{host, lldKey}

		hostLLD, exists := zl.receivedData[hostLLDKey]
		if !exists {
			hostLLD = LLDTags{}
		}
		hostLLD = append(hostLLD, in.Tags())
		zl.receivedData[hostLLDKey] = hostLLD

	}
}

// compareAndDelete returns true if the key has been already sent with the same tags.
// It compares the values of "tags" with the ones stored in p.previousReceivedData
// for the key specified.
// Also delete the key in p.previousReceivedData, to know which keys remains
func (zl *ZabbixLLD) compareAndDelete(key HostLLD, tags LLDTags) (equal bool) {
	previousTags, exists := zl.previousReceivedData[key]
	if !exists {
		return false
	}
	delete(zl.previousReceivedData, key)

	return ElementsMatch(tags, previousTags)
}

// Push ask the aggregator to generate metrics with the info accumulated
// The name of the metric will be always "lld".
// It will have only one tag, with the host.
// It will have an uniq field, with the LLD key as the key name and the JSON data as the value
// Eg.: lld,host=hostA disk.device.fstype.mode.path="{\"data\":[...
func (zl *ZabbixLLD) Push(acc telegraf.Accumulator) {
	// Iterate over the data collected in the last period.
	// Compare with the data sent the last time.
	// If different, send a new LLD.
	for key, tags := range zl.receivedData {
		// Skip already sent LLDs
		if equal := zl.compareAndDelete(key, tags); equal {
			continue
		}

		dataValues := tags.generateDataValues()

		dataValuesJSON, err := json.Marshal(dataValues)
		if err != nil {
			log.Printf("W! Marshaling to JSON LLD tags to Zabbix format: %v", err)
		}

		acc.AddFields(
			LLDName,
			map[string]interface{}{
				key.LLDKey: dataValuesJSON,
			},
			map[string]string{
				"host": key.Hostname,
			},
		)
	}

	for key := range zl.previousReceivedData {
		emptyDataValuesJSON, err := json.Marshal(map[string]interface{}{"data": []interface{}{}})
		if err != nil {
			log.Printf("W! Marshaling to JSON empty data Zabbix format: %v", err)
		}

		acc.AddFields(
			LLDName,
			map[string]interface{}{
				key.LLDKey: emptyDataValuesJSON,
			},
			map[string]string{
				"host": key.Hostname,
			},
		)
	}

	// Move current data to previousReceivedData
	// Reset receivedData and cache to receive data from the new period
	zl.previousReceivedData = zl.receivedData
	zl.receivedData = make(map[HostLLD]LLDTags)
	zl.cache = make(map[uint64]bool)

	zl.pushCount++

	// After ResetPeriod pushes, reset all status
	if zl.pushCount >= zl.ResetPeriod {
		zl.previousReceivedData = make(map[HostLLD]LLDTags)
		zl.pushCount = 0
	}
}

// Reset put the aggregator in the initial state, executed after "period"
func (zl *ZabbixLLD) Reset() {
}

// generateDataValues create the "data" structure needed by Zabbix to send LLD
// Eg.: {"data": [{"{#FOO}": "1", "{#BAR}": "2"}, {"{#FOO}": "3", "{#BAR}": "4"}]}
func (l LLDTags) generateDataValues() map[string][]map[string]string {
	values := []map[string]string{}

	for _, set := range l {
		setValues := map[string]string{}
		for tagKey, tagValue := range set {
			setValues[fmt.Sprintf("{#%s}", strings.ToUpper(tagKey))] = tagValue
		}
		values = append(values, setValues)
	}

	return map[string][]map[string]string{"data": values}
}

// lldHash returns an uniq identifier based on measurement and tags keys
// Format generated: Name.Tags_sorted
// Empty tags are ignored
func generateLLDKey(name string, tags map[string]string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty measurement name")
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("metric without tags")
	}

	tagKeys := []string{}
	for key, val := range tags {
		if val != "" {
			tagKeys = append(tagKeys, key)
		}
	}
	sort.Strings(tagKeys)

	lldKey := fmt.Sprintf("%s.%s", name, strings.Join(tagKeys, "."))

	return lldKey, nil
}

// ElementsMatch compare list of elements ignoring order
// Adapted from https://github.com/stretchr/testify/blob/85f2b59c4459e5bf57488796be8c3667cb8246d6/assert/assertions.go#L836
func ElementsMatch(listA, listB interface{}) (ok bool) {
	if isEmpty(listA) && isEmpty(listB) {
		return true
	}

	aKind := reflect.TypeOf(listA).Kind()
	bKind := reflect.TypeOf(listB).Kind()

	if aKind != reflect.Array && aKind != reflect.Slice {
		return false
	}

	if bKind != reflect.Array && bKind != reflect.Slice {
		return false
	}

	aValue := reflect.ValueOf(listA)
	bValue := reflect.ValueOf(listB)

	aLen := aValue.Len()
	bLen := bValue.Len()

	if aLen != bLen {
		return false
	}

	// Mark indexes in bValue that we already used
	visited := make([]bool, bLen)
	for i := 0; i < aLen; i++ {
		element := aValue.Index(i).Interface()
		found := false
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if ObjectsAreEqual(bValue.Index(j).Interface(), element) {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// isEmpty gets whether the specified object is considered empty or not.
func isEmpty(object interface{}) bool {

	// get nil case out of the way
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
		// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

// ObjectsAreEqual determines if two objects are considered equal.
// This function does no assertion of any kind.
func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

func init() {
	aggregators.Add("zabbix_lld", func() telegraf.Aggregator {
		return NewZabbixLLD()
	})
}
