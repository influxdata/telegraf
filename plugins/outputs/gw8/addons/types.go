package addons

// MonitorStatusWeightService defines weight of Monitor Status for multi-state comparison
var MonitorStatusWeightService = map[MonitorStatus]int{
	ServiceOk:                  0,
	ServicePending:             10,
	ServiceUnknown:             20,
	ServiceWarning:             30,
	ServiceScheduledCritical:   50,
	ServiceUnscheduledCritical: 100,
}

// MetricSampleType defines TimeSeries Metric Sample Possible Types
type MetricSampleType string

// TimeSeries Metric Sample Possible Types
const (
	Value    MetricSampleType = "Value"
	Warning                   = "Warning"
	Critical                  = "Critical"
	Min                       = "Min"
	Max                       = "Max"
)

// UnitType - Supported units are a subset of The Unified Code for Units of Measure
// (http://unitsofmeasure.org/ucum.html) standard, added as we encounter
// the need for them in monitoring contexts.
type UnitType string

// Supported units
const (
	UnitCounter UnitType = "1"
	PercentCPU           = "%{cpu}"
	KB                   = "KB"
	MB                   = "MB"
	GB                   = "GB"
)

// ComputeType defines CloudHub Compute Types
type ComputeType string

// CloudHub Compute Types
const (
	Query         ComputeType = "Query"
	Regex                     = "Regex"
	Synthetic                 = "Synthetic"
	Informational             = "Informational"
	Performance               = "Performance"
	Health                    = "Health"
)

// ValueType defines the data type of the value of a metric
type ValueType string

// Data type of the value of a metric
const (
	IntegerType     ValueType = "IntegerType"
	DoubleType                = "DoubleType"
	StringType                = "StringType"
	BooleanType               = "BooleanType"
	TimeType                  = "TimeType"
	UnspecifiedType           = "UnspecifiedType"
)

type MonitorStatus string

const (
	ServiceOk                  MonitorStatus = "SERVICE_OK"
	ServiceWarning             MonitorStatus = "SERVICE_WARNING"
	ServiceUnscheduledCritical MonitorStatus = "SERVICE_UNSCHEDULED_CRITICAL"
	ServicePending             MonitorStatus = "SERVICE_PENDING"
	ServiceScheduledCritical   MonitorStatus = "SERVICE_SCHEDULED_CRITICAL"
	ServiceUnknown             MonitorStatus = "SERVICE_UNKNOWN"
	HostUp                     MonitorStatus = "HOST_UP"
	HostUnscheduledDown        MonitorStatus = "HOST_UNSCHEDULED_DOWN"
	HostPending                MonitorStatus = "HOST_PENDING"
	HostScheduledDown          MonitorStatus = "HOST_SCHEDULED_DOWN"
	HostUnreachable            MonitorStatus = "HOST_UNREACHABLE"
	HostUnchanged              MonitorStatus = "HOST_UNCHANGED"
)

// ResourceType defines the resource type
type ResourceType string

// The resource type uniquely defining the resource type
// General Nagios Types are host and service, whereas CloudHub can have richer complexity
const (
	Host           ResourceType = "host"
	Hypervisor                  = "hypervisor"
	Instance                    = "instance"
	VirtualMachine              = "virtual-machine"
	CloudApp                    = "cloud-app"
	CloudFunction               = "cloud-function"
	LoadBalancer                = "load-balancer"
	Container                   = "container"
	Storage                     = "storage"
	Network                     = "network"
	NetworkSwitch               = "network-switch"
	NetworkDevice               = "network-device"
)

// TypedValue defines a single strongly-typed value.
type TypedValue struct {
	ValueType ValueType `json:"valueType"`

	// BoolValue: A Boolean value: true or false.
	BoolValue bool `json:"boolValue,omitempty"`

	// DoubleValue: A 64-bit double-precision floating-point number. Its
	// magnitude is approximately &plusmn;10<sup>&plusmn;300</sup> and it
	// has 16 significant digits of precision.
	DoubleValue float64 `json:"doubleValue"`

	// Int64Value: A 64-bit integer. Its range is approximately
	// &plusmn;9.2x10<sup>18</sup>.
	IntegerValue int64 `json:"integerValue"`

	// StringValue: A variable-length string value.
	StringValue string `json:"stringValue,omitempty"`

	// a time stored as full timestamp
	TimeValue *MillisecondTimestamp `json:"timeValue,omitempty"`
}

// TimeInterval defines a closed time interval. It extends from the start time
// to the end time, and includes both: [startTime, endTime]. Valid time
// intervals depend on the MetricKind of the metric value. In no case
// can the end time be earlier than the start time.
// For a GAUGE metric, the StartTime value is technically optional; if
// no value is specified, the start time defaults to the value of the
// end time, and the interval represents a single point in time. Such an
// interval is valid only for GAUGE metrics, which are point-in-time
// measurements.
// For DELTA and CUMULATIVE metrics, the start time must be earlier
// than the end time.
// In all cases, the start time of the next interval must be at least a
// microsecond after the end time of the previous interval.  Because the
// interval is closed, if the start time of a new interval is the same
// as the end time of the previous interval, data written at the new
// start time could overwrite data written at the previous end time.
type TimeInterval struct {
	// EndTime: Required. The end of the time interval.
	EndTime MillisecondTimestamp `json:"endTime,omitempty"`

	// StartTime: Optional. The beginning of the time interval. The default
	// value for the start time is the end time. The start time must not be
	// later than the end time.
	StartTime MillisecondTimestamp `json:"startTime,omitempty"`
}

type DynamicMonitoredResource struct {
	// The unique name of the resource
	Name string `json:"name,required"`
	// Type: Required. The resource type of the resource
	// General Nagios Types are hosts, whereas CloudHub can have richer complexity
	Type ResourceType `json:"type,required"`
	// Owner relationship for associations like hypervisor->virtual machine
	Owner string `json:"owner,omitempty"`
	// CloudHub Categorization of resources
	Category string `json:"category,omitempty"`
	// Optional description of this resource, such as Nagios notes
	Description string `json:"description,omitempty"`
	// Foundation Properties
	Properties map[string]TypedValue `json:"properties,omitempty"`
	// Device (usually IP address), leave empty if not available, will default to name
	Device string `json:"device,omitempty"`
	// Restrict to a Groundwork Monitor Status
	Status MonitorStatus `json:"status,required"`
	// The last status check time on this resource
	LastCheckTime MillisecondTimestamp `json:"lastCheckTime,omitempty"`
	// The next status check time on this resource
	NextCheckTime MillisecondTimestamp `json:"nextCheckTime,omitempty"`
	// Nagios plugin output string
	LastPlugInOutput string `json:"lastPluginOutput,omitempty"`
	// Services state collection
	Services []DynamicMonitoredService `json:"services"`
}

// A DynamicMonitoredService represents a Groundwork Service creating during a metrics scan.
// In cloud systems, services are usually modeled as a complex metric definition, with each sampled
// metric variation represented as as single metric time series.
//
// A DynamicMonitoredService contains a collection of TimeSeries Metrics.
// MonitoredService collections are attached to a DynamicMonitoredResource during a metrics scan.
type DynamicMonitoredService struct {
	// The unique name of the resource
	Name string `json:"name,required"`
	// Type: Required. The resource type of the resource
	// General Nagios Types are hosts, whereas CloudHub can have richer complexity
	Type ResourceType `json:"type,required"`
	// Owner relationship for associations like hypervisor->virtual machine
	Owner string `json:"owner,omitempty"`
	// CloudHub Categorization of resources
	Category string `json:"category,omitempty"`
	// Optional description of this resource, such as Nagios notes
	Description string `json:"description,omitempty"`
	// Foundation Properties
	Properties map[string]TypedValue `json:"properties,omitempty"`
	// Restrict to a Groundwork Monitor Status
	Status MonitorStatus `json:"status,required"`
	// The last status check time on this resource
	LastCheckTime MillisecondTimestamp `json:"lastCheckTime,omitempty"`
	// The next status check time on this resource
	NextCheckTime MillisecondTimestamp `json:"nextCheckTime,omitempty"`
	// Nagios plugin output string
	LastPlugInOutput string `json:"lastPluginOutput,omitempty"`
	// metrics
	Metrics []TimeSeries `json:"metrics"`
}

// ThresholdValue describes threshold
type ThresholdValue struct {
	SampleType MetricSampleType `json:"sampleType"`
	Label      string           `json:"label"`
	Value      *TypedValue      `json:"value"`
}

// TimeSeries defines a single Metric Sample, its time interval, and 0 or more thresholds
type TimeSeries struct {
	MetricName string           `json:"metricName"`
	SampleType MetricSampleType `json:"sampleType,omitEmpty"`
	// Interval: The time interval to which the data sample applies. For
	// GAUGE metrics, only the end time of the interval is used. For DELTA
	// metrics, the start and end time should specify a non-zero interval,
	// with subsequent samples specifying contiguous and non-overlapping
	// intervals. For CUMULATIVE metrics, the start and end time should
	// specify a non-zero interval, with subsequent samples specifying the
	// same start time and increasing end times, until an event resets the
	// cumulative value to zero and sets a new start time for the following
	// samples.
	Interval          *TimeInterval     `json:"interval"`
	Value             *TypedValue       `json:"value"`
	Tags              map[string]string `json:"tags,omitempty"`
	Unit              UnitType          `json:"unit,omitempty"`
	Thresholds        *[]ThresholdValue `json:"thresholds,omitempty"`
	MetricComputeType ComputeType       `json:"-"`
	MetricExpression  string            `json:"-"`
}

// DynamicResourcesWithServicesRequest defines SendResourcesWithMetrics payload
type DynamicResourcesWithServicesRequest struct {
	Context   *TracerContext             `json:"context,omitempty"`
	Resources []DynamicMonitoredResource `json:"resources"`
	Groups    []ResourceGroup            `json:"groups,omitempty"`
}

// ResourceGroup defines group entity
type ResourceGroup struct {
	GroupName   string                 `json:"groupName,required"`
	Type        GroupType              `json:"type,required"`
	Description string                 `json:"description,omitempty"`
	Resources   []MonitoredResourceRef `json:"resources,required"`
}

// GroupType defines the foundation group type
type GroupType string

// The group type uniquely defining corresponding foundation group type
const (
	HostGroup    GroupType = "HostGroup"
	ServiceGroup           = "ServiceGroup"
	CustomGroup            = "CustomGroup"
)

// MonitoredResourceRef references a MonitoredResource in a group collection
type MonitoredResourceRef struct {
	// The unique name of the resource
	Name string `json:"name,required"`
	// Type: Optional. The resource type uniquely defining the resource type
	// General Nagios Types are host and service, whereas CloudHub can have richer complexity
	Type ResourceType `json:"type,omitempty"`
	// Owner relationship for associations like host->service
	Owner string `json:"owner,omitempty"`
}

// TracerContext describes a Transit call
type TracerContext struct {
	AppType    string               `json:"appType"`
	AgentID    string               `json:"agentId"`
	TraceToken string               `json:"traceToken"`
	TimeStamp  MillisecondTimestamp `json:"timeStamp"`
	Version    VersionString        `json:"version"`
}

// VersionString defines type of constant
type VersionString string

// ModelVersion defines versioning
const (
	ModelVersion VersionString = "1.0.0"
)
