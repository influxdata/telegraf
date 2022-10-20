package proto

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/c2h5oh/datasize"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/pkg/errors"
)

const MaxUint64 = ^uint64(0)
const MaxInt64 = int64(MaxUint64 >> 1)

type Proto struct {
	HostURL            string `toml:"host_url"`
	User               string `toml:"user"`
	Password           string `toml:"password"`
	CognitoAppClientID string `toml:"cognito_app_client_id"`
	AWSRegion          string `toml:"aws_region"`
	VerifyTLS          bool   `toml:"verify_tls"`

	Log        telegraf.Logger `toml:"-"`
	serializer serializers.Serializer
	cip        *CognitoIdentityProvider
	doOnce     sync.Once
}

var sampleConfig = ``

func (f *Proto) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *Proto) Connect() error {
	var err error
	f.doOnce.Do(func() {
		config := aws.NewConfig().
			WithRegion(f.AWSRegion).
			WithCredentials(credentials.AnonymousCredentials)
		var sess *session.Session
		sess, err = session.NewSession(config)
		if err != nil {
			f.Log.Error(err)
		}

		f.cip = NewCognitoIdentityProvider(sess,
			aws.String(f.User),
			aws.String(f.Password),
			aws.String(f.CognitoAppClientID))
	})

	return err
}

func (f *Proto) Close() error {
	return nil
}

func (f *Proto) SampleConfig() string {
	return sampleConfig
}

func (f *Proto) Description() string {
	return "Send telegraf metrics as protobuf structure to service"
}

func (f *Proto) Write(metrics []telegraf.Metric) error {
	influx := Influx{}
	for _, metric := range metrics {
		b, err := f.serializer.Serialize(metric)
		if err != nil {
			f.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		switch metric.Name() {
		case "kernel":
			m := Kernel{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Kernel = append(influx.Kernel, &m)
		case "linux_sysctl_fs":
			m := LinuxSysctlFs{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.LinuxSysctlFs = append(influx.LinuxSysctlFs, &m)
		case "system":
			m := System{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.System = append(influx.System, &m)
		case "net":
			m := Net{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Net = append(influx.Net, &m)
		case "netstat":
			m := Netstat{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Netstat = append(influx.Netstat, &m)
		case "interrupts":
			m := Interrupts{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Interrupts = append(influx.Interrupts, &m)
		case "processes":
			m := Processes{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Processes = append(influx.Processes, &m)
		case "disk":
			m := Disk{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Disk = append(influx.Disk, &m)
		case "diskio":
			m := Diskio{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Diskio = append(influx.Diskio, &m)
		case "docker":
			m := DockerStats{
				Fields: &DockerStats_FIELDS{
					NContainers:          MaxInt64,
					NContainersPaused:    MaxInt64,
					NContainersRunning:   MaxInt64,
					NContainersStopped:   MaxInt64,
					NCpus:                MaxInt64,
					NGoroutines:          MaxInt64,
					NImages:              MaxInt64,
					NListenerEvents:      MaxInt64,
					NUsedFileDescriptors: MaxInt64,
					MemoryTotal:          MaxInt64,
				},
			}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.DockerStats = append(influx.DockerStats, &m)
		case "docker_container_mem":
			m := DockerMem{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.DockerMem = append(influx.DockerMem, &m)
		case "docker_container_cpu":
			m := DockerCpu{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.DockerCpu = append(influx.DockerCpu, &m)
		case "docker_container_net":
			m := DockerNet{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.DockerNet = append(influx.DockerNet, &m)
		case "docker_container_blkio":
			m := DockerBlkio{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.DockerBlkio = append(influx.DockerBlkio, &m)
		case "mem":
			m := Mem{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Mem = append(influx.Mem, &m)
		case "cpu":
			m := CPU{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Cpu = append(influx.Cpu, &m)
		case "swap":
			m := Swap{}
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			influx.Swap = append(influx.Swap, &m)
		case "tegrastats":
			m := Tegrastats{}

			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build tegrastats")
			}

			// Workaround for largest_free_block_size which is string representation of size eg. 4MB, 5kB
			blockSizeStr := gjson.Get(string(b), "fields.largest_free_block_size")
			if blockSizeStr.Exists() {
				blockSize, err := datasize.ParseString(blockSizeStr.String())
				if err != nil {
					return errors.Wrap(err, "build tegrastat parse largest_free_block_size")
				}

				m.Fields.LargestFreeBlockSizeMb = blockSize.MBytes()
			}

			influx.Tegrastats = append(influx.Tegrastats, &m)
		case "smart_device":
			m := SMART{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build smart_device")
			}
			influx.Smart = append(influx.Smart, &m)
		case "lte":
			m := LTE{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build lte")
			}
			influx.Lte = append(influx.Lte, &m)
		case "gps":
			m := GPS{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build gps")
			}
			influx.Gps = append(influx.Gps, &m)
		case "ublox-data":
			m := GPS{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build gps")
			}
			influx.Gps = append(influx.Gps, &m)
		case "pp_correction_metrics":
			m := PPCorrection{
				FieldsMap: map[string]string{},
				Tags:      &PPCorrection_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.PpCorrection = append(influx.PpCorrection, &m)
		case "able_metrics":
			m := Able{
				FieldsMap: map[string]string{},
				Tags:      &Able_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.Able = append(influx.Able, &m)
		case "camera_control_metrics":
			m := CameraControl{
				FieldsMap: map[string]string{},
				Tags:      &CameraControl_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.CameraControl = append(influx.CameraControl, &m)
		case "uploader_metrics":
			m := Uploader{
				FieldsMap: map[string]string{},
				Tags:      &Uploader_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.Uploader = append(influx.Uploader, &m)
		case "uploader_stats":
			m := UploaderStats{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build UploaderStats")
			}
			influx.UploaderStats = append(influx.UploaderStats, &m)
		case "segnet_metrics":
			m := SegNet{
				FieldsMap: map[string]string{},
				Tags:      &SegNet_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.SegNet = append(influx.SegNet, &m)
		case "detectnet_metrics":
			m := DetectNet{
				FieldsMap: map[string]string{},
				Tags:      &DetectNet_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.Detectnet = append(influx.Detectnet, &m)
		case "recorder_metrics":
			m := Recorder{
				FieldsMap: map[string]string{},
				Tags:      &Recorder_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.Recorder = append(influx.Recorder, &m)
		case "event_data_cutter_metrics":
			m := EventDataCutter{
				FieldsMap: map[string]string{},
				Tags:      &EventDataCutter_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				m.FieldsMap[k] = fmt.Sprintf("%v", v)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.EventDataCutter = append(influx.EventDataCutter, &m)
		case "event_data_cutter_stats":
			m := EventDataCutterStats{
				FieldsMap: map[string]int64{},
				Tags:      &EventDataCutterStats_Tags{},
				Name:      metric.Name(),
				Timestamp: float64(metric.Time().UTC().UnixNano()),
			}

			for k, v := range metric.Fields() {
				vv, ok := v.(float64)
				if !ok {
					continue
				}

				m.FieldsMap[k] = int64(vv)
			}

			m.Tags.Host, _ = metric.GetTag("host")
			influx.EventDataCutterStats = append(influx.EventDataCutterStats, &m)
		case "glog":
			m := Glog{}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build glog")
			}
			influx.Glog = append(influx.Glog, &m)
		case "wireless":
			m := Wireless{
				Fields: &Wireless_Fields{
					Status:       MaxInt64,
					Link:         MaxInt64,
					Level:        MaxInt64,
					Noise:        MaxInt64,
					Nwid:         MaxInt64,
					Crypt:        MaxInt64,
					Frag:         MaxInt64,
					Retry:        MaxInt64,
					Misc:         MaxInt64,
					MissedBeacon: MaxInt64,
				},
			}
			if err := json.Unmarshal(b, &m); err != nil {
				return errors.Wrap(err, "build wireless")
			}
			influx.Wireless = append(influx.Wireless, &m)
		}

	}

	accessToken, err := f.cip.GetIdAccessToken()
	if err != nil {
		return errors.Wrapf(err, "[outputs.proto] unable to get access token")
	}
	b, err := protobuf.Marshal(&influx)
	if err != nil {
		return errors.Wrap(err, "[outputs.proto]")
	}

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(b); err != nil {
		return errors.Wrap(err, "[outputs.proto]")
	}
	if err := g.Close(); err != nil {
		return errors.Wrap(err, "[outputs.proto]")
	}

	req, err := http.NewRequest(http.MethodPost, f.HostURL, &buf)
	if err != nil {
		return errors.Wrap(err, "[outputs.proto]")
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *accessToken))
	if f.VerifyTLS == false {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "[outputs.proto]")
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.Wrap(err, "[outputs.proto] failed to send metrics")
	}
	return err
}

func init() {
	outputs.Add("proto", func() telegraf.Output {
		return &Proto{
			VerifyTLS: true,
		}
	})
}
