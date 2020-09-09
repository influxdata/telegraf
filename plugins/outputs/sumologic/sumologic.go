package sumologic

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/carbon2"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/prometheus"
)

const (
	sampleConfig = `
  ## Unique URL generated for your HTTP Metrics Source.
  ## This is the address to send metrics to.
  # url = "https://events.sumologic.net/receiver/v1/http/<UniqueHTTPCollectorCode>"

  ## Data format to be used for sending metrics.
  ## This will set the "Content-Type" header accordingly.
  ## Currently supported formats: 
  ## * graphite - for Content-Type of application/vnd.sumologic.graphite
  ## * carbon2 - for Content-Type of application/vnd.sumologic.carbon2
  ## * prometheus - for Content-Type of application/vnd.sumologic.prometheus
  ##
  ## More information can be found at:
  ## https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Metrics-to-an-HTTP-Source#content-type-headers-for-metrics
  ##
  ## NOTE:
  ## When unset, telegraf will by default use the influx serializer which is currently unsupported
  ## in HTTP Source.
  data_format = "carbon2"

  ## Timeout used for HTTP request
  # timeout = "5s"
  
  ## HTTP method, one of: "POST" or "PUT". "POST" is used by default if unset.
  # method = "POST"

  ## Max HTTP request body size in bytes before compression (if applied).
  ## By default 1MB is recommended.
  ## NOTE:
  ## Bear in mind that in some serializer a metric even though serialized to multiple
  ## lines cannot be split any further so setting this very low might not work
  ## as expected.
  # max_request_body_size = 1_000_000

  ## Additional, Sumo specific options.
  ## Full list can be found here:
  ## https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Metrics-to-an-HTTP-Source#supported-http-headers

  ## Desired source name.
  ## Useful if you want to override the source name configured for the source.
  # source_name = ""

  ## Desired host name.
  ## Useful if you want to override the source host configured for the source.
  # source_host = ""

  ## Desired source category.
  ## Useful if you want to override the source category configured for the source.
  # source_category = ""

  ## Comma-separated key=value list of dimensions to apply to every metric.
  ## Custom dimensions will allow you to query your metrics at a more granular level.
  # dimensions = ""
`

	defaultClientTimeout      = 5 * time.Second
	defaultMethod             = http.MethodPost
	defaultMaxRequestBodySize = 1_000_000

	contentTypeHeader     = "Content-Type"
	carbon2ContentType    = "application/vnd.sumologic.carbon2"
	graphiteContentType   = "application/vnd.sumologic.graphite"
	prometheusContentType = "application/vnd.sumologic.prometheus"
)

type header string

const (
	sourceNameHeader     header = `X-Sumo-Name`
	sourceHostHeader     header = `X-Sumo-Host`
	sourceCategoryHeader header = `X-Sumo-Category`
	dimensionsHeader     header = `X-Sumo-Dimensions`
)

type SumoLogic struct {
	URL               string            `toml:"url"`
	Timeout           internal.Duration `toml:"timeout"`
	Method            string            `toml:"method"`
	MaxRequstBodySize config.Size       `toml:"max_request_body_size"`

	SourceName     string `toml:"source_name"`
	SourceHost     string `toml:"source_host"`
	SourceCategory string `toml:"source_category"`
	Dimensions     string `toml:"dimensions"`

	client     *http.Client
	serializer serializers.Serializer

	err     error
	headers map[string]string
}

func (s *SumoLogic) SetSerializer(serializer serializers.Serializer) {
	if s.headers == nil {
		s.headers = make(map[string]string)
	}

	switch serializer.(type) {
	case *carbon2.Serializer:
		s.headers[contentTypeHeader] = carbon2ContentType
	case *graphite.GraphiteSerializer:
		s.headers[contentTypeHeader] = graphiteContentType
	case *prometheus.Serializer:
		s.headers[contentTypeHeader] = prometheusContentType

	default:
		s.err = errors.Errorf("unsupported serializer %T", serializer)
	}

	s.serializer = serializer
}

func (s *SumoLogic) createClient(ctx context.Context) (*http.Client, error) {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: s.Timeout.Duration,
	}, nil
}

func (s *SumoLogic) Connect() error {
	if s.err != nil {
		return errors.Wrap(s.err, "sumologic: incorrect configuration")
	}

	if s.Method == "" {
		s.Method = defaultMethod
	}
	s.Method = strings.ToUpper(s.Method)
	if s.Method != http.MethodPost && s.Method != http.MethodPut {
		return fmt.Errorf("invalid method [%s] %s", s.URL, s.Method)
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = defaultClientTimeout
	}

	client, err := s.createClient(context.Background())
	if err != nil {
		return err
	}

	s.client = client

	return nil
}

func (s *SumoLogic) Close() error {
	return s.err
}

func (s *SumoLogic) Description() string {
	return "A plugin that can transmit metrics to Sumo Logic HTTP Source"
}

func (s *SumoLogic) SampleConfig() string {
	return sampleConfig
}

func (s *SumoLogic) Write(metrics []telegraf.Metric) error {
	if s.err != nil {
		return errors.Wrap(s.err, "sumologic: incorrect configuration")
	}
	if s.serializer == nil {
		return errors.New("sumologic: serializer unset")
	}

	reqBody, err := s.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	if l := len(reqBody); l > int(s.MaxRequstBodySize) {
		var (
			// Do the rounded up integer division
			numChunks  = (l + int(s.MaxRequstBodySize) - 1) / int(s.MaxRequstBodySize)
			chunks     = make([][]byte, 0, numChunks)
			numMetrics = len(metrics)
			// Do the rounded up integer division
			stepMetrics = (numMetrics + numChunks - 1) / numChunks
		)

		for i := 0; i < numMetrics; i += stepMetrics {
			boundary := i + stepMetrics
			if boundary > numMetrics {
				boundary = numMetrics - 1
			}

			chunkBody, err := s.serializer.SerializeBatch(metrics[i:boundary])
			if err != nil {
				return err
			}
			chunks = append(chunks, chunkBody)
		}

		return s.writeRequestChunks(chunks)
	}

	return s.write(reqBody)
}

func (s *SumoLogic) writeRequestChunks(chunks [][]byte) error {
	for _, reqChunk := range chunks {
		if err := s.write(reqChunk); err != nil {
			return err
		}
	}
	return nil
}

func (s *SumoLogic) write(reqBody []byte) error {
	var (
		err  error
		buff bytes.Buffer
		gz   = gzip.NewWriter(&buff)
	)

	if _, err = gz.Write(reqBody); err != nil {
		return err
	}

	if err = gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(s.Method, s.URL, &buff)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("User-Agent", internal.ProductToken())

	// Set headers coming from the configuration.
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	setHeaderIfSetInConfig(req, sourceNameHeader, s.SourceName)
	setHeaderIfSetInConfig(req, sourceHostHeader, s.SourceHost)
	setHeaderIfSetInConfig(req, sourceCategoryHeader, s.SourceCategory)
	setHeaderIfSetInConfig(req, dimensionsHeader, s.Dimensions)

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "sumologic: failed sending request to [%s]", s.URL)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf(
			"sumologic: when writing to [%s] received status code: %d",
			s.URL, resp.StatusCode,
		)
	}

	return nil
}

func setHeaderIfSetInConfig(r *http.Request, h header, value string) {
	if value != "" {
		r.Header.Set(string(h), value)
	}
}

func Default() *SumoLogic {
	return &SumoLogic{
		Timeout: internal.Duration{
			Duration: defaultClientTimeout,
		},
		Method:            defaultMethod,
		MaxRequstBodySize: defaultMaxRequestBodySize,
		headers:           make(map[string]string),
	}
}

func init() {
	outputs.Add("sumologic", func() telegraf.Output {
		return Default()
	})
}
