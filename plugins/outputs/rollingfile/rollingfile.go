package rollingfile

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type RollingFile struct {
	Header         string
	MaxFileSize    int
	MaxFileTime    int
	OpenDirectory  string
	OpenExtension  string
	CloseDirectory string
	CloseExtension string
	HostName       string
	FileName       string

	fullName  string
	closeName string
	numBytes  int
	startTime time.Time
	writer    io.Writer
	closer    io.Closer

	serializer serializers.Serializer
}

var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (f *RollingFile) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

func (f *RollingFile) Connect() error {
	return f.Open()
}

func (f *RollingFile) Open() error {
	f.startTime = time.Now()
	var of *os.File
	var err error

	filename := fmt.Sprintf("%s_%03d_%03d_%s_%s",
		f.startTime.Format("20060102_150405"),
		f.startTime.Nanosecond()/1000000,
		(f.startTime.Nanosecond()/1000)%1000,
		f.HostName,
		f.FileName)

	f.fullName = f.OpenDirectory + "/" + filename + "." + f.OpenExtension
	f.closeName = f.CloseDirectory + "/" + filename + "." + f.CloseExtension

	if _, err := os.Stat(f.fullName); os.IsNotExist(err) {
		of, err = os.Create(f.fullName)
	} else {
		of, err = os.OpenFile(f.fullName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	}

	if err != nil {
		return err
	}

	f.writer = of
	f.closer = of
	f.numBytes = 0

	if f.Header != "" {
		f.writer.Write([]byte(f.Header))
	}

	return nil
}

func (f *RollingFile) Close() error {
	err := f.closer.Close()
	if err != nil {
		return err
	}

	filein, err := os.Open(f.fullName)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(filein)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	filein.Close()

	// Delete the original file
	err = os.Remove(f.fullName)
	if err != nil {
		return err
	}

	// Open file for writing.
	fileout, _ := os.OpenFile(f.closeName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	// Write compressed data.
	if strings.HasSuffix(f.closeName, ".gz") {
		w := gzip.NewWriter(fileout)
		w.Write(content)
		w.Flush()
		w.Close()
	} else {
		w := bufio.NewWriter(fileout)
		w.Write(content)
		w.Flush()
	}
	fileout.Close()

	return nil
}

func (f *RollingFile) SampleConfig() string {
	return sampleConfig
}

func (f *RollingFile) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (f *RollingFile) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		b, err := f.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}
		_, err = f.writer.Write(b)
		if err != nil {
			return fmt.Errorf("failed to write message: %s, %s", b, err)
		}

		f.numBytes += len(b)
	}

	now := time.Now()
	dur := now.Sub(f.startTime)
	if (f.numBytes >= f.MaxFileSize) || (dur.Seconds() >= float64(f.MaxFileTime)) {
		err := f.Close()
		if err != nil {
			log.Printf("[On close]: %s", err)
		}
		f.Open()
	}
	return nil
}

func init() {
	outputs.Add("rollingfile", func() telegraf.Output {
		return &RollingFile{}
	})
}
