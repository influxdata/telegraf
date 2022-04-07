package docker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/choice"
	dockerint "github.com/influxdata/telegraf/internal/docker"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Docker object
type Docker struct {
	Endpoint       string
	ContainerNames []string `toml:"container_names" deprecated:"1.4.0;use 'container_name_include' instead"`

	GatherServices bool `toml:"gather_services"`

	Timeout          config.Duration
	PerDevice        bool     `toml:"perdevice" deprecated:"1.18.0;use 'perdevice_include' instead"`
	PerDeviceInclude []string `toml:"perdevice_include"`
	Total            bool     `toml:"total" deprecated:"1.18.0;use 'total_include' instead"`
	TotalInclude     []string `toml:"total_include"`
	TagEnvironment   []string `toml:"tag_env"`
	LabelInclude     []string `toml:"docker_label_include"`
	LabelExclude     []string `toml:"docker_label_exclude"`

	ContainerInclude []string `toml:"container_name_include"`
	ContainerExclude []string `toml:"container_name_exclude"`

	ContainerStateInclude []string `toml:"container_state_include"`
	ContainerStateExclude []string `toml:"container_state_exclude"`

	IncludeSourceTag bool `toml:"source_tag"`

	Log telegraf.Logger

	tlsint.ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	engineHost      string
	serverVersion   string
	filtersCreated  bool
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
}

// KB, MB, GB, TB, PB...human friendly
const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	defaultEndpoint = "unix:///var/run/docker.sock"
)

var (
	sizeRegex              = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?$`)
	containerStates        = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
	containerMetricClasses = []string{"cpu", "network", "blkio"}
	now                    = time.Now
)

func (d *Docker) Init() error {
	err := choice.CheckSlice(d.PerDeviceInclude, containerMetricClasses)
	if err != nil {
		return fmt.Errorf("error validating 'perdevice_include' setting : %v", err)
	}

	err = choice.CheckSlice(d.TotalInclude, containerMetricClasses)
	if err != nil {
		return fmt.Errorf("error validating 'total_include' setting : %v", err)
	}

	// Temporary logic needed for backwards compatibility until 'perdevice' setting is removed.
	if d.PerDevice {
		if !choice.Contains("network", d.PerDeviceInclude) {
			d.PerDeviceInclude = append(d.PerDeviceInclude, "network")
		}
		if !choice.Contains("blkio", d.PerDeviceInclude) {
			d.PerDeviceInclude = append(d.PerDeviceInclude, "blkio")
		}
	}

	// Temporary logic needed for backwards compatibility until 'total' setting is removed.
	if !d.Total {
		if choice.Contains("cpu", d.TotalInclude) {
			d.TotalInclude = []string{"cpu"}
		} else {
			d.TotalInclude = []string{}
		}
	}

	return nil
}

// Gather metrics from the docker server.
func (d *Docker) Gather(acc telegraf.Accumulator) error {
	if d.client == nil {
		c, err := d.getNewClient()
		if err != nil {
			return err
		}
		d.client = c
	}

	// Close any idle connections in the end of gathering
	defer d.client.Close()

	// Create label filters if not already created
	if !d.filtersCreated {
		err := d.createLabelFilters()
		if err != nil {
			return err
		}
		err = d.createContainerFilters()
		if err != nil {
			return err
		}
		err = d.createContainerStateFilters()
		if err != nil {
			return err
		}
		d.filtersCreated = true
	}

	// Get daemon info
	err := d.gatherInfo(acc)
	if err != nil {
		acc.AddError(err)
	}

	if d.GatherServices {
		err := d.gatherSwarmInfo(acc)
		if err != nil {
			acc.AddError(err)
		}
	}

	filterArgs := filters.NewArgs()
	for _, state := range containerStates {
		if d.stateFilter.Match(state) {
			filterArgs.Add("status", state)
		}
	}

	// All container states were excluded
	if filterArgs.Len() == 0 {
		return nil
	}

	// List containers
	opts := types.ContainerListOptions{
		Filters: filterArgs,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	containers, err := d.client.ContainerList(ctx, opts)
	if err == context.DeadlineExceeded {
		return errListTimeout
	}
	if err != nil {
		return err
	}

	// Get container data
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(c types.Container) {
			defer wg.Done()
			if err := d.gatherContainer(c, acc); err != nil {
				acc.AddError(err)
			}
		}(container)
	}
	wg.Wait()

	return nil
}

func (d *Docker) gatherSwarmInfo(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	services, err := d.client.ServiceList(ctx, types.ServiceListOptions{})
	if err == context.DeadlineExceeded {
		return errServiceTimeout
	}
	if err != nil {
		return err
	}

	if len(services) > 0 {
		tasks, err := d.client.TaskList(ctx, types.TaskListOptions{})
		if err != nil {
			return err
		}

		nodes, err := d.client.NodeList(ctx, types.NodeListOptions{})
		if err != nil {
			return err
		}

		running := map[string]int{}
		tasksNoShutdown := map[string]uint64{}

		activeNodes := make(map[string]struct{})
		for _, n := range nodes {
			if n.Status.State != swarm.NodeStateDown {
				activeNodes[n.ID] = struct{}{}
			}
		}

		for _, task := range tasks {
			if task.DesiredState != swarm.TaskStateShutdown {
				tasksNoShutdown[task.ServiceID]++
			}

			if task.Status.State == swarm.TaskStateRunning {
				running[task.ServiceID]++
			}
		}

		for _, service := range services {
			tags := map[string]string{}
			fields := make(map[string]interface{})
			now := time.Now()
			tags["service_id"] = service.ID
			tags["service_name"] = service.Spec.Name
			if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
				tags["service_mode"] = "replicated"
				fields["tasks_running"] = running[service.ID]
				fields["tasks_desired"] = *service.Spec.Mode.Replicated.Replicas
			} else if service.Spec.Mode.Global != nil {
				tags["service_mode"] = "global"
				fields["tasks_running"] = running[service.ID]
				fields["tasks_desired"] = tasksNoShutdown[service.ID]
			} else {
				d.Log.Error("Unknown replica mode")
			}
			// Add metrics
			acc.AddFields("docker_swarm",
				fields,
				tags,
				now)
		}
	}

	return nil
}

func (d *Docker) gatherInfo(acc telegraf.Accumulator) error {
	// Init vars
	dataFields := make(map[string]interface{})
	metadataFields := make(map[string]interface{})
	now := time.Now()

	// Get info from docker daemon
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	info, err := d.client.Info(ctx)
	if err == context.DeadlineExceeded {
		return errInfoTimeout
	}
	if err != nil {
		return err
	}

	d.engineHost = info.Name
	d.serverVersion = info.ServerVersion

	tags := map[string]string{
		"engine_host":    d.engineHost,
		"server_version": d.serverVersion,
	}

	fields := map[string]interface{}{
		"n_cpus":                  info.NCPU,
		"n_used_file_descriptors": info.NFd,
		"n_containers":            info.Containers,
		"n_containers_running":    info.ContainersRunning,
		"n_containers_stopped":    info.ContainersStopped,
		"n_containers_paused":     info.ContainersPaused,
		"n_images":                info.Images,
		"n_goroutines":            info.NGoroutines,
		"n_listener_events":       info.NEventsListener,
	}

	// Add metrics
	acc.AddFields("docker", fields, tags, now)
	acc.AddFields("docker",
		map[string]interface{}{"memory_total": info.MemTotal},
		tags,
		now)

	// Get storage metrics
	tags["unit"] = "bytes"

	var (
		// "docker_devicemapper" measurement fields
		poolName           string
		deviceMapperFields = map[string]interface{}{}
	)

	for _, rawData := range info.DriverStatus {
		name := strings.ToLower(strings.Replace(rawData[0], " ", "_", -1))
		if name == "pool_name" {
			poolName = rawData[1]
			continue
		}

		// Try to convert string to int (bytes)
		value, err := parseSize(rawData[1])
		if err != nil {
			continue
		}

		switch name {
		case "pool_blocksize",
			"base_device_size",
			"data_space_used",
			"data_space_total",
			"data_space_available",
			"metadata_space_used",
			"metadata_space_total",
			"metadata_space_available",
			"thin_pool_minimum_free_space":
			deviceMapperFields[name+"_bytes"] = value
		}

		// Legacy devicemapper measurements
		if name == "pool_blocksize" {
			// pool blocksize
			acc.AddFields("docker",
				map[string]interface{}{"pool_blocksize": value},
				tags,
				now)
		} else if strings.HasPrefix(name, "data_space_") {
			// data space
			fieldName := strings.TrimPrefix(name, "data_space_")
			dataFields[fieldName] = value
		} else if strings.HasPrefix(name, "metadata_space_") {
			// metadata space
			fieldName := strings.TrimPrefix(name, "metadata_space_")
			metadataFields[fieldName] = value
		}
	}

	if len(dataFields) > 0 {
		acc.AddFields("docker_data", dataFields, tags, now)
	}

	if len(metadataFields) > 0 {
		acc.AddFields("docker_metadata", metadataFields, tags, now)
	}

	if len(deviceMapperFields) > 0 {
		tags := map[string]string{
			"engine_host":    d.engineHost,
			"server_version": d.serverVersion,
		}

		if poolName != "" {
			tags["pool_name"] = poolName
		}

		acc.AddFields("docker_devicemapper", deviceMapperFields, tags, now)
	}

	return nil
}

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}

func (d *Docker) gatherContainer(
	container types.Container,
	acc telegraf.Accumulator,
) error {
	var v *types.StatsJSON

	// Parse container name
	var cname string
	for _, name := range container.Names {
		trimmedName := strings.TrimPrefix(name, "/")
		if !strings.Contains(trimmedName, "/") {
			cname = trimmedName
			break
		}
	}

	if cname == "" {
		return nil
	}

	if !d.containerFilter.Match(cname) {
		return nil
	}

	imageName, imageVersion := dockerint.ParseImage(container.Image)

	tags := map[string]string{
		"engine_host":       d.engineHost,
		"server_version":    d.serverVersion,
		"container_name":    cname,
		"container_image":   imageName,
		"container_version": imageVersion,
	}

	if d.IncludeSourceTag {
		tags["source"] = hostnameFromID(container.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	r, err := d.client.ContainerStats(ctx, container.ID, false)
	if err == context.DeadlineExceeded {
		return errStatsTimeout
	}
	if err != nil {
		return fmt.Errorf("error getting docker stats: %v", err)
	}

	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&v); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("error decoding: %v", err)
	}
	daemonOSType := r.OSType

	// Add labels to tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	return d.gatherContainerInspect(container, acc, tags, daemonOSType, v)
}

func (d *Docker) gatherContainerInspect(
	container types.Container,
	acc telegraf.Accumulator,
	tags map[string]string,
	daemonOSType string,
	v *types.StatsJSON,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	info, err := d.client.ContainerInspect(ctx, container.ID)
	if err == context.DeadlineExceeded {
		return errInspectTimeout
	}
	if err != nil {
		return fmt.Errorf("error inspecting docker container: %v", err)
	}

	// Add whitelisted environment variables to tags
	if len(d.TagEnvironment) > 0 {
		for _, envvar := range info.Config.Env {
			for _, configvar := range d.TagEnvironment {
				dockEnv := strings.SplitN(envvar, "=", 2)
				// check for presence of tag in whitelist
				if len(dockEnv) == 2 && len(strings.TrimSpace(dockEnv[1])) != 0 && configvar == dockEnv[0] {
					tags[dockEnv[0]] = dockEnv[1]
				}
			}
		}
	}

	if info.State != nil {
		tags["container_status"] = info.State.Status
		statefields := map[string]interface{}{
			"oomkilled":    info.State.OOMKilled,
			"pid":          info.State.Pid,
			"exitcode":     info.State.ExitCode,
			"container_id": container.ID,
		}

		finished, err := time.Parse(time.RFC3339, info.State.FinishedAt)
		if err == nil && !finished.IsZero() {
			statefields["finished_at"] = finished.UnixNano()
		} else {
			// set finished to now for use in uptime
			finished = now()
		}

		started, err := time.Parse(time.RFC3339, info.State.StartedAt)
		if err == nil && !started.IsZero() {
			statefields["started_at"] = started.UnixNano()

			uptime := finished.Sub(started)
			if finished.Before(started) {
				uptime = now().Sub(started)
			}
			statefields["uptime_ns"] = uptime.Nanoseconds()
		}

		acc.AddFields("docker_container_status", statefields, tags, now())

		if info.State.Health != nil {
			healthfields := map[string]interface{}{
				"health_status":  info.State.Health.Status,
				"failing_streak": info.ContainerJSONBase.State.Health.FailingStreak,
			}
			acc.AddFields("docker_container_health", healthfields, tags, now())
		}
	}

	d.parseContainerStats(v, acc, tags, container.ID, daemonOSType)

	return nil
}

func (d *Docker) parseContainerStats(
	stat *types.StatsJSON,
	acc telegraf.Accumulator,
	tags map[string]string,
	id string,
	daemonOSType string,
) {
	tm := stat.Read

	if tm.Before(time.Unix(0, 0)) {
		tm = time.Now()
	}

	memfields := map[string]interface{}{
		"container_id": id,
	}

	memstats := []string{
		"active_anon",
		"active_file",
		"cache",
		"hierarchical_memory_limit",
		"inactive_anon",
		"inactive_file",
		"mapped_file",
		"pgfault",
		"pgmajfault",
		"pgpgin",
		"pgpgout",
		"rss",
		"rss_huge",
		"total_active_anon",
		"total_active_file",
		"total_cache",
		"total_inactive_anon",
		"total_inactive_file",
		"total_mapped_file",
		"total_pgfault",
		"total_pgmajfault",
		"total_pgpgin",
		"total_pgpgout",
		"total_rss",
		"total_rss_huge",
		"total_unevictable",
		"total_writeback",
		"unevictable",
		"writeback",
	}
	for _, field := range memstats {
		if value, ok := stat.MemoryStats.Stats[field]; ok {
			memfields[field] = value
		}
	}
	if stat.MemoryStats.Failcnt != 0 {
		memfields["fail_count"] = stat.MemoryStats.Failcnt
	}

	if daemonOSType != "windows" {
		memfields["limit"] = stat.MemoryStats.Limit
		memfields["max_usage"] = stat.MemoryStats.MaxUsage

		mem := CalculateMemUsageUnixNoCache(stat.MemoryStats)
		memLimit := float64(stat.MemoryStats.Limit)
		memfields["usage"] = uint64(mem)
		memfields["usage_percent"] = CalculateMemPercentUnixNoCache(memLimit, mem)
	} else {
		memfields["commit_bytes"] = stat.MemoryStats.Commit
		memfields["commit_peak_bytes"] = stat.MemoryStats.CommitPeak
		memfields["private_working_set"] = stat.MemoryStats.PrivateWorkingSet
	}

	acc.AddFields("docker_container_mem", memfields, tags, tm)

	if choice.Contains("cpu", d.TotalInclude) {
		cpufields := map[string]interface{}{
			"usage_total":                  stat.CPUStats.CPUUsage.TotalUsage,
			"usage_in_usermode":            stat.CPUStats.CPUUsage.UsageInUsermode,
			"usage_in_kernelmode":          stat.CPUStats.CPUUsage.UsageInKernelmode,
			"usage_system":                 stat.CPUStats.SystemUsage,
			"throttling_periods":           stat.CPUStats.ThrottlingData.Periods,
			"throttling_throttled_periods": stat.CPUStats.ThrottlingData.ThrottledPeriods,
			"throttling_throttled_time":    stat.CPUStats.ThrottlingData.ThrottledTime,
			"container_id":                 id,
		}

		if daemonOSType != "windows" {
			previousCPU := stat.PreCPUStats.CPUUsage.TotalUsage
			previousSystem := stat.PreCPUStats.SystemUsage
			cpuPercent := CalculateCPUPercentUnix(previousCPU, previousSystem, stat)
			cpufields["usage_percent"] = cpuPercent
		} else {
			cpuPercent := calculateCPUPercentWindows(stat)
			cpufields["usage_percent"] = cpuPercent
		}

		cputags := copyTags(tags)
		cputags["cpu"] = "cpu-total"
		acc.AddFields("docker_container_cpu", cpufields, cputags, tm)
	}

	if choice.Contains("cpu", d.PerDeviceInclude) && len(stat.CPUStats.CPUUsage.PercpuUsage) > 0 {
		// If we have OnlineCPUs field, then use it to restrict stats gathering to only Online CPUs
		// (https://github.com/moby/moby/commit/115f91d7575d6de6c7781a96a082f144fd17e400)
		var percpuusage []uint64
		if stat.CPUStats.OnlineCPUs > 0 {
			percpuusage = stat.CPUStats.CPUUsage.PercpuUsage[:stat.CPUStats.OnlineCPUs]
		} else {
			percpuusage = stat.CPUStats.CPUUsage.PercpuUsage
		}

		for i, percpu := range percpuusage {
			percputags := copyTags(tags)
			percputags["cpu"] = fmt.Sprintf("cpu%d", i)
			fields := map[string]interface{}{
				"usage_total":  percpu,
				"container_id": id,
			}
			acc.AddFields("docker_container_cpu", fields, percputags, tm)
		}
	}

	totalNetworkStatMap := make(map[string]interface{})
	for network, netstats := range stat.Networks {
		netfields := map[string]interface{}{
			"rx_dropped":   netstats.RxDropped,
			"rx_bytes":     netstats.RxBytes,
			"rx_errors":    netstats.RxErrors,
			"tx_packets":   netstats.TxPackets,
			"tx_dropped":   netstats.TxDropped,
			"rx_packets":   netstats.RxPackets,
			"tx_errors":    netstats.TxErrors,
			"tx_bytes":     netstats.TxBytes,
			"container_id": id,
		}
		// Create a new network tag dictionary for the "network" tag
		if choice.Contains("network", d.PerDeviceInclude) {
			nettags := copyTags(tags)
			nettags["network"] = network
			acc.AddFields("docker_container_net", netfields, nettags, tm)
		}
		if choice.Contains("network", d.TotalInclude) {
			for field, value := range netfields {
				if field == "container_id" {
					continue
				}

				var uintV uint64
				switch v := value.(type) {
				case uint64:
					uintV = v
				case int64:
					uintV = uint64(v)
				default:
					continue
				}

				_, ok := totalNetworkStatMap[field]
				if ok {
					totalNetworkStatMap[field] = totalNetworkStatMap[field].(uint64) + uintV
				} else {
					totalNetworkStatMap[field] = uintV
				}
			}
		}
	}

	// totalNetworkStatMap could be empty if container is running with --net=host.
	if choice.Contains("network", d.TotalInclude) && len(totalNetworkStatMap) != 0 {
		nettags := copyTags(tags)
		nettags["network"] = "total"
		totalNetworkStatMap["container_id"] = id
		acc.AddFields("docker_container_net", totalNetworkStatMap, nettags, tm)
	}

	d.gatherBlockIOMetrics(acc, stat, tags, tm, id)
}

// Make a map of devices to their block io stats
func getDeviceStatMap(blkioStats types.BlkioStats) map[string]map[string]interface{} {
	deviceStatMap := make(map[string]map[string]interface{})

	for _, metric := range blkioStats.IoServiceBytesRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_service_bytes_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServicedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_serviced_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoQueuedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_queue_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServiceTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_service_time_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoWaitTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_wait_time_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoMergedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_merged_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["io_time_recursive"] = metric.Value
	}

	for _, metric := range blkioStats.SectorsRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["sectors_recursive"] = metric.Value
	}
	return deviceStatMap
}

func (d *Docker) gatherBlockIOMetrics(
	acc telegraf.Accumulator,
	stat *types.StatsJSON,
	tags map[string]string,
	tm time.Time,
	id string,
) {
	perDeviceBlkio := choice.Contains("blkio", d.PerDeviceInclude)
	totalBlkio := choice.Contains("blkio", d.TotalInclude)
	blkioStats := stat.BlkioStats
	deviceStatMap := getDeviceStatMap(blkioStats)

	totalStatMap := make(map[string]interface{})
	for device, fields := range deviceStatMap {
		fields["container_id"] = id
		if perDeviceBlkio {
			iotags := copyTags(tags)
			iotags["device"] = device
			acc.AddFields("docker_container_blkio", fields, iotags, tm)
		}
		if totalBlkio {
			for field, value := range fields {
				if field == "container_id" {
					continue
				}

				var uintV uint64
				switch v := value.(type) {
				case uint64:
					uintV = v
				case int64:
					uintV = uint64(v)
				default:
					continue
				}

				_, ok := totalStatMap[field]
				if ok {
					totalStatMap[field] = totalStatMap[field].(uint64) + uintV
				} else {
					totalStatMap[field] = uintV
				}
			}
		}
	}
	if totalBlkio {
		totalStatMap["container_id"] = id
		iotags := copyTags(tags)
		iotags["device"] = "total"
		acc.AddFields("docker_container_blkio", totalStatMap, iotags, tm)
	}
}

func copyTags(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Parses the human-readable size string into the amount it represents.
func parseSize(sizeStr string) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return -1, fmt.Errorf("invalid size: %s", sizeStr)
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	uMap := map[string]int64{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	unitPrefix := strings.ToLower(matches[3])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= float64(mul)
	}

	return int64(size), nil
}

func (d *Docker) createContainerFilters() error {
	// Backwards compatibility for deprecated `container_names` parameter.
	if len(d.ContainerNames) > 0 {
		d.ContainerInclude = append(d.ContainerInclude, d.ContainerNames...)
	}

	containerFilter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}
	d.containerFilter = containerFilter
	return nil
}

func (d *Docker) createLabelFilters() error {
	labelFilter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return err
	}
	d.labelFilter = labelFilter
	return nil
}

func (d *Docker) createContainerStateFilters() error {
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	stateFilter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return err
	}
	d.stateFilter = stateFilter
	return nil
}

func (d *Docker) getNewClient() (Client, error) {
	if d.Endpoint == "ENV" {
		return d.newEnvClient()
	}

	tlsConfig, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	return d.newClient(d.Endpoint, tlsConfig)
}

func init() {
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{
			PerDevice:        true,
			PerDeviceInclude: []string{"cpu"},
			TotalInclude:     []string{"cpu", "blkio", "network"},
			Timeout:          config.Duration(time.Second * 5),
			Endpoint:         defaultEndpoint,
			newEnvClient:     NewEnvClient,
			newClient:        NewClient,
			filtersCreated:   false,
		}
	})
}
