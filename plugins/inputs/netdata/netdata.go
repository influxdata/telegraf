package netdata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const chartsURL = "/api/v1/charts"

const description = `Read metrics from netdata`

const sampleConfig = `
  ## Number of points to gather.
  ## By default inputs plugin run every 10 seconds, by asking 10 points
  ## to netdata we get measurement for each second. This value is capped by "update every" config of netdata.
  ## For example you can set
  ## points = 20
  ## interval = "10m"
  ## to get measurement every 30 seconds (and polling netdata every 10 minutes)
  ## default to 10
  points = 10

  ## Grouping method if multiple collected values are to be grouped in order to return fewer points.
  ## methods supported "min", "max", "average", "sum", "incremental-sum"
  ## cf : https://github.com/firehol/netdata/wiki/REST-API-v1
  ## default to average
  group = "average"

  ## Should measurement be prefixed by the netdata hostname ?
  ## Ex : netdata.host1.apps.cpu (with hostname),  netdata.apps.cpu (without hostname)
  ## default to true
  prefixChartIDByHostname = true

  ## Netdata servers
  [[inputs.netdata.servers]]
    url="http://netdata-host:19999"

  [[inputs.netdata.servers]]
    url="https://netdata-host2"
    basicAuthUsername="user"
    basicAuthPassword="pass"
`

type Netdata struct {
	Servers                 []Server
	Points                  int
	Group                   string
	PrefixChartIDByHostname bool

	client *http.Client
}

type Server struct {
	Url               string
	BasicAuthUsername string
	BasicAuthPassword string
	lastGatherTime    int64
	currGatherTime    int64
}

type NestedError struct {
	Err       error
	NestedErr error
}

func (ne NestedError) Error() string {
	return ne.Err.Error() + ": " + ne.NestedErr.Error()
}

func Errorf(err error, msg string, format ...interface{}) error {
	return NestedError{
		NestedErr: err,
		Err:       fmt.Errorf(msg, format...),
	}
}

type ResponseCharts struct {
	Hostname string           `json:"hostname"`
	Charts   map[string]Chart `json:"charts"`
}

type Chart struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Family     string                    `json:"family"`
	DataURL    string                    `json:"data_url"`
	Dimensions map[string]ChartDimension `json:"dimensions"`
}

type ChartDimension struct {
	Name string `json:"name"`
}

type ResponseData struct {
	DimensionNames []string   `json:"dimension_names"`
	Result         DataResult `json:"result"`
}

type DataResult struct {
	Data [][]interface{} `json:"data"`
}

func init() {
	inputs.Add("netdata", func() telegraf.Input {
		return &Netdata{
			Points: 10,
			Group:  "average",
			PrefixChartIDByHostname: true,
		}
	})
}

func (n *Netdata) Description() string {
	return description
}

func (n *Netdata) SampleConfig() string {
	return sampleConfig
}

func (n *Netdata) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if n.client == nil {
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		}
		n.client = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
	}

	for idx := range n.Servers {
		wg.Add(1)
		go func(server *Server) {
			defer wg.Done()
			n.gatherServer(acc, server)
		}(&n.Servers[idx])
	}

	wg.Wait()
	return nil
}

func (n *Netdata) gatherServer(acc telegraf.Accumulator, server *Server) {
	server.currGatherTime = time.Now().Unix()

	// first time gather data from the past 10s
	if server.lastGatherTime == 0 {
		server.lastGatherTime = time.Now().Unix() - 10
	}

	chartRequestURL, err := url.Parse(server.Url + chartsURL)
	if err != nil {
		acc.AddError(Errorf(err, "error parsing charts url : %s", server.Url+chartsURL))
		return
	}

	chartRespBody, err := n.sendRequest(chartRequestURL, server)
	if err != nil {
		acc.AddError(Errorf(err, "error getting netdata charts, url : %s", chartRequestURL.String()))
		return
	}

	var chartResp ResponseCharts
	if err := json.Unmarshal([]byte(chartRespBody), &chartResp); err != nil {
		acc.AddError(Errorf(err, "error unmarshalling netdata charts"))
		return
	}

	for _, chart := range chartResp.Charts {
		dataStringURL := n.buildChartDataURL(server, chart)

		dataRequestURL, err := url.Parse(dataStringURL)

		if err != nil {
			acc.AddError(Errorf(err, "error parsing chart data url : %s", dataStringURL))
			continue
		}

		dataRespBody, err := n.sendRequest(dataRequestURL, server)
		if err != nil {
			acc.AddError(Errorf(err, "error getting netdata chart data, url : %s", dataRequestURL.String()))
			continue
		}

		var dataResp ResponseData
		if err := json.Unmarshal([]byte(dataRespBody), &dataResp); err != nil {
			acc.AddError(Errorf(err, "error unmarshalling netdata chart data"))
			continue
		}

		dimensions := strings.Join(dataResp.DimensionNames, ",")

		var measurement string
		if n.PrefixChartIDByHostname {
			measurement = "netdata." + chartResp.Hostname + "." + chart.ID
		} else {
			measurement = "netdata." + chart.ID
		}

		for _, chartData := range dataResp.Result.Data {
			fields := make(map[string]interface{})
			time := time.Unix(int64(chartData[0].(float64)), 0)
			values := chartData[1:]
			for i, dimension := range dataResp.DimensionNames {
				if values[i] != nil {
					fields[dimension] = values[i]
				}
			}

			tags := map[string]string{
				"server":     chartRequestURL.Host,
				"hostname":   chartResp.Hostname,
				"type":       chart.Type,
				"family":     chart.Family,
				"dimensions": dimensions,
			}

			acc.AddFields(measurement, fields, tags, time)
		}
	}
	server.lastGatherTime = server.currGatherTime
}

func (n *Netdata) buildChartDataURL(s *Server, c Chart) string {
	return s.Url +
		c.DataURL +
		"&group=" + n.Group +
		"&options=absolute|jsonwrap|seconds&after=" +
		strconv.FormatInt(s.lastGatherTime, 10) +
		"&before=" +
		strconv.FormatInt(s.currGatherTime, 10) +
		"&points=" +
		strconv.Itoa(n.Points)
}

func (n *Netdata) sendRequest(url *url.URL, server *Server) (string, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	method := "GET"
	content := bytes.NewBufferString("")
	if server.BasicAuthPassword != "" && server.BasicAuthUsername != "" {
		headers["Authorization"] = "Basic " + base64.URLEncoding.EncodeToString([]byte(server.BasicAuthUsername+":"+server.BasicAuthPassword))
	}
	req, err := http.NewRequest(method, url.String(), content)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			url.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(body), err
	}
	return string(body), err
}
