package iotservice

import (
	"errors"
	"time"
)

type MethodCall struct {
	MethodName      string                 `json:"methodName,omitempty"`
	ConnectTimeout  uint                   `json:"connectTimeoutInSeconds,omitempty"`
	ResponseTimeout uint                   `json:"responseTimeoutInSeconds,omitempty"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
}

type MethodResult struct {
	Status  int                    `json:"status,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type DeviceStatus string

const (
	Enabled  DeviceStatus = "enabled"
	Disabled DeviceStatus = "disabled"
)

type ConnectionState string

const (
	Connected    ConnectionState = "Connected"
	Disconnected ConnectionState = "Disconnected"
)

type Device struct {
	DeviceID                   string                 `json:"deviceId,omitempty"`
	GenerationID               string                 `json:"generationId,omitempty"`
	ETag                       string                 `json:"etag,omitempty"`
	ConnectionState            ConnectionState        `json:"connectionState,omitempty"`
	Status                     DeviceStatus           `json:"status,omitempty"`
	StatusReason               string                 `json:"statusReason,omitempty"`
	ConnectionStateUpdatedTime *MicrosoftTime         `json:"connectionStateUpdatedTime,omitempty"`
	StatusUpdatedTime          *MicrosoftTime         `json:"statusUpdatedTime,omitempty"`
	LastActivityTime           *MicrosoftTime         `json:"lastActivityTime,omitempty"`
	CloudToDeviceMessageCount  uint                   `json:"cloudToDeviceMessageCount,omitempty"`
	Authentication             *Authentication        `json:"authentication,omitempty"`
	Capabilities               map[string]interface{} `json:"capabilities,omitempty"`
}

type Module struct {
	ModuleID                   string          `json:"moduleId,omitempty"`
	DeviceID                   string          `json:"deviceId,omitempty"`
	GenerationID               string          `json:"generationId,omitempty"`
	ETag                       string          `json:"etag,omitempty"`
	ConnectionState            ConnectionState `json:"connectionState,omitempty"`
	ConnectionStateUpdatedTime *MicrosoftTime  `json:"connectionStateUpdatedTime,omitempty"`
	LastActivityTime           *MicrosoftTime  `json:"lastActivityTime,omitempty"`
	CloudToDeviceMessageCount  uint            `json:"cloudToDeviceMessageCount,omitempty"`
	Authentication             *Authentication `json:"authentication,omitempty"`
	ManagedBy                  string          `json:"managedBy,omitempty"`
}

type Authentication struct {
	SymmetricKey   *SymmetricKey   `json:"symmetricKey,omitempty"`
	X509Thumbprint *X509Thumbprint `json:"x509Thumbprint,omitempty"`
	Type           AuthType        `json:"type,omitempty"`
}

// AuthType is device authentication type.
type AuthType string

const (
	// AuthSAS uses symmetric keys to sign requests.
	AuthSAS AuthType = "sas"

	// AuthSelfSigned self signed certificate with a thumbprint.
	AuthSelfSigned AuthType = "selfSigned"

	// AuthCA certificate signed by a registered certificate authority.
	AuthCA AuthType = "certificateAuthority"
)

type X509Thumbprint struct {
	PrimaryThumbprint   string `json:"primaryThumbprint,omitempty"`
	SecondaryThumbprint string `json:"secondaryThumbprint,omitempty"`
}

type SymmetricKey struct {
	PrimaryKey   string `json:"primaryKey,omitempty"`
	SecondaryKey string `json:"secondaryKey,omitempty"`
}

type Twin struct {
	DeviceID                  string                 `json:"deviceId,omitempty"`
	ETag                      string                 `json:"etag,omitempty"`
	DeviceETag                string                 `json:"deviceEtag,omitempty"`
	Status                    DeviceStatus           `json:"status,omitempty"`
	StatusReason              string                 `json:"statusReason,omitempty"`
	StatusUpdateTime          *MicrosoftTime         `json:"statusUpdateTime,omitempty"`
	ConnectionState           ConnectionState        `json:"connectionState,omitempty"`
	LastActivityTime          *MicrosoftTime         `json:"lastActivityTime,omitempty"`
	CloudToDeviceMessageCount uint                   `json:"cloudToDeviceMessageCount,omitempty"`
	AuthenticationType        string                 `json:"authenticationType,omitempty"`
	X509Thumbprint            *X509Thumbprint        `json:"x509Thumbprint,omitempty"`
	Version                   int                    `json:"version,omitempty"`
	Tags                      map[string]interface{} `json:"tags,omitempty"`
	Properties                *Properties            `json:"properties,omitempty"`
	Capabilities              map[string]interface{} `json:"capabilities,omitempty"`
}

type ModuleTwin struct {
	DeviceID           string          `json:"deviceId,omitempty"`
	ModuleID           string          `json:"moduleId,omitempty"`
	ETag               string          `json:"etag,omitempty"`
	DeviceETag         string          `json:"deviceEtag,omitempty"`
	Status             DeviceStatus    `json:"status,omitempty"`
	StatusUpdateTime   *MicrosoftTime  `json:"statusUpdateTime,omitempty"`
	ConnectionState    ConnectionState `json:"connectionState,omitempty"`
	LastActivityTime   *MicrosoftTime  `json:"lastActivityTime,omitempty"`
	AuthenticationType string          `json:"authenticationType,omitempty"`
	X509Thumbprint     *X509Thumbprint `json:"x509Thumbprint,omitempty"`
	Version            uint            `json:"version,omitempty"`
	Properties         *Properties     `json:"properties,omitempty"`
}

type Properties struct {
	Desired  map[string]interface{} `json:"desired,omitempty"`
	Reported map[string]interface{} `json:"reported,omitempty"`
}

type Stats struct {
	DisabledDeviceCount uint `json:"disabledDeviceCount,omitempty"`
	EnabledDeviceCount  uint `json:"enabledDeviceCount,omitempty"`
	TotalDeviceCount    uint `json:"totalDeviceCount,omitempty"`
}

type Configuration struct {
	ID                 string                `json:"id,omitempty"`
	SchemaVersion      string                `json:"schemaVersion,omitempty"`
	Labels             map[string]string     `json:"labels,omitempty"`
	Content            *ConfigurationContent `json:"content,omitempty"`
	TargetCondition    string                `json:"targetCondition,omitempty"`
	CreatedTimeUTC     *time.Time            `json:"createdTimeUtc,omitempty"`
	LastUpdatedTimeUTC *time.Time            `json:"lastUpdatedTimeUtc,omitempty"`
	Priority           uint                  `json:"priority,omitempty"`
	SystemMetrics      *ConfigurationMetrics `json:"systemMetrics,omitempty"`
	Metrics            *ConfigurationMetrics `json:"metrics,omitempty"`
	ETag               string                `json:"etag,omitempty"`
}

type ConfigurationContent struct {
	ModulesContent map[string]interface{} `json:"modulesContent,omitempty"`
	DeviceContent  map[string]interface{} `json:"deviceContent,omitempty"`
}

type ConfigurationMetrics struct {
	Results map[string]uint   `json:"results,omitempty"`
	Queries map[string]string `json:"queries,omitempty"`
}

type JobType string

const (
	JobExport JobType = "export"
	JobImport JobType = "import"
)

type Job struct {
	Type                   JobType `json:"type"`
	InputBlobContainerURI  string  `json:"inputBlobContainerUri"`
	OutputBlobContainerURI string  `json:"outputBlobContainerUri,omitempty"`
	ExcludeKeysInExport    bool    `json:"excludeKeysInExport"`
}

type BulkResult struct {
	IsSuccessful bool         `json:"isSuccessful"`
	Errors       []*BulkError `json:"errors"`

	// TODO: figure out the structure of a warning
	Warnings []interface{} `json:"warnings"`
}

type BulkError struct {
	DeviceID    string `json:"deviceId"`
	ErrorCode   uint   `json:"errorCode"`
	ErrorStatus string `json:"errorStatus"`
}

type JobV2 struct {
	JobID  string      `json:"jobId"`
	Type   JobV2Type   `json:"type"`
	Status JobV2Status `json:"status,omitempty"`

	CloudToDeviceMethod *DeviceMethodParams `json:"cloudToDeviceMethod,omitempty"`
	UpdateTwin          interface{}         `json:"updateTwin,omitempty"`

	QueryCondition            string    `json:"queryCondition"`
	StartTime                 time.Time `json:"startTime"`
	MaxExecutionTimeInSeconds uint      `json:"maxExecutionTimeInSeconds"`
}

type JobV2Type string

const (
	JobTypeUnknown      JobV2Type = "unknown"
	JobTypeUpdateTwin   JobV2Type = "scheduleUpdateTwin"
	JobTypeDeviceMethod JobV2Type = "scheduleDeviceMethod"
)

type JobV2Status string

const (
	JobStatusUnknown   JobV2Status = "unknown"
	JobStatusQueued    JobV2Status = "queued"
	JobStatusScheduled JobV2Status = "scheduled"
	JobStatusRunning   JobV2Status = "running"
	JobStatusCancelled JobV2Status = "cancelled"
	JobStatusCompleted JobV2Status = "completed"
)

type DeviceMethodParams struct {
	MethodName       string      `json:"methodName"`
	Payload          interface{} `json:"payload"`
	TimeoutInSeconds uint        `json:"timeoutInSeconds"`
}

type BadRequestError struct {
	Message          string `json:"Message"`
	ExceptionMessage string `json:"ExceptionMessage"`
}

func (e *BadRequestError) Error() string {
	return "bad request: " + e.Message
}

type Deployment struct {
	ID              string                `json:"id"`
	Labels          map[string]string     `json:"labels"`
	Priority        uint                  `json:"priority"`
	TargetCondition string                `json:"targetCondition"`
	Content         *DeploymentContent    `json:"content"`
	Metrics         *ConfigurationMetrics `json:"metrics"`
	ETag            string                `json:"etag"`
}

type DeploymentContent struct {
	ModulesContent map[string]interface{} `json:"modulesContent"`
}

// MicrosoftTime is a hack to parse time json attributes that
// don't follow RFC3339 and don't put timezone at the end of timestamp.
type MicrosoftTime struct {
	time.Time
}

func (t *MicrosoftTime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return errors.New("malformed time")
	}
	layout := "2006-01-02T15:04:05"
	if b[len(b)-2] == 'Z' {
		layout = "2006-01-02T15:04:05Z"
	}
	n, err := time.Parse(layout, string(b[1:len(b)-1]))
	if err != nil {
		return err
	}
	t.Time = n
	return nil
}
