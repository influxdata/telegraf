package dcos

import (
	"context"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultMaxConnections  = 10
	defaultResponseTimeout = 20 * time.Second
)

var (
	nodeDimensions = []string{
		"hostname",
		"path",
		"interface",
	}
	containerDimensions = []string{
		"hostname",
		"container_id",
		"task_name",
	}
	appDimensions = []string{
		"hostname",
		"container_id",
		"task_name",
	}
)

type DCOS struct {
	ClusterURL string `toml:"cluster_url"`

	ServiceAccountID         string `toml:"service_account_id"`
	ServiceAccountPrivateKey string

	TokenFile string

	NodeInclude      []string
	NodeExclude      []string
	ContainerInclude []string
	ContainerExclude []string
	AppInclude       []string
	AppExclude       []string

	MaxConnections  int
	ResponseTimeout config.Duration
	tls.ClientConfig

	client Client
	creds  Credentials

	initialized     bool
	nodeFilter      filter.Filter
	containerFilter filter.Filter
	appFilter       filter.Filter
}

func (d *DCOS) Gather(acc telegraf.Accumulator) error {
	err := d.init()
	if err != nil {
		return err
	}

	ctx := context.Background()

	token, err := d.creds.Token(ctx, d.client)
	if err != nil {
		return err
	}
	d.client.SetToken(token)

	summary, err := d.client.GetSummary(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, node := range summary.Slaves {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			d.GatherNode(ctx, acc, summary.Cluster, node)
		}(node.ID)
	}
	wg.Wait()

	return nil
}

func (d *DCOS) GatherNode(ctx context.Context, acc telegraf.Accumulator, cluster, node string) {
	if !d.nodeFilter.Match(node) {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m, err := d.client.GetNodeMetrics(ctx, node)
		if err != nil {
			acc.AddError(err)
			return
		}
		d.addNodeMetrics(acc, cluster, m)
	}()

	d.GatherContainers(ctx, acc, cluster, node)
	wg.Wait()
}

func (d *DCOS) GatherContainers(ctx context.Context, acc telegraf.Accumulator, cluster, node string) {
	containers, err := d.client.GetContainers(ctx, node)
	if err != nil {
		acc.AddError(err)
		return
	}

	var wg sync.WaitGroup
	for _, container := range containers {
		if d.containerFilter.Match(container.ID) {
			wg.Add(1)
			go func(container string) {
				defer wg.Done()
				m, err := d.client.GetContainerMetrics(ctx, node, container)
				if err != nil {
					if err, ok := err.(APIError); ok && err.StatusCode == 404 {
						return
					}
					acc.AddError(err)
					return
				}
				d.addContainerMetrics(acc, cluster, m)
			}(container.ID)
		}

		if d.appFilter.Match(container.ID) {
			wg.Add(1)
			go func(container string) {
				defer wg.Done()
				m, err := d.client.GetAppMetrics(ctx, node, container)
				if err != nil {
					if err, ok := err.(APIError); ok && err.StatusCode == 404 {
						return
					}
					acc.AddError(err)
					return
				}
				d.addAppMetrics(acc, cluster, m)
			}(container.ID)
		}
	}
	wg.Wait()
}

type point struct {
	tags   map[string]string
	labels map[string]string
	fields map[string]interface{}
}

func (d *DCOS) createPoints(m *Metrics) []*point {
	points := make(map[string]*point)
	for _, dp := range m.Datapoints {
		fieldKey := strings.Replace(dp.Name, ".", "_", -1)

		tags := dp.Tags
		if tags == nil {
			tags = make(map[string]string)
		}

		if dp.Unit == "bytes" && !strings.HasSuffix(fieldKey, "_bytes") {
			fieldKey = fieldKey + "_bytes"
		}

		fieldKey = strings.TrimPrefix(fieldKey, "dcos_metrics_module_")

		tagset := make([]string, 0, len(tags))
		for k, v := range tags {
			tagset = append(tagset, k+"="+v)
		}
		sort.Strings(tagset)
		seriesParts := make([]string, 0, len(tagset))
		seriesParts = append(seriesParts, tagset...)
		seriesKey := strings.Join(seriesParts, ",")

		p, ok := points[seriesKey]
		if !ok {
			p = &point{}
			p.tags = tags
			p.labels = make(map[string]string)
			p.fields = make(map[string]interface{})
			points[seriesKey] = p
		}

		if dp.Unit == "bytes" {
			p.fields[fieldKey] = int64(dp.Value)
		} else {
			p.fields[fieldKey] = dp.Value
		}
	}

	results := make([]*point, 0, len(points))
	for _, p := range points {
		for k, v := range m.Dimensions {
			switch v := v.(type) {
			case string:
				p.tags[k] = v
			case map[string]string:
				if k == "labels" {
					for k, v := range v {
						p.labels[k] = v
					}
				}
			}
		}
		results = append(results, p)
	}
	return results
}

func (d *DCOS) addMetrics(acc telegraf.Accumulator, cluster, mname string, m *Metrics, tagDimensions []string) {
	tm := time.Now()

	points := d.createPoints(m)

	for _, p := range points {
		tags := make(map[string]string)
		tags["cluster"] = cluster
		for _, tagkey := range tagDimensions {
			v, ok := p.tags[tagkey]
			if ok {
				tags[tagkey] = v
			}
		}
		for k, v := range p.labels {
			tags[k] = v
		}

		acc.AddFields(mname, p.fields, tags, tm)
	}
}

func (d *DCOS) addNodeMetrics(acc telegraf.Accumulator, cluster string, m *Metrics) {
	d.addMetrics(acc, cluster, "dcos_node", m, nodeDimensions)
}

func (d *DCOS) addContainerMetrics(acc telegraf.Accumulator, cluster string, m *Metrics) {
	d.addMetrics(acc, cluster, "dcos_container", m, containerDimensions)
}

func (d *DCOS) addAppMetrics(acc telegraf.Accumulator, cluster string, m *Metrics) {
	d.addMetrics(acc, cluster, "dcos_app", m, appDimensions)
}

func (d *DCOS) init() error {
	if !d.initialized {
		err := d.createFilters()
		if err != nil {
			return err
		}

		if d.client == nil {
			client, err := d.createClient()
			if err != nil {
				return err
			}
			d.client = client
		}

		if d.creds == nil {
			creds, err := d.createCredentials()
			if err != nil {
				return err
			}
			d.creds = creds
		}

		d.initialized = true
	}
	return nil
}

func (d *DCOS) createClient() (Client, error) {
	tlsCfg, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	address, err := url.Parse(d.ClusterURL)
	if err != nil {
		return nil, err
	}

	client := NewClusterClient(
		address,
		time.Duration(d.ResponseTimeout),
		d.MaxConnections,
		tlsCfg,
	)

	return client, nil
}

func (d *DCOS) createCredentials() (Credentials, error) {
	if d.ServiceAccountID != "" && d.ServiceAccountPrivateKey != "" {
		bs, err := os.ReadFile(d.ServiceAccountPrivateKey)
		if err != nil {
			return nil, err
		}

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bs)
		if err != nil {
			return nil, err
		}

		creds := &ServiceAccount{
			AccountID:  d.ServiceAccountID,
			PrivateKey: privateKey,
		}
		return creds, nil
	} else if d.TokenFile != "" {
		creds := &TokenCreds{
			Path: d.TokenFile,
		}
		return creds, nil
	} else {
		creds := &NullCreds{}
		return creds, nil
	}
}

func (d *DCOS) createFilters() error {
	var err error
	d.nodeFilter, err = filter.NewIncludeExcludeFilter(
		d.NodeInclude, d.NodeExclude)
	if err != nil {
		return err
	}

	d.containerFilter, err = filter.NewIncludeExcludeFilter(
		d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}

	d.appFilter, err = filter.NewIncludeExcludeFilter(
		d.AppInclude, d.AppExclude)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("dcos", func() telegraf.Input {
		return &DCOS{
			MaxConnections:  defaultMaxConnections,
			ResponseTimeout: config.Duration(defaultResponseTimeout),
		}
	})
}
