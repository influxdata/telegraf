package leofs

import (
	"bufio"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const oid = ".1.3.6.1.4.1.35450"

// For Manager Master
const defaultEndpoint = "127.0.0.1:4020"

type ServerType int

const (
	ServerTypeManagerMaster ServerType = iota
	ServerTypeManagerSlave
	ServerTypeStorage
	ServerTypeGateway
)

type LeoFS struct {
	Servers []string
}

var KeyMapping = map[ServerType][]string{
	ServerTypeManagerMaster: {
		"num_of_processes",
		"total_memory_usage",
		"system_memory_usage",
		"processes_memory_usage",
		"ets_memory_usage",
		"num_of_processes_5min",
		"total_memory_usage_5min",
		"system_memory_usage_5min",
		"processes_memory_usage_5min",
		"ets_memory_usage_5min",
		"used_allocated_memory",
		"allocated_memory",
		"used_allocated_memory_5min",
		"allocated_memory_5min",
	},
	ServerTypeManagerSlave: {
		"num_of_processes",
		"total_memory_usage",
		"system_memory_usage",
		"processes_memory_usage",
		"ets_memory_usage",
		"num_of_processes_5min",
		"total_memory_usage_5min",
		"system_memory_usage_5min",
		"processes_memory_usage_5min",
		"ets_memory_usage_5min",
		"used_allocated_memory",
		"allocated_memory",
		"used_allocated_memory_5min",
		"allocated_memory_5min",
	},
	ServerTypeStorage: {
		"num_of_processes",
		"total_memory_usage",
		"system_memory_usage",
		"processes_memory_usage",
		"ets_memory_usage",
		"num_of_processes_5min",
		"total_memory_usage_5min",
		"system_memory_usage_5min",
		"processes_memory_usage_5min",
		"ets_memory_usage_5min",
		"num_of_writes",
		"num_of_reads",
		"num_of_deletes",
		"num_of_writes_5min",
		"num_of_reads_5min",
		"num_of_deletes_5min",
		"num_of_active_objects",
		"total_objects",
		"total_size_of_active_objects",
		"total_size",
		"num_of_replication_messages",
		"num_of_sync-vnode_messages",
		"num_of_rebalance_messages",
		"used_allocated_memory",
		"allocated_memory",
		"used_allocated_memory_5min",
		"allocated_memory_5min",
	},
	ServerTypeGateway: {
		"num_of_processes",
		"total_memory_usage",
		"system_memory_usage",
		"processes_memory_usage",
		"ets_memory_usage",
		"num_of_processes_5min",
		"total_memory_usage_5min",
		"system_memory_usage_5min",
		"processes_memory_usage_5min",
		"ets_memory_usage_5min",
		"num_of_writes",
		"num_of_reads",
		"num_of_deletes",
		"num_of_writes_5min",
		"num_of_reads_5min",
		"num_of_deletes_5min",
		"count_of_cache-hit",
		"count_of_cache-miss",
		"total_of_files",
		"total_cached_size",
		"used_allocated_memory",
		"allocated_memory",
		"used_allocated_memory_5min",
		"allocated_memory_5min",
	},
}

var serverTypeMapping = map[string]ServerType{
	"4020": ServerTypeManagerMaster,
	"4021": ServerTypeManagerSlave,
	"4010": ServerTypeStorage,
	"4011": ServerTypeStorage,
	"4012": ServerTypeStorage,
	"4013": ServerTypeStorage,
	"4000": ServerTypeGateway,
	"4001": ServerTypeGateway,
}

var sampleConfig = `
  ## An array of URI to gather stats about LeoFS.
  ## Specify an ip or hostname with port. ie 127.0.0.1:4020
  servers = ["127.0.0.1:4021"]
`

func (l *LeoFS) SampleConfig() string {
	return sampleConfig
}

func (l *LeoFS) Description() string {
	return "Read metrics from a LeoFS Server via SNMP"
}

func (l *LeoFS) Gather(acc telegraf.Accumulator) error {
	if len(l.Servers) == 0 {
		l.gatherServer(defaultEndpoint, ServerTypeManagerMaster, acc)
		return nil
	}
	var wg sync.WaitGroup
	var outerr error
	for _, endpoint := range l.Servers {
		_, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("Unable to parse the address:%s, err:%s", endpoint, err)
		}
		port, err := retrieveTokenAfterColon(endpoint)
		if err != nil {
			return err
		}
		st, ok := serverTypeMapping[port]
		if !ok {
			st = ServerTypeStorage
		}
		wg.Add(1)
		go func(endpoint string, st ServerType) {
			defer wg.Done()
			outerr = l.gatherServer(endpoint, st, acc)
		}(endpoint, st)
	}
	wg.Wait()
	return outerr
}

func (l *LeoFS) gatherServer(
	endpoint string,
	serverType ServerType,
	acc telegraf.Accumulator,
) error {
	cmd := exec.Command("snmpwalk", "-v2c", "-cpublic", endpoint, oid)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Start()
	defer internal.WaitTimeout(cmd, time.Second*5)
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		return fmt.Errorf("Unable to retrieve the node name")
	}
	nodeName, err := retrieveTokenAfterColon(scanner.Text())
	if err != nil {
		return err
	}
	nodeNameTrimmed := strings.Trim(nodeName, "\"")
	tags := map[string]string{
		"node": nodeNameTrimmed,
	}
	i := 0

	fields := make(map[string]interface{})
	for scanner.Scan() {
		key := KeyMapping[serverType][i]
		val, err := retrieveTokenAfterColon(scanner.Text())
		if err != nil {
			return err
		}
		fVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("Unable to parse the value:%s, err:%s", val, err)
		}
		fields[key] = fVal
		i++
	}
	acc.AddFields("leofs", fields, tags)
	return nil
}

func retrieveTokenAfterColon(line string) (string, error) {
	tokens := strings.Split(line, ":")
	if len(tokens) != 2 {
		return "", fmt.Errorf("':' not found in the line:%s", line)
	}
	return strings.TrimSpace(tokens[1]), nil
}

func init() {
	inputs.Add("leofs", func() telegraf.Input {
		return &LeoFS{}
	})
}
