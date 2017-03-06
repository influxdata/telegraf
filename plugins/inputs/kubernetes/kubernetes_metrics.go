package kubernetes

import "time"

// SummaryMetrics represents all the summary data about a paritcular node retrieved from a kubelet
type SummaryMetrics struct {
	Node NodeMetrics  `json:"node"`
	Pods []PodMetrics `json:"pods"`
}

// NodeMetrics represents detailed information about a node
type NodeMetrics struct {
	NodeName         string             `json:"nodeName"`
	SystemContainers []ContainerMetrics `json:"systemContainers"`
	StartTime        time.Time          `json:"startTime"`
	CPU              CPUMetrics         `json:"cpu"`
	Memory           MemoryMetrics      `json:"memory"`
	Network          NetworkMetrics     `json:"network"`
	FileSystem       FileSystemMetrics  `json:"fs"`
	Runtime          RuntimeMetrics     `json:"runtime"`
}

// ContainerMetrics represents the metric data collect about a container from the kubelet
type ContainerMetrics struct {
	Name      string            `json:"name"`
	StartTime time.Time         `json:"startTime"`
	CPU       CPUMetrics        `json:"cpu"`
	Memory    MemoryMetrics     `json:"memory"`
	RootFS    FileSystemMetrics `json:"rootfs"`
	LogsFS    FileSystemMetrics `json:"logs"`
}

// RuntimeMetrics contains metric data on the runtime of the system
type RuntimeMetrics struct {
	ImageFileSystem FileSystemMetrics `json:"imageFs"`
}

// CPUMetrics represents the cpu usage data of a pod or node
type CPUMetrics struct {
	Time                 time.Time `json:"time"`
	UsageNanoCores       int64     `json:"usageNanoCores"`
	UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
}

// PodMetrics contains metric data on a given pod
type PodMetrics struct {
	PodRef     PodReference       `json:"podRef"`
	StartTime  *time.Time         `json:"startTime"`
	Containers []ContainerMetrics `json:"containers"`
	Network    NetworkMetrics     `json:"network"`
	Volumes    []VolumeMetrics    `json:"volume"`
}

// PodReference is how a pod is identified
type PodReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// MemoryMetrics represents the memory metrics for a pod or node
type MemoryMetrics struct {
	Time            time.Time `json:"time"`
	AvailableBytes  int64     `json:"availableBytes"`
	UsageBytes      int64     `json:"usageBytes"`
	WorkingSetBytes int64     `json:"workingSetBytes"`
	RSSBytes        int64     `json:"rssBytes"`
	PageFaults      int64     `json:"pageFaults"`
	MajorPageFaults int64     `json:"majorPageFaults"`
}

// FileSystemMetrics represents disk usage metrics for a pod or node
type FileSystemMetrics struct {
	AvailableBytes int64 `json:"availableBytes"`
	CapacityBytes  int64 `json:"capacityBytes"`
	UsedBytes      int64 `json:"usedBytes"`
}

// NetworkMetrics represents network usage data for a pod or node
type NetworkMetrics struct {
	Time     time.Time `json:"time"`
	RXBytes  int64     `json:"rxBytes"`
	RXErrors int64     `json:"rxErrors"`
	TXBytes  int64     `json:"txBytes"`
	TXErrors int64     `json:"txErrors"`
}

// VolumeMetrics represents the disk usage data for a given volume
type VolumeMetrics struct {
	Name           string `json:"name"`
	AvailableBytes int64  `json:"availableBytes"`
	CapacityBytes  int64  `json:"capacityBytes"`
	UsedBytes      int64  `json:"usedBytes"`
}
