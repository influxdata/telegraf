package sematext

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"net/http"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/processors"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/sender"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/serializer"
	"github.com/influxdata/telegraf/plugins/outputs/sematext/tags"
)

const (
	defaultSematextMetricsReceiverURL = "https://spm-receiver.sematext.com"
)

// Sematext struct contains configuration read from Telegraf config and a few runtime objects.
// We'll use one separate instance of Telegraf for each monitored service. Therefore, token for particular service
// will be configured on Sematext output level
type Sematext struct {
	ReceiverURL string          `toml:"receiver_url"`
	Token       string          `toml:"token"`
	ProxyServer string          `toml:"proxy_server"`
	Username    string          `toml:"username"`
	Password    string          `toml:"password"`
	Log         telegraf.Logger `toml:"-"`
	tls.ClientConfig

	metricsURL       string
	sender           *sender.Sender
	senderConfig     *sender.Config
	serializer       serializer.MetricSerializer
	metricProcessors []processors.MetricProcessor
	batchProcessors  []processors.BatchProcessor
}

const sampleConfig = `
  ## Docs at https://sematext.com/docs/monitoring provide info about getting
  ## started with Sematext monitoring.

  ## URL of your Sematext metrics receiver. US-region metrics receiver is used
  ## in this example (it is also the default when receiver_url value is empty),
  ## but address of e.g. Sematext EU-region metrics receiver can be used
  ## instead.
  receiver_url = "https://spm-receiver.sematext.com"

  ## Token of the App to which the data is sent. Create an App of appropriate
  ## type in Sematext UI, instructions will show its token which can be used
  ## here.
  token = ""

  ## Optional TLS Config.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Optional flag for ignoring tls certificate check.
  # insecure_skip_verify = false
`

// Connect is no-op for Sematext output plugin, everything was set up before in Init() method
func (s *Sematext) Connect() error {
	return nil
}

// Close Closes the Sematext output
func (s *Sematext) Close() error {
	s.sender.Close()

	for _, mp := range s.metricProcessors {
		mp.Close()
	}

	for _, bp := range s.batchProcessors {
		bp.Close()
	}

	return nil
}

// SampleConfig Returns a sample configuration for the Sematext output
func (s *Sematext) SampleConfig() string {
	return sampleConfig
}

// Description returns the description for the Sematext output
func (s *Sematext) Description() string {
	return "Use telegraf to send metrics to Sematext"
}

// Init performs full initialization of Sematext output
func (s *Sematext) Init() error {
	if len(s.Token) == 0 {
		return fmt.Errorf("'token' is a required field for Sematext output")
	}
	if len(s.ReceiverURL) == 0 {
		s.ReceiverURL = defaultSematextMetricsReceiverURL
	}

	var proxyURL *url.URL

	if s.ProxyServer != "" {
		var err error
		proxyURL, err = url.Parse(s.ProxyServer)
		if err != nil {
			return fmt.Errorf("invalid url %s for the proxy server: %v", s.ProxyServer, err)
		}
	}

	tlsConfig, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	s.senderConfig = &sender.Config{
		ProxyURL:  proxyURL,
		Username:  s.Username,
		Password:  s.Password,
		TLSConfig: tlsConfig,
	}
	s.sender = sender.NewSender(s.senderConfig)
	s.metricsURL = s.ReceiverURL + "/write?db=metrics"

	s.initProcessors()

	s.serializer = serializer.NewMetricSerializer(s.Log)

	s.Log.Infof("Sematext output started with Token=%s, ReceiverUrl=%s, ProxyServer=%s", s.Token, s.ReceiverURL,
		s.ProxyServer)

	return nil
}

// initProcessors instantiates all metric processors that will be used to prepare metrics/tags for sending to Sematext
func (s *Sematext) initProcessors() {
	// add more processors as they are implemented
	s.metricProcessors = []processors.MetricProcessor{
		processors.NewToken(s.Token),
		processors.NewHost(s.Log),
		processors.NewHandleCounter(),
		processors.NewContainerTags(),
		processors.NewMetricType(),
	}
	s.batchProcessors = []processors.BatchProcessor{
		// rename processor has to run before metainfo processor to ensure metainfo processor uses final metric names
		processors.NewRename(),
		processors.NewHeartbeat(),
		processors.NewMetainfo(s.Log, s.Token, s.ReceiverURL, s.senderConfig),
	}
}

// Write sends metrics to Sematext backend and handles the response
func (s *Sematext) Write(metrics []telegraf.Metric) error {
	s.Log.Debugf("Sematext.Write() called with %d metrics in the slice", len(metrics))
	processedMetrics, err := s.processMetrics(metrics)

	if err != nil {
		// error means the whole batch should be discarded without sending it. To achieve that, we have to return nil
		s.Log.Errorf("error while preparing to send metrics to Sematext, the batch will be dropped: %v", err)
		return nil
	}

	s.Log.Debugf("Preparing to send %d processed metrics", len(processedMetrics))

	if len(processedMetrics) > 0 {
		body := s.serializer.Write(processedMetrics)

		s.Log.Debugf("Sending metrics to %s : %s", s.metricsURL, body)

		res, err := s.sender.Request("POST", s.metricsURL, "text/plain; charset=utf-8", body)
		if err != nil {
			// error will happen in case of e.g. network connectivity issues; it is unrelated to response code and
			// therefore it is OK to retry it
			s.Log.Errorf("error while sending to %s : %s", s.metricsURL, err.Error())
			return err
		}
		defer res.Body.Close()

		s.Log.Debugf("Sending metrics to %s, response status code: %d", s.metricsURL, res.StatusCode)

		return s.handleResponse(res)
	}

	return nil
}

func (s *Sematext) handleResponse(res *http.Response) error {
	success, badRequest := checkResponseStatus(res)

	if !success {
		errorMsg := fmt.Sprintf("received %d status code, message = %q while sending to %s",
			res.StatusCode, res.Status, s.metricsURL)

		if badRequest {
			// shouldn't be re-sent as bad request will continue to be a bad request
			s.Log.Errorf("%s - request will be dropped", errorMsg)
			return nil
		}

		// otherwise it is some temporary error from the backend and we should retry
		return fmt.Errorf(errorMsg)
	}

	return nil
}

func checkResponseStatus(res *http.Response) (success bool, badRequest bool) {
	success = res.StatusCode >= 200 && res.StatusCode < 300
	badRequest = res.StatusCode >= 400 && res.StatusCode < 500

	return success, badRequest
}

// processMetrics returns an error only when the whole batch of metrics should be discarded
// batchProcessors run first, metricProcessors follow later
func (s *Sematext) processMetrics(metrics []telegraf.Metric) ([]telegraf.Metric, error) {
	s.Log.Debugf("Starting processing of slice of %d metrics", len(metrics))

	if metricsAlreadyProcessed(metrics) {
		// in case some batch was fully processed before, we don't want to process it once again
		s.Log.Debugf("Skipping processing of already processed slice of %d metrics", len(metrics))
		return metrics, nil
	}

	for _, p := range s.batchProcessors {
		var err error
		metrics, err = p.Process(metrics)

		if err != nil {
			s.Log.Errorf("error while running batch processors in Sematext output: %v", err)
			return metrics, err
		}
	}

	processedMetrics := make([]telegraf.Metric, 0, len(metrics))

	for _, metric := range metrics {
		metricOk := true

		// don't process the metrics that were already processed before
		if !metricAlreadyProcessed(metric) {
			for _, p := range s.metricProcessors {
				err := p.Process(metric)

				if err != nil {
					// log the message, mark the metric to be skipped, skip other processors
					s.Log.Warnf("can't process metric: %s in Sematext output, error : %s", metric, err.Error())
					metricOk = false
					break
				}
			}
		}

		if metricOk {
			processedMetrics = append(processedMetrics, metric)
		}
	}

	markMetricsProcessed(processedMetrics)

	return processedMetrics, nil
}

func metricsAlreadyProcessed(metrics []telegraf.Metric) bool {
	// return that batch hasn't been processed yet if any of its metrics hasn't been processed
	for _, m := range metrics {
		if !metricAlreadyProcessed(m) {
			return false
		}
	}

	return true
}

func metricAlreadyProcessed(metric telegraf.Metric) bool {
	_, processed := metric.GetTag(tags.SematextProcessedTag)
	return processed
}

func markMetricsProcessed(metrics []telegraf.Metric) {
	for _, m := range metrics {
		m.AddTag(tags.SematextProcessedTag, tags.SematextProcessedTag)
	}
}

func init() {
	outputs.Add("sematext", func() telegraf.Output {
		return &Sematext{}
	})
}
