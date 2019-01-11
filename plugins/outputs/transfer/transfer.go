package transfer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/srclosson/telegraf/filter"
)

type TransferItem struct {
	source string
	dest   []*url.URL
	err    *url.URL
}

type TransferEntry struct {
	Destination   []string
	Error         string
	Verbose       int
	Retries       int
	RetryWait     string
	TempExtension string
	Tagpass       map[string][]string
	Namepass      map[string][]string
	Fieldpass     map[string][]string

	retryWait time.Duration
	tfrs      map[string]Transferer
	tagpass   map[string][]filter.Filter
	namepass  map[string][]filter.Filter
	fieldpass map[string][]filter.Filter
}

type Transfer struct {
	Entry        []*TransferEntry
	RemoveSource int
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

func (t *TransferEntry) TransferItem(id int, item *TransferItem) {
	for _, dest := range item.dest {
		tfr := t.tfrs[dest.Scheme]
		destFile := dest.Path
		if len(t.TempExtension) > 0 {
			dest.Path = dest.Path + t.TempExtension
		}
		for i := 0; i < t.Retries; i++ {
			// Strategy: if there is no error directory, send until retries
			// expire, or successful. if there is an error directory, send
			// there and move on.
			err := tfr.Send(item.source, dest)
			if err != nil {
				if err != nil {
					log.Printf("[%d] ERROR [%s] (transfer.Send): %v => %v [%v]", id, dest.Scheme, item.source, dest, err)
					break
				}
				if item.err != nil {
					etfr := t.tfrs[item.err.Scheme]
					err := etfr.Send(item.source, item.err)
					if err != nil {
						log.Printf("[%d] ERROR [%s] (error.Send): %v => %v [%v]", id, item.err.Scheme, item.source, item.err, err)
						break
					}
				}

				// We don't have an error directory, and the send failed.
				// Try again, after a configurable wait
				time.Sleep(t.retryWait)
				continue
			}

			// Send successful
			// Rename if we are using a temp extension
			if len(t.TempExtension) > 0 {
				tfr.Rename(dest, destFile)
			}

			if t.Verbose == 1 {
				log.Printf("[%d] Sent: %s", id, destFile)
			}
			break
		}
	}
}

func (t *TransferEntry) Connect() error {
	var err error

	t.tfrs = make(map[string]Transferer)

	for _, dest := range t.Destination {
		var name string
		urlSplit := strings.Split(dest, ":")
		if len(urlSplit) > 0 {
			name = urlSplit[0]
		}
		switch name {
		case "file":
			if _, found := t.tfrs[name]; !found {
				t.tfrs["file"] = NewFileTransferer()
			}
			break
		case "ftp":
			if _, found := t.tfrs[name]; !found {
				t.tfrs["ftp"] = NewFtpTransferer()
			}
			break
		case "sftp":
			if _, found := t.tfrs[name]; !found {
				t.tfrs["sftp"] = NewSftpTransferer()
			}
			break
		default:
			return errors.New("Unknown file transferer")
		}
	}

	t.namepass = make(map[string][]filter.Filter)
	for name, value := range t.Namepass {
		filter, err := filter.Compile(value)
		if err != nil {
			return err
		}

		t.namepass[name] = append(t.namepass[name], filter)
	}

	t.tagpass = make(map[string][]filter.Filter)
	for name, value := range t.Tagpass {
		filter, err := filter.Compile(value)
		if err != nil {
			return err
		}

		t.tagpass[name] = append(t.tagpass[name], filter)
	}

	t.fieldpass = make(map[string][]filter.Filter)
	for name, value := range t.Fieldpass {
		filter, err := filter.Compile(value)
		if err != nil {
			return err
		}

		t.fieldpass[name] = append(t.fieldpass[name], filter)
	}

	t.retryWait, err = time.ParseDuration(t.RetryWait)
	if err != nil {
		return err
	}

	return nil
}

func (t *Transfer) SampleConfig() string {
	return sampleConfig
}

func (t *Transfer) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (t *TransferEntry) Template(in string, attributes map[string]interface{}) string {
	tmpl, err := template.New("temp").Parse(in)
	if err != nil {
		fmt.Println("E! [template]", err)
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, attributes)
	if err != nil {
		fmt.Println("E! [template.Execute]", err)
	}

	out := b.String()
	return out
}

func (t *TransferItem) AddDest(d string) error {
	u, err := url.Parse(d)
	if err != nil {
		return err
	}
	t.dest = append(t.dest, u)
	return nil
}

func (t *TransferEntry) Write(source string, attributes map[string]interface{}, pwg *sync.WaitGroup) error {
	defer pwg.Done()
	errTemplate := t.Template(t.Error, attributes)
	errUrl, err := url.Parse(errTemplate)
	if err != nil || len(errUrl.String()) == 0 {
		errUrl = nil
	}
	transfer := &TransferItem{
		source: source,
		err:    errUrl,
	}
	for _, dest := range t.Destination {
		d := t.Template(dest, attributes)
		err := transfer.AddDest(d)
		if err != nil {
			log.Println("E! ERROR adding destination")
		}
	}

	t.TransferItem(0, transfer)

	return nil
}

func (t *Transfer) Connect() error {
	for _, entry := range t.Entry {
		err := entry.Connect()
		if err != nil {
			log.Println("E! [Connect]", err)
			return err
		}
	}

	return nil
}

func (t *Transfer) Write(metrics []telegraf.Metric) error {
	var wg sync.WaitGroup
	for _, metric := range metrics {
		if metric.Name() != "fileinfo" {
			log.Printf("E!: Only fileinfo format accepted by transfer output")
			continue
		}
		if !metric.HasField("dir") {
			log.Println("E!: Fileinfo needs 'dir' field", metric)
			continue
		}
		if !metric.HasField("base") {
			log.Println("E! Fileinfo needs 'base' field", metric)
			continue
		}

		fields := metric.Fields()
		tags := metric.Tags()
		source := fmt.Sprintf("%s/%s", fields["dir"].(string), fields["base"].(string))
		attributes := make(map[string]interface{})
		for key, val := range fields {
			attributes[key] = val
		}

		for key, val := range tags {
			attributes[key] = val
		}

		for _, entry := range t.Entry {
			filtered := false
			for _, filters := range entry.namepass {
				for _, filter := range filters {
					if !filter.Match(metric.Name()) {
						filtered = true
					}
				}
			}

			for id, filters := range entry.tagpass {
				tag := metric.Tags()[id]
				for _, filter := range filters {
					if !filter.Match(tag) {
						filtered = true
					}
				}
			}

			for id, filters := range entry.fieldpass {
				field := metric.Fields()[id].(string)
				for _, filter := range filters {
					if !filter.Match(field) {
						filtered = true
					}
				}
			}

			if !filtered {
				wg.Add(1)
				go entry.Write(source, attributes, &wg)
			}
		}
		wg.Wait()

		if t.RemoveSource == 1 {
			err := os.Remove(source)
			if err != nil {
				log.Printf("ERROR (Remove): %v", err)
			}
		}
	}
	return nil
}

func (t *Transfer) Close() error {
	return nil
}

func init() {
	outputs.Add("transfer", func() telegraf.Output {
		return &Transfer{}
	})
}
