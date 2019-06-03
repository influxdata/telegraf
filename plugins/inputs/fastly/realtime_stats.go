package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/influxdata/telegraf"
	"log"
	"reflect"
	"strings"
	"time"
)

func (f *Fastly) collectRealtimeStats(acc telegraf.Accumulator) error {
	for _, service := range f.services {
		go f.collectRealtimeStatsForService(acc, service)
	}
	return nil
}

func (f *Fastly) collectRealtimeStatsForService(acc telegraf.Accumulator, service *fastly.Service) {
	input := fastly.GetRealtimeStatsInput{
		Service:   service.ID,
		Timestamp: f.rtUpdateTracker.LastUpdate(service.ID),
	}
	resp, err := f.rtClient.GetRealtimeStats(&input)
	respTime := time.Now()
	if err != nil {
		log.Println("E! [inputs.fastly] Error while fetching realtime stats:", err)
		return
	}
	if resp.Error != "" && !strings.HasPrefix(resp.Error, "No data available") {
		log.Println("E! [inputs.fastly] Error in realtime stats response:", resp.Error)
		return
	}
	f.collectRealtimeStatsFromAPIResponse(acc, service, resp, respTime)
	// TODO: Collect Service stats (Versions, ActiveVersion, UpdatedAt)
	f.rtUpdateTracker.TrackUpdate(service.ID, resp.Timestamp)
}

func (f *Fastly) collectRealtimeStatsFromAPIResponse(acc telegraf.Accumulator, service *fastly.Service, resp *fastly.RealtimeStatsResponse, respTime time.Time) {
	for _, data := range resp.Data {
		for dcName, dcStats := range data.Datacenter {
			f.collectRealtimeStatsFromDataCenter(acc, service, dcName, dcStats, respTime)
		}
	}
}

func (f *Fastly) collectRealtimeStatsFromDataCenter(acc telegraf.Accumulator, service *fastly.Service, dcName string, dcStats *fastly.Stats, respTime time.Time) {
	tags := map[string]string{
		"service_name": service.Name,
		"service_id":   service.ID,
		"datacenter":   dcName,
	}
	// TODO: Skip: WAF, ObjectSize, ancient TLS versions
	dcStatElements := reflect.ValueOf(dcStats).Elem()
	fields := make(map[string]interface{}, dcStatElements.NumField())
	for i := 0; i < dcStatElements.NumField(); i++ {
		typeField := dcStatElements.Type().Field(i)
		//varName := typeField.Name
		//varType := typeField.Type
		accFieldName := typeField.Tag.Get("mapstructure")
		varValue := dcStatElements.Field(i).Interface()
		fields[accFieldName] = varValue
	}
	acc.AddFields(statPrefix+"service.stats", fields, tags, respTime)
}

type realtimeUpdateTracker struct {
	services map[string]realtimeUpdateTrackerRecord
}

func (r *realtimeUpdateTracker) LastUpdate(serviceId string) uint64 {
	if record, ok := r.services[serviceId]; ok {
		return record.lastUpdate
	}
	return 0
}

func (r *realtimeUpdateTracker) TrackUpdate(serviceId string, ts uint64) {
	if record, ok := r.services[serviceId]; ok {
		record.lastUpdate = ts
	} else {
		r.services[serviceId] = realtimeUpdateTrackerRecord{lastUpdate: ts}
	}
}

type realtimeUpdateTrackerRecord struct {
	lastUpdate uint64
}

func newRealtimeUpdateTracker() *realtimeUpdateTracker {
	return &realtimeUpdateTracker{
		services: make(map[string]realtimeUpdateTrackerRecord),
	}
}
