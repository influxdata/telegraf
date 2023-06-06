//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package intel_baseband

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// plugin name. Exposed with all metrics
	pluginName = "intel_baseband"

	// VF Metrics
	vfCodeBlocks = "Code Blocks"
	vfDataBlock  = "Data (Bytes)"

	// Engine Metrics
	engineBlock = "Per Engine"

	// Socket extensions
	socketExtension  = ".sock"
	logFileExtension = ".log"

	// UnreachableSocketBehavior Values
	unreachableSocketBehaviorError  = "error"
	unreachableSocketBehaviorIgnore = "ignore"

	defaultAccessSocketTimeout   = config.Duration(time.Second)
	defaultWaitForTelemetryDelay = config.Duration(50 * time.Millisecond)
)

var unreachableSocketBehaviors = []string{unreachableSocketBehaviorError, unreachableSocketBehaviorIgnore}

//go:embed sample.conf
var sampleConfig string

type Baseband struct {
	// required params
	SocketPath  string `toml:"socket_path"`
	FileLogPath string `toml:"log_file_path"`

	//optional params
	UnreachableSocketBehavior string          `toml:"unreachable_socket_behavior"`
	SocketAccessTimeout       config.Duration `toml:"socket_access_timeout"`
	WaitForTelemetryDelay     config.Duration `toml:"wait_for_telemetry_delay"`

	Log      telegraf.Logger `toml:"-"`
	logConn  *LogConnector
	sockConn *SocketConnector
}

func (b *Baseband) SampleConfig() string {
	return sampleConfig
}

// Init performs one time setup of the plugin
func (b *Baseband) Init() error {
	var err error

	if b.SocketAccessTimeout < 0 {
		return fmt.Errorf("socket_access_timeout should be positive number or equal to 0 (to disable timeouts)")
	}
	if b.WaitForTelemetryDelay < 0 {
		return fmt.Errorf("wait_for_telemetry_delay should be positive number or equal to 0 (to disable delay)")
	}

	// Filling default values
	// Check UnreachableSocketBehavior
	if len(b.UnreachableSocketBehavior) == 0 {
		b.UnreachableSocketBehavior = unreachableSocketBehaviorError
	} else if err = choice.Check(b.UnreachableSocketBehavior, unreachableSocketBehaviors); err != nil {
		return fmt.Errorf("unreachable_socket_behavior: %w", err)
	}

	// Validate Socket path
	if b.SocketPath, err = b.checkFilePath(b.SocketPath, Socket); err != nil {
		return fmt.Errorf("socket_path: %w", err)
	}

	// Validate log file path
	if b.FileLogPath, err = b.checkFilePath(b.FileLogPath, Log); err != nil {
		return fmt.Errorf("log_file_path: %w", err)
	}

	// Create Log Connector
	b.logConn = newLogConnector(b.FileLogPath)

	// Create Socket Connector
	b.sockConn = newSocketConnector(b.SocketPath, b.SocketAccessTimeout, b.WaitForTelemetryDelay)
	return nil
}

func (b *Baseband) Gather(acc telegraf.Accumulator) error {
	err := b.sockConn.dumpTelemetryToLog()
	if err != nil {
		return err
	}

	// Read the log
	err = b.logConn.readLogFile()
	if err != nil {
		return err
	}

	err = b.logConn.readNumVFs()
	if err != nil {
		return fmt.Errorf("couldn't get the number of VFs: %w", err)
	}
	// b.numVFs less than 0 means that we are reading the file for the first time (or occurred discontinuity in file availability)
	if b.logConn.getNumVFs() <= 0 {
		return errors.New("error in accessing information about the amount of VF")
	}

	// rawData eg: 12 0
	if err = b.gatherVFMetric(acc, vfCodeBlocks); err != nil {
		return fmt.Errorf("couldn't get %q metric: %w", vfCodeBlocks, err)
	}

	// rawData eg: 12 0
	if err = b.gatherVFMetric(acc, vfDataBlock); err != nil {
		return fmt.Errorf("couldn't get %q metric: %w", vfDataBlock, err)
	}

	// rawData eg: 12 0 0 0 0 0
	if err = b.gatherEngineMetric(acc, engineBlock); err != nil {
		return fmt.Errorf("couldn't get %q metric: %w", engineBlock, err)
	}
	return nil
}

func (b *Baseband) gatherVFMetric(acc telegraf.Accumulator, metricName string) error {
	metrics, err := b.logConn.getMetrics(metricName)
	if err != nil {
		return fmt.Errorf("error accessing information about the metric %q: %w", metricName, err)
	}

	for _, metric := range metrics {
		if len(metric.data) != b.logConn.getNumVFs() {
			return fmt.Errorf("data is inconsistent, number of metrics in the file for %d VFs, the number of VFs read is %d",
				len(metric.data), b.logConn.numVFs)
		}

		for i := range metric.data {
			value, err := logMetricDataToValue(metric.data[i])
			if err != nil {
				return err
			}
			fields := map[string]interface{}{}
			tags := map[string]string{}

			tags["operation"] = metric.operationName
			tags["metric"] = metricNameToTagName(metricName)
			tags["vf"] = fmt.Sprintf("%v", i)
			fields["value"] = value
			acc.AddGauge(pluginName, fields, tags)
		}
	}
	return nil
}

func (b *Baseband) gatherEngineMetric(acc telegraf.Accumulator, metricName string) error {
	metrics, err := b.logConn.getMetrics(metricName)
	if err != nil {
		return fmt.Errorf("error in accessing information about the metric %q: %w", metricName, err)
	}

	for _, metric := range metrics {
		for i := range metric.data {
			value, err := logMetricDataToValue(metric.data[i])
			if err != nil {
				return err
			}
			fields := map[string]interface{}{}
			tags := map[string]string{}

			tags["operation"] = metric.operationName
			tags["metric"] = metricNameToTagName(metricName)
			tags["engine"] = fmt.Sprintf("%v", i)
			fields["value"] = value
			acc.AddGauge(pluginName, fields, tags)
		}
	}
	return nil
}

// Validate the provided path and return the clean version of it
// if UnreachableSocketBehavior = error -> return error, otherwise ignore the error
func (b *Baseband) checkFilePath(path string, fileType FileType) (resultPath string, err error) {
	if resultPath, err = validatePath(path, fileType); err != nil {
		return "", err
	}

	if err = checkFile(path, fileType); err != nil {
		if b.UnreachableSocketBehavior == unreachableSocketBehaviorError {
			return "", err
		}
		b.Log.Warn(err)
	}
	return resultPath, nil
}

func newBaseband() *Baseband {
	return &Baseband{
		SocketAccessTimeout:   defaultAccessSocketTimeout,
		WaitForTelemetryDelay: defaultWaitForTelemetryDelay,
	}
}

func init() {
	inputs.Add("intel_baseband", func() telegraf.Input {
		return newBaseband()
	})
}
