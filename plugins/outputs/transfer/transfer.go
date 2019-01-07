package transfer

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type TransferItem struct {
	source string
	dest   []*url.URL
	err    *url.URL
}

type Transfer struct {
	Destination []string
	Error       string
	Concurrency int
	Verbose     int
	Retries     int
	RetryWait   string

	retryWait time.Duration
	tfrs      map[string]Transferer
	ch        chan TransferItem
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

func (t *Transfer) Transferer(id int) {
	for true {
		item := <-t.ch
		for _, dest := range item.dest {
			tfr := t.tfrs[dest.Scheme]
			for i := 0; i < t.Retries; i++ {
				// Strategy: if there is no error directory, send until retries
				// expire, or successful. if there is an error directory, send
				// there and move on.
				err := tfr.Send(item.source, dest)
				if err != nil {
					if item.err != nil {
						etfr := t.tfrs[item.err.Scheme]
						err := etfr.Send(item.source, item.err)
						if err != nil {
							log.Printf("[%d] ERROR (CopyFile): %v", id, err)
							break
						}
					}

					// We don't have an error directory, and the send failed.
					// Try again, after a configurable wait
					time.Sleep(t.retryWait)
					continue
				}

				// Send successful
				log.Printf("[%d] Sent: %s", id, dest)
				break
			}

		}

		err := os.Remove(item.source)
		if err != nil {
			log.Printf("[%d] ERROR (Remove): %v", id, err)
		}
	}
}

func (t *Transfer) Connect() error {
	var err error

	t.retryWait, err = time.ParseDuration(t.RetryWait)
	if err != nil {
		return err
	}

	t.ch = make(chan TransferItem, t.Concurrency)
	for i := 0; i < t.Concurrency; i++ {
		go t.Transferer(i)
	}
	return nil
}

func (t *Transfer) SampleConfig() string {
	return sampleConfig
}

func (t *Transfer) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (t *Transfer) Template(in string, attributes map[string]interface{}) string {
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

func (t *Transfer) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		if metric.Name() == "fileinfo" {
			fields := metric.Fields()
			tags := metric.Tags()
			attributes := make(map[string]interface{})
			for key, val := range fields {
				attributes[key] = val
			}

			for key, val := range tags {
				attributes[key] = val
			}
			errTemplate := t.Template(t.Error, attributes)
			errUrl, err := url.Parse(errTemplate)
			if err != nil || len(errUrl.String()) == 0 {
				errUrl = nil
			}
			transfer := TransferItem{
				source: fmt.Sprintf("%s/%s", fields["dir"].(string), fields["base"].(string)),
				err:    errUrl,
			}
			for _, destination := range t.Destination {
				dest := t.Template(destination, attributes)
				transfer.AddDest(dest)
			}

			t.ch <- transfer
		} else {
			log.Printf("E!: Only fileinfo format accepted by transfer output")
		}
	}
	return nil
}

func (t *Transfer) Close() error {
	return nil
}

func init() {
	outputs.Add("transfer", func() telegraf.Output {
		tfrs := make(map[string]Transferer)
		tfrs["ftp"] = NewFtpTransferer()
		tfrs["file"] = NewFileTransferer()
		tfrs["sftp"] = NewSftpTransferer()
		return &Transfer{
			tfrs: tfrs,
		}
	})
}
