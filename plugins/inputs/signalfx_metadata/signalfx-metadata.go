package signalfxMetadata

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// plugin_version
const pluginVersion = "0.0.30"

var sampleConfig = `
  ## SignalFx metadata plugin reports metadata properties for the host
  ## Process List Collection Settings
  ## boolean indicating whether to emit proccess list information
  # ProcessInfo = true
  ## number of go routines used to collect the process list (must be 1 or greater)
  # NumberOfGoRoutines = 3
  ## The buffer size should be greater than or equal to the length of all 
  ## processes on the host
  # BufferSize = 10000
`

// NewSFXMeta - returns a new SignalFx metadata plugin context
func NewSFXMeta() *SFXMeta {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &SFXMeta{
		BufferSize:         10000,
		NumberOfGoRoutines: 3,
		ProcessInfo:        true,
		nextMetadataSend:   0,
		nextMetadataSendInterval: []int64{
			r.Int63n(60),
			60,
			r.Int63n(60) + 3600,
			r.Int63n(600) + 86400},
		aws: NewAWSInfo(),
	}
}

// SFXMeta - struct context for the SignalFx metadata plugin
type SFXMeta struct {
	BufferSize               int
	NumberOfGoRoutines       int
	ProcessInfo              bool
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
	var infoFunctions = []func() map[string]string{
		GetCPUInfo,
		GetKernelInfo,
		GetMemory,
		s.aws.GetAWSInfo,
	}
	wg := &sync.WaitGroup{}
	for _, funct := range infoFunctions {
		wg.Add(1)
		go func(f func() map[string]string) {
			i := f()
			// Emit the properties
			for prop, value := range i {
				if err := emitProperty(acc, prop, value); err != nil {
					log.Println("E! Input [signalfx-metadata] ", err)
				}
			}
			wg.Done()
		}(funct)
	}
	if err := emitProperty(acc, "host_metadata_version", pluginVersion); err != nil {
		log.Println("E! Input [signalfx-metadata] ", err)
	}
	wg.Wait()
}

// Gather - read method for SignalFx metadata plugin
func (s *SFXMeta) Gather(acc telegraf.Accumulator) error {
	if s.processInfo == nil && s.ProcessInfo {
		s.processInfo = NewProcessInfo(s.BufferSize, s.NumberOfGoRoutines)
	}
	wg := sync.WaitGroup{}
	if s.ProcessInfo {
		log.Println("D! Input [signalfx-metadata] collecting process info")
		wg.Add(1)
		go func() {
			top, err := s.processInfo.GetTop()
			if err == nil {
				emitTop(acc, top, pluginVersion)
			}
			wg.Done()
		}()
	}

	if s.nextMetadataSend == 0 {
		dither := s.nextMetadataSendInterval[0]
		// Pop off the interval
		s.nextMetadataSendInterval = s.nextMetadataSendInterval[1:]
		s.nextMetadataSend = time.Now().Add(time.Duration(dither) * time.Second).Unix()
		log.Printf("I! Input [signalfx-metadata] adding small dither of %v seconds before sending notifications \n", dither)
	}
	if s.nextMetadataSend < time.Now().Unix() {
		s.sendNotifications(acc)
		if len(s.nextMetadataSendInterval) > 1 {
			dither := s.nextMetadataSendInterval[0]
			s.nextMetadataSendInterval = s.nextMetadataSendInterval[1:]
			s.nextMetadataSend = time.Now().Add(time.Duration(dither) * time.Second).Unix()
			log.Printf("I! Input [signalfx-metadata] till next metadata %v seconds\n", s.nextMetadataSend-time.Now().Unix())
		} else {
			s.nextMetadataSend = time.Now().Add(time.Duration(s.nextMetadataSendInterval[0]) * time.Second).Unix()
		}
	}
	wg.Wait()
	return nil
}

func init() {
	inputs.Add("signalfx-metadata", func() telegraf.Input {
		return NewSFXMeta()
	})
}

func emitProperty(acc telegraf.Accumulator, property string, value string) error {
	if value != "" && property != "" {
		acc.AddGauge("signalfx-metadata",
			map[string]interface{}{ // fields
				"value": value,
			},
			map[string]string{ // tags
				"sf_metric": "objects.host-meta-data",
				"property":  property,
				"plugin":    "signalfx-metadata",
				"severity":  "4",
			},
			time.Now())
	}
	return nil
}

func emitTop(acc telegraf.Accumulator, top string, version string) {
	if top != "" {
		acc.AddGauge("signalfx-metadata",
			map[string]interface{}{
				"value": top,
			},
			map[string]string{ // tags
				"sf_metric": "objects.top-info",
				"plugin":    "signalfx-metadata",
				"severity":  "4",
				"version":   version,
			},
			time.Now())
	}
}
