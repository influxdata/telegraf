package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

const (
	lldName = "lld"
	empty   = `{"data":[]}`
)

type lldInfo struct {
	Hostname string
	Key      string
	DataHash uint64
	Data     map[uint64]map[string]string
}

func (i *lldInfo) hash() uint64 {
	ids := make([]uint64, 0, len(i.Data))
	for id := range i.Data {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	h := fnv.New64a()
	// Write cannot fail
	_ = binary.Write(h, internal.HostEndianness, lldSeriesID(i.Hostname, i.Key))
	h.Write([]byte{0})
	_ = binary.Write(h, internal.HostEndianness, ids)

	return h.Sum64()
}

func (i *lldInfo) metric(hostTag string) (telegraf.Metric, error) {
	values := make([]map[string]string, 0, len(i.Data))
	for _, v := range i.Data {
		values = append(values, v)
	}
	data := map[string]interface{}{"data": values}
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return metric.New(
		lldName,
		map[string]string{
			hostTag: i.Hostname,
		},
		map[string]interface{}{
			i.Key: buf,
		},
		time.Now(),
	), nil
}

type zabbixLLD struct {
	log telegraf.Logger

	// current is the collection of metrics added during the recent period
	current map[uint64]lldInfo

	// previous stores the hashes of metrics received during the previous period
	previous map[uint64]lldInfo

	// lastClear store the time of the last clear of the LLD data
	lastClear time.Time

	// clearInterval after this number of pushes, all data is considered new.
	// The idea behind this parameter is to resend known LLDs with low freq in case
	// previous sent was not processed by Zabbix.
	clearInterval config.Duration

	// hostTag is the name of the tag that contains the host name
	hostTag string
}

// Push returns a slice of metrics to send to Zabbix with the LLD data using the accumulated info.
// The name of the metric will be always "lld" (const).
// It will have only one tag, with the host.
// It will have an uniq field, with the LLD key as the key name and the JSON data as the value
// Eg.: lld,host=hostA disk.device.fstype.mode.path="{\"data\":[...
func (zl *zabbixLLD) Push() []telegraf.Metric {
	metrics := make([]telegraf.Metric, 0, len(zl.current))
	newPrevious := make(map[uint64]lldInfo, len(zl.current))

	// Iterate over the data collected in the closing period and determine which
	// data needs to be send. This can be done by comparing the complete data
	// hash with what was previously sent (i.e. during last push). If different,
	// send a new LLD.
	seen := make(map[uint64]bool, len(zl.current))
	for series, info := range zl.current {
		dataHash := info.hash()

		// Get the hash of the complete data and remember the data for next period
		newPrevious[series] = lldInfo{
			Hostname: info.Hostname,
			Key:      info.Key,
			DataHash: dataHash,
		}
		seen[series] = true

		// Skip already sent data
		previous, found := zl.previous[series]
		if found && previous.DataHash == dataHash {
			continue
		}

		// For unseen series or series with new tags, we should send/resend
		// the data for discovery
		m, err := info.metric(zl.hostTag)
		if err != nil {
			zl.log.Warnf("Marshaling to JSON LLD tags in Zabbix format: %v", err)
			continue
		}
		metrics = append(metrics, m)
	}

	// Check if we have seen the LLD in this cycle and send an empty LLD otherwise
	for series, info := range zl.previous {
		if seen[series] {
			continue
		}
		m := metric.New(
			lldName,
			map[string]string{
				zl.hostTag: info.Hostname,
			},
			map[string]interface{}{
				info.Key: []byte(empty),
			},
			time.Now(),
		)
		metrics = append(metrics, m)
	}

	// Replace "previous" with the data of this period or clear previous
	// if enough time has passed
	if time.Since(zl.lastClear) < time.Duration(zl.clearInterval) {
		zl.previous = newPrevious
	} else {
		zl.previous = make(map[uint64]lldInfo, len(zl.previous))
		zl.lastClear = time.Now()
	}

	// Clear current
	zl.current = make(map[uint64]lldInfo, len(zl.current))

	return metrics
}

// Add parse a metric and add it to the LLD cache.
func (zl *zabbixLLD) Add(in telegraf.Metric) error {
	// Extract all necessary information from the metric
	// Get the metric tags. The tag-list is already sorted by key name
	tagList := in.TagList()

	// Iterate over the tags and extract
	//   - the hostname contained in the host tag
	//   - the key-values of the tags WITHOUT the host tag
	//   - the LLD-key for sending the metric in the form
	//     <metric>.<tag key 1>[,<tag key 2>...,<tag key N>]
	//   - a hash for the metric
	var hostname string
	keys := make([]string, 0, len(tagList))
	data := make(map[string]string, len(tagList))
	for _, tag := range tagList {
		// Extract the host key and skip everything else
		if tag.Key == zl.hostTag {
			hostname = tag.Value
			continue
		}

		// Collect the tag keys for generating the lld-key later
		if tag.Value != "" {
			keys = append(keys, tag.Key)
		}

		// Prepare the data for lld-metric
		key := "{#" + strings.ToUpper(tag.Key) + "}"
		data[key] = tag.Value
	}

	if len(keys) == 0 {
		// Ignore metrics without tags (apart from the host tag)
		return nil
	}
	key := in.Name() + "." + strings.Join(keys, ".")

	// If hostname is not defined, use the hostname of the system
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			return fmt.Errorf("no hostname found and unable to get hostname from system: %w", err)
		}
	}

	// Try to lookup the Zabbix series in the already received metrics and
	// create a new one if not found
	series := lldSeriesID(hostname, key)
	if _, found := zl.current[series]; !found {
		zl.current[series] = lldInfo{
			Hostname: hostname,
			Key:      key,
			Data:     make(map[uint64]map[string]string),
		}
	}
	zl.current[series].Data[in.HashID()] = data

	return nil
}

func lldSeriesID(hostname, key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(hostname))
	h.Write([]byte{0})
	h.Write([]byte(key))
	h.Write([]byte{0})
	return h.Sum64()
}
