package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/docker"
	docker_stats "github.com/influxdata/telegraf/plugins/common/docker"
)

func (d *Docker) gatherInfo(acc telegraf.Accumulator) error {
	now := time.Now()

	// Get info from docker daemon
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	info, err := d.client.Info(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("timeout retrieving docker engine info")
		}
		return fmt.Errorf("getting info failed: %w", err)
	}

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
	acc.AddFields("docker", map[string]interface{}{"memory_total": info.MemTotal}, tags, now)

	// Get storage metrics
	tags["unit"] = "bytes"

	var poolName string
	deviceMapperFields := make(map[string]interface{}, len(info.DriverStatus))
	dataFields := make(map[string]interface{})
	metadataFields := make(map[string]interface{})
	for _, rawData := range info.DriverStatus {
		name := strings.ToLower(strings.ReplaceAll(rawData[0], " ", "_"))
		if name == "pool_name" {
			poolName = rawData[1]
			continue
		}

		// Try to convert string to int (bytes)
		value, err := parseSize(rawData[1])
		if err != nil {
			d.Log.Debugf("parsing size %q failed: %v", rawData[1], err)
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
			acc.AddFields("docker", map[string]interface{}{"pool_blocksize": value}, tags, now)
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

func (d *Docker) gatherSwarmInfo(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	services, err := d.client.ServiceList(ctx, swarm.ServiceListOptions{})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("timeout retrieving swarm service list")
		}
		return fmt.Errorf("getting service list failed: %w", err)
	}

	if len(services) > 0 {
		tasks, err := d.client.TaskList(ctx, swarm.TaskListOptions{})
		if err != nil {
			return fmt.Errorf("getting task list failed: %w", err)
		}

		tasksNoShutdown := make(map[string]uint64, len(tasks))
		running := make(map[string]int, len(tasks))
		for _, task := range tasks {
			if task.DesiredState != swarm.TaskStateShutdown {
				tasksNoShutdown[task.ServiceID]++
			}

			if task.Status.State == swarm.TaskStateRunning {
				running[task.ServiceID]++
			}
		}

		for _, service := range services {
			tags := make(map[string]string, 3)
			fields := make(map[string]interface{}, 2)
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
			} else if service.Spec.Mode.ReplicatedJob != nil {
				tags["service_mode"] = "replicated_job"
				fields["tasks_running"] = running[service.ID]
				if service.Spec.Mode.ReplicatedJob.MaxConcurrent != nil {
					fields["max_concurrent"] = *service.Spec.Mode.ReplicatedJob.MaxConcurrent
				}
				if service.Spec.Mode.ReplicatedJob.TotalCompletions != nil {
					fields["total_completions"] = *service.Spec.Mode.ReplicatedJob.TotalCompletions
				}
			} else if service.Spec.Mode.GlobalJob != nil {
				tags["service_mode"] = "global_job"
				fields["tasks_running"] = running[service.ID]
			} else {
				d.Log.Errorf("Unknown replica mode %v", service.Spec.Mode)
				continue
			}
			// Add metrics
			acc.AddFields("docker_swarm", fields, tags, now)
		}
	}

	return nil
}

func (d *Docker) gatherDiskUsage(acc telegraf.Accumulator, opts types.DiskUsageOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	du, err := d.client.DiskUsage(ctx, opts)
	if err != nil {
		return fmt.Errorf("getting disk usage failed: %w", err)
	}

	now := time.Now()

	// Layers size
	fields := map[string]interface{}{
		"layers_size": du.LayersSize,
	}

	tags := map[string]string{
		"engine_host":    d.engineHost,
		"server_version": d.serverVersion,
	}

	acc.AddFields("docker_disk_usage", fields, tags, now)

	// Containers
	for _, cntnr := range du.Containers {
		fields := map[string]interface{}{
			"size_rw":      cntnr.SizeRw,
			"size_root_fs": cntnr.SizeRootFs,
		}

		imageName, imageVersion := docker.ParseImage(cntnr.Image)

		tags := map[string]string{
			"engine_host":       d.engineHost,
			"server_version":    d.serverVersion,
			"container_name":    parseContainerName(cntnr.Names),
			"container_image":   imageName,
			"container_version": imageVersion,
		}

		if d.IncludeSourceTag {
			tags["source"] = hostnameFromID(cntnr.ID)
		}

		acc.AddFields("docker_disk_usage", fields, tags, now)
	}

	// Images
	for _, image := range du.Images {
		fields := map[string]interface{}{
			"size":        image.Size,
			"shared_size": image.SharedSize,
		}

		tags := map[string]string{
			"engine_host":    d.engineHost,
			"server_version": d.serverVersion,
			"image_id":       image.ID[7:19], // remove "sha256:" and keep the first 12 characters
		}

		if len(image.RepoTags) > 0 {
			imageName, imageVersion := docker.ParseImage(image.RepoTags[0])
			tags["image_name"] = imageName
			tags["image_version"] = imageVersion
		}

		acc.AddFields("docker_disk_usage", fields, tags, now)
	}

	// Volumes
	for _, volume := range du.Volumes {
		fields := map[string]interface{}{
			"size": volume.UsageData.Size,
		}

		tags := map[string]string{
			"engine_host":    d.engineHost,
			"server_version": d.serverVersion,
			"volume_name":    volume.Name,
		}

		acc.AddFields("docker_disk_usage", fields, tags, now)
	}

	return nil
}

func (d *Docker) gatherContainerInfo(acc telegraf.Accumulator, cntnr container.Summary) (map[string]string, error) {
	containerName := parseContainerName(cntnr.Names)
	if containerName == "" || !d.containerFilter.Match(containerName) || !d.stateFilter.Match(cntnr.State) {
		return nil, nil
	}
	imageName, imageVersion := docker.ParseImage(cntnr.Image)
	tags := map[string]string{
		"engine_host":       d.engineHost,
		"server_version":    d.serverVersion,
		"container_name":    containerName,
		"container_image":   imageName,
		"container_version": imageVersion,
	}
	if d.IncludeSourceTag {
		tags["source"] = hostnameFromID(cntnr.ID)
	}

	// Add labels to tags
	for k, label := range cntnr.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	// Inspect the container
	ctxInspect, cancelInspect := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancelInspect()

	info, err := d.client.ContainerInspect(ctxInspect, cntnr.ID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errors.New("timeout retrieving container environment")
		}
		return nil, fmt.Errorf("inspecting container failed: %w", err)
	}

	// Add whitelisted environment variables to tags
	for _, configvar := range d.TagEnvironment {
		for _, envvar := range info.Config.Env {
			dockEnv := strings.SplitN(envvar, "=", 2)
			// check for presence of tag in whitelist
			if len(dockEnv) == 2 && len(strings.TrimSpace(dockEnv[1])) != 0 && configvar == dockEnv[0] {
				tags[dockEnv[0]] = dockEnv[1]
			}
		}
	}

	if info.State != nil {
		tags["container_status"] = info.State.Status
		addStateMetric(acc, &info, tags, cntnr.ID)
		addHealthMetric(acc, &info, tags)
	}

	return tags, nil
}

func (d *Docker) gatherContainerStats(acc telegraf.Accumulator, tags map[string]string, id string) error {
	// Only parse response if we got a stats
	ctxStats, cancelStats := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancelStats()

	// Get container stats
	r, err := d.client.ContainerStats(ctxStats, id, false)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("timeout retrieving container stats")
		}
		return fmt.Errorf("getting container stats failed: %w", err)
	}
	defer r.Body.Close()

	daemonOSType := r.OSType
	var stats *container.StatsResponse
	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("error decoding stats response: %w", err)
		}
		// EOF is expected for non-running containers (e.g. exited, created);
		// continue with nil stats so inspect metrics are still collected.
	}

	// For Podman, fix the CPU stats using cache if available
	if d.isPodman && stats != nil {
		d.fixPodmanCPUStats(id, stats)
	}
	if stats == nil {
		return nil
	}

	// Fix the reading timestamp if necessary
	timestamp := stats.Read
	if timestamp.Before(time.Unix(0, 0)) {
		timestamp = time.Now()
	}

	addMemoryMetrics(acc, stats, tags, id, daemonOSType, timestamp)
	d.addCPUMetrics(acc, stats, tags, id, daemonOSType, timestamp)
	d.addNetworkMetrics(acc, stats, tags, id, timestamp)
	d.addBlockIOMetrics(acc, stats, tags, id, timestamp)

	return nil
}

func addStateMetric(acc telegraf.Accumulator, info *container.InspectResponse, tags map[string]string, id string) {
	statefields := map[string]interface{}{
		"oomkilled":     info.State.OOMKilled,
		"pid":           info.State.Pid,
		"exitcode":      info.State.ExitCode,
		"restart_count": info.RestartCount,
		"container_id":  id,
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
}

func addHealthMetric(acc telegraf.Accumulator, info *container.InspectResponse, tags map[string]string) {
	if info.State.Health == nil {
		return
	}

	healthfields := map[string]interface{}{
		"health_status":  info.State.Health.Status,
		"failing_streak": info.ContainerJSONBase.State.Health.FailingStreak,
	}
	acc.AddFields("docker_container_health", healthfields, tags, now())
}

func addMemoryMetrics(acc telegraf.Accumulator, stats *container.StatsResponse, tags map[string]string, id, daemonOSType string, ts time.Time) {
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
		if value, ok := stats.MemoryStats.Stats[field]; ok {
			memfields[field] = value
		}
	}
	if stats.MemoryStats.Failcnt != 0 {
		memfields["fail_count"] = stats.MemoryStats.Failcnt
	}

	if daemonOSType != "windows" {
		memfields["limit"] = stats.MemoryStats.Limit
		memfields["max_usage"] = stats.MemoryStats.MaxUsage

		mem := docker_stats.CalculateMemUsageUnixNoCache(stats.MemoryStats)
		memLimit := float64(stats.MemoryStats.Limit)
		memfields["usage"] = uint64(mem)
		memfields["usage_percent"] = docker_stats.CalculateMemPercentUnixNoCache(memLimit, mem)
	} else {
		memfields["commit_bytes"] = stats.MemoryStats.Commit
		memfields["commit_peak_bytes"] = stats.MemoryStats.CommitPeak
		memfields["private_working_set"] = stats.MemoryStats.PrivateWorkingSet
	}

	acc.AddFields("docker_container_mem", memfields, tags, ts)
}

func (d *Docker) addCPUMetrics(acc telegraf.Accumulator, stats *container.StatsResponse, tags map[string]string, id, daemonOSType string, ts time.Time) {
	if slices.Contains(d.TotalInclude, "cpu") {
		cpufields := map[string]interface{}{
			"usage_total":                  stats.CPUStats.CPUUsage.TotalUsage,
			"usage_in_usermode":            stats.CPUStats.CPUUsage.UsageInUsermode,
			"usage_in_kernelmode":          stats.CPUStats.CPUUsage.UsageInKernelmode,
			"usage_system":                 stats.CPUStats.SystemUsage,
			"throttling_periods":           stats.CPUStats.ThrottlingData.Periods,
			"throttling_throttled_periods": stats.CPUStats.ThrottlingData.ThrottledPeriods,
			"throttling_throttled_time":    stats.CPUStats.ThrottlingData.ThrottledTime,
			"container_id":                 id,
		}

		if daemonOSType != "windows" {
			previousCPU := stats.PreCPUStats.CPUUsage.TotalUsage
			previousSystem := stats.PreCPUStats.SystemUsage
			cpuPercent := docker_stats.CalculateCPUPercentUnix(previousCPU, previousSystem, stats)
			cpufields["usage_percent"] = cpuPercent
		} else {
			cpuPercent := docker_stats.CalculateCPUPercentWindows(stats)
			cpufields["usage_percent"] = cpuPercent
		}

		cputags := maps.Clone(tags)
		cputags["cpu"] = "cpu-total"

		acc.AddFields("docker_container_cpu", cpufields, cputags, ts)
	}

	if slices.Contains(d.PerDeviceInclude, "cpu") && len(stats.CPUStats.CPUUsage.PercpuUsage) > 0 {
		// If we have OnlineCPUs field, then use it to restrict stats gathering to only Online CPUs
		// (https://github.com/moby/moby/commit/115f91d7575d6de6c7781a96a082f144fd17e400)
		percpuusage := stats.CPUStats.CPUUsage.PercpuUsage
		if stats.CPUStats.OnlineCPUs > 0 {
			percpuusage = stats.CPUStats.CPUUsage.PercpuUsage[:stats.CPUStats.OnlineCPUs]
		}

		for i, percpu := range percpuusage {
			percputags := maps.Clone(tags)
			percputags["cpu"] = fmt.Sprintf("cpu%d", i)
			fields := map[string]interface{}{
				"usage_total":  percpu,
				"container_id": id,
			}

			acc.AddFields("docker_container_cpu", fields, percputags, ts)
		}
	}
}

func (d *Docker) addNetworkMetrics(acc telegraf.Accumulator, stats *container.StatsResponse, tags map[string]string, id string, ts time.Time) {
	totalNetworkStatMap := make(map[string]uint64)
	for network, netstats := range stats.Networks {
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
		if slices.Contains(d.PerDeviceInclude, "network") {
			nettags := maps.Clone(tags)
			nettags["network"] = network

			acc.AddFields("docker_container_net", netfields, nettags, ts)
		}

		if slices.Contains(d.TotalInclude, "network") {
			for field, raw := range netfields {
				if field == "container_id" {
					continue
				}

				var value uint64
				switch v := raw.(type) {
				case uint64:
					value = v
				case int64:
					value = uint64(v)
				default:
					continue
				}
				totalNetworkStatMap[field] += value
			}
		}
	}

	// totalNetworkStatMap could be empty if container is running with --net=host.
	if slices.Contains(d.TotalInclude, "network") && len(totalNetworkStatMap) != 0 {
		nettags := maps.Clone(tags)
		nettags["network"] = "total"

		fields := make(map[string]interface{}, len(totalNetworkStatMap)+1)
		fields["container_id"] = id
		for k, v := range totalNetworkStatMap {
			fields[k] = v
		}
		acc.AddFields("docker_container_net", fields, nettags, ts)
	}
}

func (d *Docker) addBlockIOMetrics(acc telegraf.Accumulator, stat *container.StatsResponse, tags map[string]string, id string, ts time.Time) {
	// Make a map of devices to their block io ioStats
	ioStats := make(map[string]map[string]interface{})
	for _, metric := range stat.BlkioStats.IoServiceBytesRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		if _, ok := ioStats[device]; !ok {
			ioStats[device] = make(map[string]interface{})
		}
		field := "io_service_bytes_recursive_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoServicedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		if _, ok := ioStats[device]; !ok {
			ioStats[device] = make(map[string]interface{}, 1)
		}
		field := "io_serviced_recursive_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoQueuedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := "io_queue_recursive_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoServiceTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := "io_service_time_recursive_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoWaitTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := "io_wait_time_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoMergedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := "io_merged_recursive_" + strings.ToLower(metric.Op)
		ioStats[device][field] = metric.Value
	}
	for _, metric := range stat.BlkioStats.IoTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		ioStats[device]["io_time_recursive"] = metric.Value
	}
	for _, metric := range stat.BlkioStats.SectorsRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		ioStats[device]["sectors_recursive"] = metric.Value
	}

	totalStatMap := make(map[string]uint64)
	for device, statFields := range ioStats {
		if slices.Contains(d.PerDeviceInclude, "blkio") {
			iotags := maps.Clone(tags)
			iotags["device"] = device
			fields := maps.Clone(statFields)
			fields["container_id"] = id
			acc.AddFields("docker_container_blkio", fields, iotags, ts)
		}

		if slices.Contains(d.TotalInclude, "blkio") {
			for field, raw := range statFields {
				if field == "container_id" {
					continue
				}

				var value uint64
				switch v := raw.(type) {
				case uint64:
					value = v
				case int64:
					value = uint64(v)
				default:
					continue
				}
				totalStatMap[field] += value
			}
		}
	}

	if slices.Contains(d.TotalInclude, "blkio") {
		iotags := maps.Clone(tags)
		iotags["device"] = "total"

		fields := make(map[string]interface{}, len(totalStatMap)+1)
		fields["container_id"] = id
		for k, v := range totalStatMap {
			fields[k] = v
		}
		acc.AddFields("docker_container_blkio", fields, iotags, ts)
	}
}
