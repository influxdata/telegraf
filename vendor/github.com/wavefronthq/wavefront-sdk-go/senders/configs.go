package senders

const (
	defaultBatchSize          = 10000
	defaultBufferSize         = 50000
	defaultFlushInterval      = 1
	defaultProxyFlushInterval = 5
)

// Configuration for the direct ingestion sender
type DirectConfiguration struct {
	Server string // Wavefront URL of the form https://<INSTANCE>.wavefront.com
	Token  string // Wavefront API token with direct data ingestion permission

	// Optional configuration properties. Default values should suffice for most use cases.
	// override the defaults only if you wish to set higher values.

	// max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
	BatchSize int

	// size of internal buffer beyond which received data is dropped.
	// helps with handling brief increases in data and buffering on errors.
	// separate buffers are maintained per data type (metrics, spans and distributions)
	// buffers are not pre-allocated to max size and vary based on actual usage.
	// defaults to 50,000. higher values could use more memory.
	MaxBufferSize int

	// interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
	// together with batch size controls the max theoretical throughput of the sender.
	FlushIntervalSeconds int
}

// Configuration for the proxy sender
type ProxyConfiguration struct {
	Host string // the hostname of the Wavefront proxy

	// At least one port should be set below.

	MetricsPort      int // metrics port on which the proxy is listening on, typically 2878.
	DistributionPort int // distribution port on which the proxy is listening on, typically 40000.
	TracingPort      int // tracing port on which the proxy is listening on.

	FlushIntervalSeconds int // defaults to 1 second
}
