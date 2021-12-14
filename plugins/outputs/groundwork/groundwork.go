package groundwork

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gwos/tcg/sdk/clients"
	"github.com/gwos/tcg/sdk/logper"
	"github.com/gwos/tcg/sdk/transit"
	"github.com/hashicorp/go-uuid"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const sampleConfig = `
  ## URL of your groundwork instance.
  url = "https://groundwork.example.com"

  ## Agent uuid for GroundWork API Server.
  agent_id = ""

  ## Username and password to access GroundWork API.
  username = ""
  password = ""

  ## Default display name for the host with services(metrics).
  # default_host = "telegraf"

  ## Default service state.
  # default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname.
  # resource_tag = "host"
`

type Groundwork struct {
	Server              string          `toml:"url"`
	AgentID             string          `toml:"agent_id"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	DefaultHost         string          `toml:"default_host"`
	DefaultServiceState string          `toml:"default_service_state"`
	ResourceTag         string          `toml:"resource_tag"`
	Log                 telegraf.Logger `toml:"-"`
	client              clients.GWClient
}

func (g *Groundwork) SampleConfig() string {
	return sampleConfig
}

func (g *Groundwork) Init() error {
	if g.Server == "" {
		return errors.New("no 'url' provided")
	}
	if g.AgentID == "" {
		return errors.New("no 'agent_id' provided")
	}
	if g.Username == "" {
		return errors.New("no 'username' provided")
	}
	if g.Password == "" {
		return errors.New("no 'password' provided")
	}
	if g.DefaultHost == "" {
		return errors.New("no 'default_host' provided")
	}
	if g.ResourceTag == "" {
		return errors.New("no 'resource_tag' provided")
	}
	if !validStatus(g.DefaultServiceState) {
		return errors.New("invalid 'default_service_state' provided")
	}

	g.client = clients.GWClient{
		AppName: "telegraf",
		AppType: "TELEGRAF",
		GWConnection: &clients.GWConnection{
			HostName:           g.Server,
			UserName:           g.Username,
			Password:           g.Password,
			IsDynamicInventory: true,
		},
	}

	logper.SetLogger(
		func(fields interface{}, format string, a ...interface{}) {
			g.Log.Error(adaptLog(fields, format, a...))
		},
		func(fields interface{}, format string, a ...interface{}) {
			g.Log.Warn(adaptLog(fields, format, a...))
		},
		func(fields interface{}, format string, a ...interface{}) {
			g.Log.Info(adaptLog(fields, format, a...))
		},
		func(fields interface{}, format string, a ...interface{}) {
			g.Log.Debug(adaptLog(fields, format, a...))
		},
		func() bool { return telegraf.Debug },
	)
	return nil
}

func (g *Groundwork) Connect() error {
	err := g.client.Connect()
	if err != nil {
		return fmt.Errorf("could not log in: %v", err)
	}
	return nil
}

func (g *Groundwork) Close() error {
	err := g.client.Disconnect()
	if err != nil {
		return fmt.Errorf("could not log out: %v", err)
	}
	return nil
}

func (g *Groundwork) Write(metrics []telegraf.Metric) error {
	resourceToServicesMap := make(map[string][]transit.DynamicMonitoredService)
	for _, metric := range metrics {
		resource, service, err := g.parseMetric(metric)
		if err != nil {
			g.Log.Errorf("%v", err)
			continue
		}
		resourceToServicesMap[resource] = append(resourceToServicesMap[resource], *service)
	}

	var resources []transit.DynamicMonitoredResource
	for resourceName, services := range resourceToServicesMap {
		resources = append(resources, transit.DynamicMonitoredResource{
			BaseResource: transit.BaseResource{
				BaseTransitData: transit.BaseTransitData{
					Name: resourceName,
					Type: transit.Host,
				},
			},
			Status:        transit.HostUp,
			LastCheckTime: transit.NewTimestamp(),
			NextCheckTime: transit.NewTimestamp(), // Temporary work around to avoid error from the server.
			Services:      services,
		})
	}

	traceToken, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	requestJSON, err := json.Marshal(transit.DynamicResourcesWithServicesRequest{
		Context: &transit.TracerContext{
			AppType:    "TELEGRAF",
			AgentID:    g.AgentID,
			TraceToken: traceToken,
			TimeStamp:  transit.NewTimestamp(),
			Version:    transit.ModelVersion,
		},
		Resources: resources,
		Groups:    nil,
	})

	if err != nil {
		return err
	}

	_, err = g.client.SendResourcesWithMetrics(context.Background(), requestJSON)
	if err != nil {
		return fmt.Errorf("error while sending: %v", err)
	}

	return nil
}

func (g *Groundwork) Description() string {
	return "Send telegraf metrics to GroundWork Monitor"
}

func init() {
	outputs.Add("groundwork", func() telegraf.Output {
		return &Groundwork{
			ResourceTag:         "host",
			DefaultHost:         "telegraf",
			DefaultServiceState: string(transit.ServiceOk),
		}
	})
}

func (g *Groundwork) parseMetric(metric telegraf.Metric) (string, *transit.DynamicMonitoredService, error) {
	resource := g.DefaultHost
	if value, present := metric.GetTag(g.ResourceTag); present {
		resource = value
	}

	service := metric.Name()
	if value, present := metric.GetTag("service"); present {
		service = value
	}

	status := g.DefaultServiceState
	value, statusPresent := metric.GetTag("status")
	if validStatus(value) {
		status = value
	}

	message, _ := metric.GetTag("message")

	unitType := string(transit.UnitCounter)
	if value, present := metric.GetTag("unitType"); present {
		unitType = value
	}

	var critical float64
	value, criticalPresent := metric.GetTag("critical")
	if criticalPresent {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			critical = s
		}
	}

	var warning float64
	value, warningPresent := metric.GetTag("warning")
	if warningPresent {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			warning = s
		}
	}

	lastCheckTime := transit.NewTimestamp()
	lastCheckTime.Time = metric.Time()
	serviceObject := transit.DynamicMonitoredService{
		BaseTransitData: transit.BaseTransitData{
			Name:  service,
			Type:  transit.Service,
			Owner: resource,
		},
		Status:           transit.MonitorStatus(status),
		LastCheckTime:    lastCheckTime,
		NextCheckTime:    lastCheckTime, // Temporary work around to avoid error from the server.
		LastPlugInOutput: message,
		Metrics:          nil,
	}

	for _, value := range metric.FieldList() {
		var thresholds []transit.ThresholdValue
		if warningPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Warning,
				Label:      value.Key + "_wn",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: warning,
				},
			})
		}
		if criticalPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Critical,
				Label:      value.Key + "_cr",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: critical,
				},
			})
		}

		typedValue := new(transit.TypedValue)
		err := typedValue.FromInterface(value.Value)
		if err != nil {
			return "", nil, err
		}
		if typedValue.ValueType == transit.StringType {
			g.Log.Warn("string values are not supported, skipping")
			continue
		}

		serviceObject.Metrics = append(serviceObject.Metrics, transit.TimeSeries{
			MetricName: value.Key,
			SampleType: transit.Value,
			Interval: &transit.TimeInterval{
				EndTime:   lastCheckTime,
				StartTime: lastCheckTime, // Temporary work around to avoid error from the server.
			},
			Value:      typedValue,
			Unit:       transit.UnitType(unitType),
			Thresholds: &thresholds,
		})
	}

	if !statusPresent {
		serviceStatus, err := transit.CalculateServiceStatus(&serviceObject.Metrics)
		if err != nil {
			g.Log.Infof("could not calculate service status, reverting to default_service_state: %v", err)
			serviceObject.Status = transit.MonitorStatus(g.DefaultServiceState)
		}
		serviceObject.Status = serviceStatus
	}

	return resource, &serviceObject, nil
}

func validStatus(status string) bool {
	switch transit.MonitorStatus(status) {
	case transit.ServiceOk, transit.ServiceWarning, transit.ServicePending, transit.ServiceScheduledCritical,
		transit.ServiceUnscheduledCritical, transit.ServiceUnknown:
		return true
	}
	return false
}

func adaptLog(fields interface{}, format string, a ...interface{}) string {
	buf := &bytes.Buffer{}
	if format != "" {
		_, _ = fmt.Fprintf(buf, format, a...)
	}
	fmtField := func(k string, v interface{}) {
		format := " %s:"
		if len(k) == 0 {
			format = " "
		}
		if _, ok := v.(int); ok {
			format += "%d"
		} else {
			format += "%q"
		}
		_, _ = fmt.Fprintf(buf, format, k, v)
	}
	if ff, ok := fields.(interface {
		LogFields() (map[string]interface{}, map[string][]byte)
	}); ok {
		m1, m2 := ff.LogFields()
		for k, v := range m1 {
			fmtField(k, v)
		}
		for k, v := range m2 {
			fmtField(k, v)
		}
	} else if ff, ok := fields.(map[string]interface{}); ok {
		for k, v := range ff {
			fmtField(k, v)
		}
	} else if ff, ok := fields.([]interface{}); ok {
		for _, v := range ff {
			fmtField("", v)
		}
	}
	out := buf.Bytes()
	if len(out) > 1 {
		out = append(bytes.ToUpper(out[0:1]), out[1:]...)
	}
	return string(out)
}
