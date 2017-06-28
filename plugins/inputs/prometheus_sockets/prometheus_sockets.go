package prometheus_sockets

import (
	"github.com/influxdata/telegraf/internal"
)

// PrometheusSocketWalker gets all unix sockets form a specified path and fetches the prometheus metrics thereof
type PrometheusSocketWalker struct {
	// SocketPaths contains all directories from which sockets are harvested
	SocketPaths []string `toml:"socket_paths"`

	// URL is the path part of the prometheus metrics handler
	URL string `toml:"url_path"`

	// ResponseTimeout is the timeout of a single request
	ResponseTimeout internal.Duration `toml:"response_timeout"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
}

var sampleConfig = `
  ## An array of directories from which sockets are harvested
  socket_paths = ["/var/run/prometheus_sockets", "/tmp/sockets/prometheus"]

  # url path of the prometheus handler
  # must be the same for all sockets
  url_path = /path/to/bearer/token

  ## Specify timeout duration for slower prometheus clients (default is 3s)
  # response_timeout = "3s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

// SampleConfig returns the sample configuration
func (p *PrometheusSocketWalker) SampleConfig() string {

	return sampleConfig
}

// Description of the plugin
func (p *PrometheusSocketWalker) Description() string {
	return "Pulls prometheus metrics from one or many unix sockets in a directory"
}
