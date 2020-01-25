// (c) 2019 Sony Interactive Entertainment Inc.
package jtinative

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	// Proto extensions need to be imported to be registerd
	// the are never called directly.
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/agentd"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/ancpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/authd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_ancp_stats_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_pppoe_stats_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_rsmon_debug_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_rsmon_stats_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_smd_queue_stats_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-smgd_sub_mgmt_network_stats_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/bbe-statsd-telemetry_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/chassisd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/cmerror"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/cmerror_data"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/cpu_memory_utilization"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/dcd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/eventd"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/fabric"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/firewall"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/inline_jflow"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/ipsec_telemetry"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/jdhcpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/jkhmd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/jl2tpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/jpppd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/junos-xmlproxyd_junos-rsvp-interface"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/junos-xmlproxyd_junos-rtg-task-memory"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/kernel-ifstate-render"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/kmd_render"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/l2ald_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/l2ald_oc_intf"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/l2cpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/lacpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/logical_port"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/lsp_stats"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/mib2d_arp_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/mib2d_nd6_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/mib2d_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/npu_memory_utilization"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/npu_utilization"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/optics"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/packet_stats"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pbj"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_ifl_oc"
	"github.com/influxdata/telegraf/plugins/parsers/jtinative/telemetry_top"
	// Conflict Extension registration number. - PR1426871
	//_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_egress_oc"
	//_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_te_ip_oc"
	//_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_te_bsid_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_ingress_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_sid_egress_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_mpls_sr_sid_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_npu_resource"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfe_port_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/pfed_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/port"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/port_exp"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/qmon"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rmopd_render"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_bgp_rib_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_ipv6_ra_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_isis_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_loc_rib_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_ni_bgp_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_rsvp_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/rpd_te_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/session_telemetry"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/smid_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/sr_stats_per_if_egress"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/sr_stats_per_if_ingress"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/sr_stats_per_sid"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/svcset_telemetry"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/vrrpd_oc"
	_ "github.com/influxdata/telegraf/plugins/parsers/jtinative/xmlproxyd_show_local_interface_OC"
)

var protoFiles = [...]string{
	"plugins/parsers/jtinative/telemetry_top/telemetry_top.proto",
	"plugins/parsers/jtinative/agentd/agentd.proto",
	"plugins/parsers/jtinative/ancpd_oc/ancpd_oc.proto",
	"plugins/parsers/jtinative/authd_oc/authd_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_ancp_stats_oc/bbe-smgd_ancp_stats_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_pppoe_stats_oc/bbe-smgd_pppoe_stats_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_rsmon_debug_oc/bbe-smgd_rsmon_debug_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_rsmon_stats_oc/bbe-smgd_rsmon_stats_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_smd_queue_stats_oc/bbe-smgd_smd_queue_stats_oc.proto",
	"plugins/parsers/jtinative/bbe-smgd_sub_mgmt_network_stats_oc/bbe-smgd_sub_mgmt_network_stats_oc.proto",
	"plugins/parsers/jtinative/bbe-statsd-telemetry_oc/bbe-statsd-telemetry_oc.proto",
	"plugins/parsers/jtinative/chassisd_oc/chassisd_oc.proto",
	"plugins/parsers/jtinative/cmerror/cmerror.proto",
	"plugins/parsers/jtinative/cmerror_data/cmerror_data.proto",
	"plugins/parsers/jtinative/cpu_memory_utilization/cpu_memory_utilization.proto",
	"plugins/parsers/jtinative/dcd_oc/dcd_oc.proto",
	"plugins/parsers/jtinative/eventd/eventd.proto",
	"plugins/parsers/jtinative/fabric/fabric.proto",
	"plugins/parsers/jtinative/firewall/firewall.proto",
	"plugins/parsers/jtinative/inline_jflow/inline_jflow.proto",
	"plugins/parsers/jtinative/ipsec_telemetry/ipsec_telemetry.proto",
	"plugins/parsers/jtinative/jdhcpd_oc/jdhcpd_oc.proto",
	"plugins/parsers/jtinative/jkhmd_oc/jkhmd_oc.proto",
	"plugins/parsers/jtinative/jl2tpd_oc/jl2tpd_oc.proto",
	"plugins/parsers/jtinative/jpppd_oc/jpppd_oc.proto",
	"plugins/parsers/jtinative/junos-xmlproxyd_junos-rsvp-interface/junos-xmlproxyd_junos-rsvp-interface.proto",
	"plugins/parsers/jtinative/junos-xmlproxyd_junos-rtg-task-memory/junos-xmlproxyd_junos-rtg-task-memory.proto",
	"plugins/parsers/jtinative/kernel-ifstate-render/kernel-ifstate-render.proto",
	"plugins/parsers/jtinative/kmd_render/kmd_render.proto",
	"plugins/parsers/jtinative/l2ald_oc/l2ald_oc.proto",
	"plugins/parsers/jtinative/l2ald_oc_intf/l2ald_oc_intf.proto",
	"plugins/parsers/jtinative/l2cpd_oc/l2cpd_oc.proto",
	"plugins/parsers/jtinative/lacpd_oc/lacpd_oc.proto",
	"plugins/parsers/jtinative/logical_port/logical_port.proto",
	"plugins/parsers/jtinative/lsp_stats/lsp_stats.proto",
	"plugins/parsers/jtinative/mib2d_arp_oc/mib2d_arp_oc.proto",
	"plugins/parsers/jtinative/mib2d_nd6_oc/mib2d_nd6_oc.proto",
	"plugins/parsers/jtinative/mib2d_oc/mib2d_oc.proto",
	"plugins/parsers/jtinative/npu_memory_utilization/npu_memory_utilization.proto",
	"plugins/parsers/jtinative/npu_utilization/npu_utilization.proto",
	"plugins/parsers/jtinative/optics/optics.proto",
	"plugins/parsers/jtinative/packet_stats/packet_stats.proto",
	"plugins/parsers/jtinative/pbj/pbj.proto",
	"plugins/parsers/jtinative/pfe_ifl_oc/pfe_ifl_oc.proto",
	// Conflict Extension registration number. - PR1426871
	//"plugins/parsers/jtinative/pfe_mpls_sr_egress_oc/pfe_mpls_sr_egress_oc.proto",
	//"plugins/parsers/jtinative/pfe_mpls_sr_te_bsid_oc/pfe_mpls_sr_te_bsid_oc.proto",
	//"plugins/parsers/jtinative/pfe_mpls_sr_te_ip_oc/pfe_mpls_sr_te_ip_oc.proto",
	"plugins/parsers/jtinative/pfe_mpls_sr_ingress_oc/pfe_mpls_sr_ingress_oc.proto",
	"plugins/parsers/jtinative/pfe_mpls_sr_sid_egress_oc/pfe_mpls_sr_sid_egress_oc.proto",
	"plugins/parsers/jtinative/pfe_mpls_sr_sid_oc/pfe_mpls_sr_sid_oc.proto",
	"plugins/parsers/jtinative/pfe_npu_resource/pfe_npu_resource.proto",
	"plugins/parsers/jtinative/pfe_port_oc/pfe_port_oc.proto",
	"plugins/parsers/jtinative/pfed_oc/pfed_oc.proto",
	"plugins/parsers/jtinative/port/port.proto",
	"plugins/parsers/jtinative/port_exp/port_exp.proto",
	"plugins/parsers/jtinative/qmon/qmon.proto",
	"plugins/parsers/jtinative/rmopd_render/rmopd_render.proto",
	"plugins/parsers/jtinative/rpd_bgp_rib_oc/rpd_bgp_rib_oc.proto",
	"plugins/parsers/jtinative/rpd_ipv6_ra_oc/rpd_ipv6_ra_oc.proto",
	"plugins/parsers/jtinative/rpd_isis_oc/rpd_isis_oc.proto",
	"plugins/parsers/jtinative/rpd_loc_rib_oc/rpd_loc_rib_oc.proto",
	"plugins/parsers/jtinative/rpd_ni_bgp_oc/rpd_ni_bgp_oc.proto",
	"plugins/parsers/jtinative/rpd_rsvp_oc/rpd_rsvp_oc.proto",
	"plugins/parsers/jtinative/rpd_te_oc/rpd_te_oc.proto",
	"plugins/parsers/jtinative/session_telemetry/session_telemetry.proto",
	"plugins/parsers/jtinative/smid_oc/smid_oc.proto",
	"plugins/parsers/jtinative/sr_stats_per_if_egress/sr_stats_per_if_egress.proto",
	"plugins/parsers/jtinative/sr_stats_per_if_ingress/sr_stats_per_if_ingress.proto",
	"plugins/parsers/jtinative/sr_stats_per_sid/sr_stats_per_sid.proto",
	"plugins/parsers/jtinative/svcset_telemetry/svcset_telemetry.proto",
	"plugins/parsers/jtinative/vrrpd_oc/vrrpd_oc.proto",
	"plugins/parsers/jtinative/xmlproxyd_show_local_interface_OC/xmlproxyd_show_local_interface_OC.proto",
}

var messageDescriptorMap map[string]*descriptor.DescriptorProto = make(map[string]*descriptor.DescriptorProto)

func init() {
	for _, path := range protoFiles {
		fd, err := extractFile(proto.FileDescriptor(path))
		if err == nil {
			for _, desc := range fd.MessageType {
				messageDescriptorMap[*desc.Name] = desc
			}
		} else {
			log.Printf("Error in proto file: %s, %s", path, err)
		}
	}
}

type JTINativeParser struct {
	DefaultTags                  map[string]string
	JTINativeMeasurementOverride []map[string]string
	JTINativeTagOverride         []map[string]string
	JTINativeConvertTag          []string
	JTINativeConvertField        []string
	JTIStrAsTag                  bool
	measurementGlobs             []glob.Glob
	measurementNames             []string
	tagGlobs                     []glob.Glob
	tagNames                     []string
}

func (p *JTINativeParser) BuildOverrides() {
	log.Println("I! JTI String as Tag: ", p.JTIStrAsTag)
	log.Println("I! JTI Tag Convert Array: ", p.JTINativeConvertTag)
	log.Println("I! JTI Field Convert Array: ", p.JTINativeConvertField)
	for _, t := range p.JTINativeMeasurementOverride {
		for k, v := range t {
			g := glob.MustCompile(k)
			p.measurementGlobs = append(p.measurementGlobs, g)
			p.measurementNames = append(p.measurementNames, v)
			log.Printf("I! JTI Measurement Overrides: %s %s", k, v)
		}
	}
	for _, t := range p.JTINativeTagOverride {
		for k, v := range t {
			g := glob.MustCompile(k)
			p.tagGlobs = append(p.tagGlobs, g)
			p.tagNames = append(p.tagNames, v)
			log.Printf("I! JTI Tag Overrides: %s %s", k, v)
		}
	}
}

func (p *JTINativeParser) getMeasurementName(sensor string) string {
	for i, g := range p.measurementGlobs {
		if g.Match(sensor) {
			return p.measurementNames[i]
		}
	}
	return sensor
}

func (p *JTINativeParser) getTagName(path []string, name string) string {
	fullPath := strings.Join(append(path, name), ".")
	for i, g := range p.tagGlobs {
		if g.Match(fullPath) {
			return p.tagNames[i]
		}
	}
	return name
}

func (p *JTINativeParser) getTagConvert(path string) bool {
	for _, s := range p.JTINativeConvertTag {
		if s == path {
			return true
		}
	}
	return false
}

func (p *JTINativeParser) getFieldConvert(path string) bool {
	for _, s := range p.JTINativeConvertField {
		if s == path {
			return true
		}
	}
	return false
}

func (p *JTINativeParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	startTime := time.Now()
	ts := &telemetry_top.TelemetryStream{}
	err := proto.Unmarshal(buf, ts)
	if err != nil {
		log.Printf("W! proto unmarshaling error: %s", err)
		log.Printf("D! proto unmarshaling error: %s", buf)
		return nil, err
	}

	TelegrafMetrics := make([]telegraf.Metric, 0)
	timestampMs := int64(ts.GetTimestamp())
	timestamp := time.Unix(0, int64(timestampMs*1000000))

	sensorName := ts.GetSensorName()
	splitSensor := strings.Split(ts.GetSensorName(), ":")
	if len(splitSensor) > 1 {
		sensorName = splitSensor[1]
	}

	tag := map[string]string{
		"device":       ts.GetSystemId(),
		"sensor":       sensorName,
		"measurement":  p.getMeasurementName(sensorName),
		"component":    fmt.Sprintf("%v", ts.GetComponentId()),
		"subcomponent": fmt.Sprintf("%v", ts.GetSubComponentId()),
	}
	for k, v := range p.DefaultTags {
		tag[k] = v
	}

	em, err := proto.GetExtension(ts.Enterprise, telemetry_top.E_JuniperNetworks)
	if err == nil {
		if message, ok := em.(proto.Message); ok {
			for _, ext := range proto.RegisteredExtensions(message) {
				ep, err := proto.GetExtension(message, ext)
				if err == nil {
					if pm, ok := ep.(proto.Message); ok {
						p.dump(pm, timestamp, tag, make([]string, 0), &TelegrafMetrics)
					}
				}
			}
		}
	}
	log.Printf("D! Device: %s, Sensor: %s took %s producing %d metrics", tag["device"], tag["sensor"], time.Since(startTime), len(TelegrafMetrics))
	return TelegrafMetrics, nil
}

func (p *JTINativeParser) ParseLine(line string) (telegraf.Metric, error) {
	// Expecting protobuf message not a single line
	// no need to implement
	return nil, errors.New("ParseLine is not implemented")
}

func (p *JTINativeParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *JTINativeParser) dump(pb proto.Message, timestamp time.Time, tags map[string]string, path []string, TelegrafMetrics *[]telegraf.Metric) {

	ourTags := make(map[string]string)
	for k, v := range tags {
		ourTags[k] = v
	}

	metrics := make(map[string]interface{})
	nestedPbm := make(map[string][]proto.Message)

	pbv := reflect.ValueOf(pb)
	pbv = pbv.Elem()
	for i := 0; i < pbv.NumField(); i++ {
		field := pbv.Field(i)
		sfield := pbv.Type().Field(i)
		if strings.HasPrefix(sfield.Name, "XXX_") {
			continue
		}

		switch field.Kind() {
		case reflect.Invalid:
			log.Printf("D! Invalid: %s", strings.Join(append(path, sfield.Name), "."))
			continue
		case reflect.Slice:
			// Repeated Messages
			for i := 0; i < field.Len(); i++ {
				sliceval := field.Index(i)
				if pm, ok := sliceval.Interface().(proto.Message); ok {
					nestedPbm[sfield.Name] = append(nestedPbm[sfield.Name], pm)
				} else {
					log.Printf("D! Device: %s, Sensor: %s, Path: %s; Slice not message %v", ourTags["device"], ourTags["sensor"], strings.Join(path, "."), sliceval)
				}
			}
		case reflect.Ptr:
			ptrval := field.Elem()
			if ptrval.Kind() == reflect.Invalid {
				// Skip pointers to empty PB messages.
				continue
			}
			if pm, ok := field.Interface().(proto.Message); ok {
				nestedPbm[sfield.Name] = append(nestedPbm[sfield.Name], pm)
				continue
			}
			isTag := (ptrval.Kind() == reflect.String && p.JTIStrAsTag)

			protoDescriptor := messageDescriptorMap[proto.MessageName(pb)]
			if len(protoDescriptor.Field) >= i {
				fieldDescriptor := protoDescriptor.Field[i]
				iskey, err := checkIfKeyOptions(fieldDescriptor)
				if err == nil {
					isTag = iskey
				}
			}
			if p.getTagConvert(strings.Join(append(path, sfield.Name), ".")) {
				isTag = true
			}
			if p.getFieldConvert(strings.Join(append(path, sfield.Name), ".")) {
				isTag = false
			}
			if isTag {
				ourTags[p.getTagName(path, sfield.Name)] = fmt.Sprintf("%v", ptrval.Interface())
				continue
			}
			metrics[strings.Join(append(path, sfield.Name), ".")] = ptrval.Interface()
		}
	}
	for name, idx := range nestedPbm {
		for _, pm := range idx {
			p.dump(pm, timestamp, ourTags, append(path, name), TelegrafMetrics)
		}
	}

	if len(metrics) > 0 {
		measurement := ourTags["measurement"]
		delete(ourTags, "measurement")
		TelegrafMetric, err := metric.New(measurement, ourTags, metrics, timestamp)
		if err == nil {
			*TelegrafMetrics = append(*TelegrafMetrics, TelegrafMetric)
		}
	}
}

func checkIfKeyOptions(fd *descriptor.FieldDescriptorProto) (bool, error) {
	ex, err := proto.GetExtension(fd.Options, telemetry_top.E_TelemetryOptions)
	if err != nil {
		return false, err
	}

	if message, ok := ex.(proto.Message); ok {
		opt := reflect.ValueOf(message)
		isKeyFunc := opt.MethodByName("GetIsKey")
		outputSlice := isKeyFunc.Call(nil)
		if len(outputSlice) > 0 {
			switch outputSlice[0].Kind() {
			case reflect.Bool:
				return outputSlice[0].Bool(), nil
			}
		}
	}
	return false, nil
}

func extractFile(gz []byte) (*descriptor.FileDescriptorProto, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress descriptor: %v", err)
	}

	fd := new(descriptor.FileDescriptorProto)
	if err := proto.Unmarshal(b, fd); err != nil {
		return nil, fmt.Errorf("malformed FileDescriptorProto: %v", err)
	}
	return fd, nil
}
