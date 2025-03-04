package kubernetes

import "time"

// summaryMetrics represents all the summary data about a particular node retrieved from a kubelet
type summaryMetrics struct {
	Node nodeMetrics  `json:"node"`
	Pods []podMetrics `json:"pods"`
}

// nodeMetrics represents detailed information about a node
type nodeMetrics struct {
	NodeName         string             `json:"nodeName"`
	SystemContainers []containerMetrics `json:"systemContainers"`
	StartTime        time.Time          `json:"startTime"`
	CPU              cpuMetrics         `json:"cpu"`
	Memory           memoryMetrics      `json:"memory"`
	Network          networkMetrics     `json:"network"`
	FileSystem       fileSystemMetrics  `json:"fs"`
	Runtime          runtimeMetrics     `json:"runtime"`
}

// containerMetrics represents the metric data collect about a container from the kubelet
type containerMetrics struct {
	Name      string            `json:"name"`
	StartTime time.Time         `json:"startTime"`
	CPU       cpuMetrics        `json:"cpu"`
	Memory    memoryMetrics     `json:"memory"`
	RootFS    fileSystemMetrics `json:"rootfs"`
	LogsFS    fileSystemMetrics `json:"logs"`
}

// runtimeMetrics contains metric data on the runtime of the system
type runtimeMetrics struct {
	ImageFileSystem fileSystemMetrics `json:"imageFs"`
}

// cpuMetrics represents the cpu usage data of a pod or node
type cpuMetrics struct {
	Time                 time.Time `json:"time"`
	UsageNanoCores       int64     `json:"usageNanoCores"`
	UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
}

// podMetrics contains metric data on a given pod
type podMetrics struct {
	PodRef     podReference       `json:"podRef"`
	StartTime  *time.Time         `json:"startTime"`
	Containers []containerMetrics `json:"containers"`
	Network    networkMetrics     `json:"network"`
	Volumes    []volumeMetrics    `json:"volume"`
}

// podReference is how a pod is identified
type podReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// memoryMetrics represents the memory metrics for a pod or node
type memoryMetrics struct {
	Time            time.Time `json:"time"`
	AvailableBytes  int64     `json:"availableBytes"`
	UsageBytes      int64     `json:"usageBytes"`
	WorkingSetBytes int64     `json:"workingSetBytes"`
	RSSBytes        int64     `json:"rssBytes"`
	PageFaults      int64     `json:"pageFaults"`
	MajorPageFaults int64     `json:"majorPageFaults"`
}

// fileSystemMetrics represents disk usage metrics for a pod or node
type fileSystemMetrics struct {
	AvailableBytes int64 `json:"availableBytes"`
	CapacityBytes  int64 `json:"capacityBytes"`
	UsedBytes      int64 `json:"usedBytes"`
}

// networkMetrics represents network usage data for a pod or node
type networkMetrics struct {
	Time     time.Time `json:"time"`
	RXBytes  int64     `json:"rxBytes"`
	RXErrors int64     `json:"rxErrors"`
	TXBytes  int64     `json:"txBytes"`
	TXErrors int64     `json:"txErrors"`
}

// volumeMetrics represents the disk usage data for a given volume
type volumeMetrics struct {
	Name           string `json:"name"`
	AvailableBytes int64  `json:"availableBytes"`
	CapacityBytes  int64  `json:"capacityBytes"`
	UsedBytes      int64  `json:"usedBytes"`
}
