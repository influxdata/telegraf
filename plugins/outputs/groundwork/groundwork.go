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

type metricMeta struct {
	group    string
	resource string
}

type Groundwork struct {
	Server              string          `toml:"url"`
	AgentID             string          `toml:"agent_id"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	DefaultHost         string          `toml:"default_host"`
	DefaultServiceState string          `toml:"default_service_state"`
	GroupTag            string          `toml:"group_tag"`
	ResourceTag         string          `toml:"resource_tag"`
	Log                 telegraf.Logger `toml:"-"`
	client              clients.GWClient
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
	groupMap := make(map[string][]transit.ResourceRef)
	resourceToServicesMap := make(map[string][]transit.MonitoredService)
	for _, metric := range metrics {
		meta, service, err := g.parseMetric(metric)
		if err != nil {
			g.Log.Errorf("%v", err)
			continue
		}
		resource := meta.resource
		resourceToServicesMap[resource] = append(resourceToServicesMap[resource], *service)

		group := meta.group
		if len(group) != 0 {
			resRef := transit.ResourceRef{
				Name: resource,
				Type: transit.ResourceTypeHost,
			}
			if refs, ok := groupMap[group]; ok {
				refs = append(refs, resRef)
				groupMap[group] = refs
			} else {
				groupMap[group] = []transit.ResourceRef{resRef}
			}
		}
	}

	groups := make([]transit.ResourceGroup, 0, len(groupMap))
	for groupName, refs := range groupMap {
		groups = append(groups, transit.ResourceGroup{
			GroupName: groupName,
			Resources: refs,
			Type:      transit.HostGroup,
		})
	}

	var resources []transit.MonitoredResource
	for resourceName, services := range resourceToServicesMap {
		resources = append(resources, transit.MonitoredResource{
			BaseResource: transit.BaseResource{
				BaseInfo: transit.BaseInfo{
					Name: resourceName,
					Type: transit.ResourceTypeHost,
				},
			},
			MonitoredInfo: transit.MonitoredInfo{
				Status:        transit.HostUp,
				LastCheckTime: transit.NewTimestamp(),
			},
			Services: services,
		})
	}

	traceToken, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	requestJSON, err := json.Marshal(transit.ResourcesWithServicesRequest{
		Context: &transit.TracerContext{
			AppType:    "TELEGRAF",
			AgentID:    g.AgentID,
			TraceToken: traceToken,
			TimeStamp:  transit.NewTimestamp(),
			Version:    transit.ModelVersion,
		},
		Resources: resources,
		Groups:    groups,
	})

	if err != nil {
		return err
	}

	_, err = g.client.SendResourcesWithMetrics(context.Background(), requestJSON)
	if err != nil {
		return fmt.Errorf("error while sending: %w", err)
	}

	return nil
}

func init() {
	outputs.Add("groundwork", func() telegraf.Output {
		return &Groundwork{
			GroupTag:            "group",
			ResourceTag:         "host",
			DefaultHost:         "telegraf",
			DefaultServiceState: string(transit.ServiceOk),
		}
	})
}

func (g *Groundwork) parseMetric(metric telegraf.Metric) (metricMeta, *transit.MonitoredService, error) {
	group, _ := metric.GetTag(g.GroupTag)

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
	serviceObject := transit.MonitoredService{
		BaseInfo: transit.BaseInfo{
			Name:  service,
			Type:  transit.ResourceTypeService,
			Owner: resource,
		},
		MonitoredInfo: transit.MonitoredInfo{
			Status:           transit.MonitorStatus(status),
			LastCheckTime:    lastCheckTime,
			NextCheckTime:    lastCheckTime, // if not added, GW will make this as LastCheckTime + 5 mins
			LastPluginOutput: message,
		},
		Metrics: nil,
	}

	for _, value := range metric.FieldList() {
		var thresholds []transit.ThresholdValue
		if warningPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Warning,
				Label:      value.Key + "_wn",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: &warning,
				},
			})
		}
		if criticalPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Critical,
				Label:      value.Key + "_cr",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: &critical,
				},
			})
		}

		typedValue := transit.NewTypedValue(value.Value)
		if typedValue == nil {
			g.Log.Warnf("could not convert type %T, skipping field %s: %v", value.Value, value.Key, value.Value)
			continue
		}
		if typedValue.ValueType == transit.StringType {
			g.Log.Warnf("string values are not supported, skipping field %s: %q", value.Key, value.Value)
			continue
		}

		serviceObject.Metrics = append(serviceObject.Metrics, transit.TimeSeries{
			MetricName: value.Key,
			SampleType: transit.Value,
			Interval: &transit.TimeInterval{
				EndTime: lastCheckTime,
			},
			Value:      typedValue,
			Unit:       transit.UnitType(unitType),
			Thresholds: thresholds,
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

	return metricMeta{resource: resource, group: group}, &serviceObject, nil
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
