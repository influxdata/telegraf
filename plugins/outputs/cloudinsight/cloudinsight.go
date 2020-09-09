package cloudinsight

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/samjegal/fincloud-sdk-for-go/services/cloudinsight"
	"github.com/samjegal/go-fincloud-helpers/sender"
)

// CloudInsight main structure
type CloudInsight struct {
	// Region data center name
	Region string `toml:"region"`

	// NBP Cloud Insight Custom Schema Credentials (filter)
	ProductName string `toml:"product_name"`
	ProductKey  string `toml:"cw_key"`

	// NBP Credentials
	AccessKey     string `toml:"access_key"`
	SecretKey     string `toml:"secret_key"`
	ApiGatewayKey string `toml:"api_gateway_key"`

	// TODO: add cloudinsight client
	client *cloudinsight.BaseClient

	// TODO: deprecated because replaced tag value
	// Make dimension by custom schema. InstanceId used by server instance.
	//InstanceId  string `toml:"instance_id"`

	filter string

	timeFunc func() time.Time
}

var sampleConfig = `
  ## Financial Cloud Region (ncloud.com/fin-ncloud.com)
  region = "FKR"

  ##
  access_key = ""
  secret_key = ""

  ##
  api_gateway_key = ""
`

func (c *CloudInsight) Description() string {
	return "Configuration for NBP Cloud Insight output plugin."
}

func (c *CloudInsight) SampleConfig() string {
	return sampleConfig
}

func (c *CloudInsight) Init() error {
	r, _ := regexp.Compile("[a-zA-Z]+/[a-zA-Z]+")
	if r.MatchString(c.ProductName) {
		f := strings.Split(c.ProductName, "/")
		c.filter = f[1]
	} else {
		return fmt.Errorf("could not read cloudinsight product name for custom metric")
	}

	c.client = &cloudinsight.BaseClient{
		Client:  autorest.NewClientWithUserAgent(cloudinsight.UserAgent()),
		BaseURI: cloudinsight.DefaultBaseURI,
	}
	c.client.Sender = sender.BuildSender("fincloud")
	c.authorize(c.client)

	return nil
}

// authorize set credential for function using api gateway key
func (c *CloudInsight) authorize(client *cloudinsight.BaseClient) {
	client.AccessKey = c.AccessKey
	client.Secretkey = c.SecretKey
	client.APIGatewayAPIKey = c.ApiGatewayKey
}

func (c *CloudInsight) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(float64(60)*float64(time.Second)))
	defer cancel()

	// check get custom metric from cloudinsight managed service
	resp, err := cloudinsight.SchemaClient{BaseClient: *c.client}.Get(ctx, c.ProductName, c.ProductKey)
	if err != nil {
		return fmt.Errorf("could not read cloudinsight product schema")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not read cloudinsight product schema for %q", c.ProductName)
	}

	return nil
}

func (c *CloudInsight) Close() error {
	c.client = nil
	return nil
}

type cloudInsightMetric struct {
	Time       time.Time
	Dimensions map[string]string
	Name       string
	Data       []*cloudInsightData
}

type cloudInsightData struct {
	Metric string
	Value  interface{}
}

func (c *CloudInsight) Write(metrics []telegraf.Metric) error {
	var cimetrics = make(map[uint64]*cloudInsightMetric, len(metrics))
	for _, m := range metrics {
		if m.Name() != c.filter {
			continue
		}

		id := hashTagKey(m)
		if cim, ok := cimetrics[id]; !ok {
			cmm, err := convert(m)
			if err != nil {
				log.Printf("E! [outputs.cloudinsight]: could not create cloudinsight metric for %q", m.Name())
				continue
			}
			cimetrics[id] = cmm
		} else {
			cmm, err := convert(m)
			if err != nil {
				log.Printf("E! [outputs.cloudinsight]: could not create cloudinsight metric for %q", m.Name())
				continue
			}
			cimetrics[id].Data = append(cim.Data, cmm.Data...)
		}
	}

	if len(cimetrics) == 0 {
		return nil
	}

	var data []interface{} // -> map[string]interface{}{}
	for _, m := range cimetrics {
		e := map[string]interface{}{}
		for n, v := range m.Dimensions {
			e[n] = v
		}

		for _, d := range m.Data {
			e[d.Metric] = d.Value
		}

		data = append(data, e)
	}

	return c.send(cloudinsight.CollectorRequest{
		CwKey: &c.ProductKey,
		Data:  &data,
	})

	return nil
}

func convert(m telegraf.Metric) (*cloudInsightMetric, error) {
	dimensions := make(map[string]string, len(m.TagList()))
	for _, tag := range m.TagList() {
		if _, ok := dimensions[tag.Key]; !ok {
			dimensions[tag.Key] = tag.Value
		}
	}

	var data []*cloudInsightData
	for k, v := range m.Fields() {
		data = append(data, &cloudInsightData{
			Metric: k,
			Value:  v,
		})
	}

	return &cloudInsightMetric{
		Time:       m.Time(),
		Dimensions: dimensions,
		Name:       m.Name(),
		Data:       data,
	}, nil
}

func hashTagKey(m telegraf.Metric) uint64 {
	h := fnv.New64a()
	h.Write([]byte(m.Name()))
	for _, tag := range m.TagList() {
		if tag.Key == "" {
			continue
		}
		h.Write([]byte(tag.Key))
		h.Write([]byte(tag.Value))
	}
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, uint64(m.Time().UnixNano()))
	h.Write(b[:n])
	return h.Sum64()
}

func (c *CloudInsight) send(data cloudinsight.CollectorRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(float64(60)*float64(time.Second)))
	defer cancel()

	_, err := cloudinsight.CollectorClient{BaseClient: *c.client}.SendMethod(ctx, data)
	if err != nil {
		return nil
	}

	return nil
}

func init() {
	outputs.Add("cloudinsight", func() telegraf.Output {
		return &CloudInsight{
			timeFunc: time.Now,
		}
	})
}
