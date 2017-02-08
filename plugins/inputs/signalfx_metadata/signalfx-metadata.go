package signalfxMetadata

import (
	"log"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// plugin_version
const pluginVersion = "0.0.29"

var sampleConfig = `
  ## SignalFx metadata plugin reports metadata properties for the host
`

// NewSFXMeta - returns a new SignalFx metadata plugin context
func NewSFXMeta() *SFXMeta {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &SFXMeta{
		nextMetadataSend: 0,
		nextMetadataSendInterval: []int64{
			r.Int63n(60),
			60,
			r.Int63n(60) + 3600,
			r.Int63n(600) + 86400},
		aws:         NewAWSInfo(),
		processInfo: NewProcessInfo(),
	}
}

// SFXMeta - struct context for the SignalFx metadata plugin
type SFXMeta struct {
	nextMetadataSend         int64
	nextMetadataSendInterval []int64
	aws                      *AWSInfo
	processInfo              *ProcessInfo
}

// Description - Description of the SignalFx metadata plugin
func (s *SFXMeta) Description() string {
	return "Send host metadata to SignalFx"
}

// SampleConfig - Returns the sample configuration
func (s *SFXMeta) SampleConfig() string {
	return sampleConfig
}

func (s *SFXMeta) sendNotifications(acc telegraf.Accumulator) {
	var info = make(map[string]string)
	GetCPUInfo(info)
	GetKernelInfo(info)
	s.aws.GetAWSInfo(info)
	GetMemory(info)
	info["host_metadata_version"] = pluginVersion
	// info["host_telegraf_version"] = s.getTelegrafVersion()

	// Emit the properties
	for prop, value := range info {
		if err := emitProperty(acc, prop, value); err != nil {
			log.Println("E! Input [signalfx-metadata] ", err)
		}
	}
}

// Gather - read method for SignalFx metadata plugin
func (s *SFXMeta) Gather(acc telegraf.Accumulator) error {

	top, err := s.processInfo.GetTop()
	if err == nil {
		emitTop(acc, string(top), pluginVersion)
	}

	if s.nextMetadataSend == 0 {
		dither := s.nextMetadataSendInterval[0]
		// Pop off the interval
		s.nextMetadataSendInterval = s.nextMetadataSendInterval[1:]
		s.nextMetadataSend = time.Now().Add(time.Duration(dither) * time.Second).Unix()
		log.Printf("I! adding small dither of %v seconds before sending notifications", dither)
	}
	if s.nextMetadataSend < time.Now().Unix() {
		s.sendNotifications(acc)
		if len(s.nextMetadataSendInterval) > 1 {
			dither := s.nextMetadataSendInterval[0]
			s.nextMetadataSendInterval = s.nextMetadataSendInterval[1:]
			s.nextMetadataSend = time.Now().Add(time.Duration(dither) * time.Second).Unix()
			log.Printf("I! Input [signalfx-metadata] till next metadata %v seconds", s.nextMetadataSend-time.Now().Unix())
		} else {
			s.nextMetadataSend = time.Now().Add(time.Duration(s.nextMetadataSendInterval[0]) * time.Second).Unix()
		}
	}

	return nil
}

func (s *SFXMeta) getTelegrafVersion() string {
	var response string
	return response
}

func init() {
	inputs.Add("signalfx-metadata", func() telegraf.Input {
		return NewSFXMeta()
	})
}

func emitProperty(acc telegraf.Accumulator, property string, value string) error {
	var tags = make(map[string]string)
	var fields = make(map[string]interface{})
	tags["sf_metric"] = "objects.host-meta-data"
	tags["property"] = property
	tags["plugin"] = "signalfx-metadata"
	tags["severity"] = "4"
	fields["value"] = value
	if value != "" && property != "" {
		acc.AddGauge("signalfx-metadata", fields, tags, time.Now())
	}
	return nil
}

func emitTop(acc telegraf.Accumulator, top string, version string) {
	var tags = make(map[string]string)
	var fields = make(map[string]interface{})
	tags["sf_metric"] = "objects.top-info"
	tags["plugin"] = "signalfx-metadata"
	tags["severity"] = "4"
	tags["version"] = version
	fields["value"] = top
	if top != "" {
		acc.AddGauge("signalfx-metadata", fields, tags, time.Now())
	}
}
