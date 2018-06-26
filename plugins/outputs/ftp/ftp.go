package ftp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beeker1121/goque"
	"github.com/jlaffaye/ftp"
	"github.com/srclosson/telegraf/plugins/outputs"

	"github.com/influxdata/telegraf"
)

type FtpItem struct {
	Source   string
	Relative string
	Dest     string
	Meta     map[string]string
	Data     []byte
	Files    []string
	Created  time.Time
}

type Ftp struct {
	Incoming    string
	Outgoing    string
	Error       string
	FilePattern string
	Destination string
	Username    string
	Password    string
	DataDir     string
	Concurrency int
	Concatenate int
	ForceMax    bool
	MinSize     int
	MaxConcat   int
	MaxDuration time.Duration

	pq        *goque.PrefixQueue
	queue     chan FtpItem
	fileRegex *regexp.Regexp
	outdir    string
	conn      []*ftp.ServerConn
	writer    io.Writer
	closers   []io.Closer
}

var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metricf.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (f *Ftp) Transferer(id int, conn *ftp.ServerConn) {
	for true {
		item := <-f.queue

		// Send the file
		fmt.Printf("Sending (%d)[%d]: %s\n", id, len(item.Data), item.Dest)
		r := bytes.NewReader(item.Data)
		err := conn.StorFrom(item.Dest, r, 0)
		if err != nil {
			// We could try to create the dest dir, but for now... just throw the file away
			log.Printf("ERROR [ftp.storfrom] [%s]: %s", item.Dest, err)
			item.Move(f.Incoming, f.Outgoing, f.Error, false)
		}

		item.Move(f.Incoming, f.Outgoing, f.Error, true)
	}
}

func (f *FtpItem) Move(inDir string, outDir string, errorDir string, success bool) {
	if len(inDir) > 0 && len(outDir) > 0 && len(errorDir) > 0 {
		files := append(f.Files, f.Relative)
		for _, filename := range files {
			if success {
				// Move to Archive dir
				err := os.Rename(inDir+filename, outDir+filename)
				if err != nil {
					log.Println("ERROR [outgoing.rename]", err)
				}
			} else {
				// Move to Bad dir
				err := os.Rename(inDir+filename, errorDir+filename)
				if err != nil {
					log.Println("ERROR [error.rename]", err)
				}
			}
		}
	}
}

func (f *Ftp) HandleFtpItem(item *FtpItem) {
	id := []byte(item.Meta["id"])
	queueItem, err := f.pq.Peek(id)
	if err == nil {
		var masterItem FtpItem
		err := queueItem.ToObject(&masterItem)
		if err != nil {
			fmt.Println("ERROR [peek]: ", err)
		}

		masterItem.Data = append(masterItem.Data, item.Data...)
		masterItem.Files = append(masterItem.Files, item.Relative)
		if len(masterItem.Data) > f.MinSize || len(masterItem.Files) > f.MaxConcat {
			f.pq.Dequeue(id)
			f.queue <- masterItem
		} else {
			_, err = f.pq.UpdateObject(id, queueItem.ID, masterItem)
			if err != nil {
				fmt.Println("ERROR [update]: ", err)
			}
		}
	} else {
		if len(item.Data) > f.MinSize {
			f.queue <- *item
		} else {
			f.pq.EnqueueObject([]byte(id), item)
		}
	}
}

func (f *Ftp) GetDestName(tags map[string]string) string {
	outdir := f.outdir
	if strings.Contains(outdir, "{{") {
		split := strings.Split(outdir, "{{")
		for i := 1; i < len(split); i++ {
			tag := split[i][0:strings.Index(split[i], "}}")]
			outdir = strings.Replace(outdir, "{{"+tag+"}}", tags[tag], -1)
		}
	}

	return outdir
}

func (f *Ftp) NewFtpItem(relative_path string, tags map[string]string) *FtpItem {
	var err error
	item := new(FtpItem)

	item.Relative = relative_path
	name := f.Incoming + relative_path
	inbase := filepath.Base(name)
	match := f.fileRegex.FindStringSubmatch(inbase)
	item.Meta = make(map[string]string)
	for i, name := range f.fileRegex.SubexpNames() {
		if i != 0 && name != "" {
			item.Meta[name] = match[i]
		}
	}

	item.Data, err = ioutil.ReadFile(name)
	if err != nil {
		log.Println("ERROR [ReadFile]: ", err)
	}

	item.Source = name
	item.Dest = fmt.Sprintf("%s/%s", f.GetDestName(tags), inbase)

	return item
}

func (f *Ftp) CreateConnection(pwg *sync.WaitGroup) {
	defer pwg.Done()
	splitDest := strings.Split(f.Destination, "/")
	if len(splitDest) < 2 {
		log.Fatalf("Invalid ftp destination: %s", f.Destination)
	}

	ftpaddr := splitDest[0]
	conn, err := ftp.Connect(ftpaddr)
	if err != nil {
		log.Fatalf("%s: %s", f.Destination, err)
	}

	f.conn = append(f.conn, conn)
}

func (f *Ftp) Connect() error {
	var err error

	if os.RemoveAll(f.DataDir) != nil {
		log.Fatal("Remove Prefix Queue Data Directory", err)
	}

	f.pq, err = goque.OpenPrefixQueue(f.DataDir)
	if err != nil {
		log.Fatalln("ERROR [prefixQueue]: ", err)
	}
	f.queue = make(chan FtpItem, f.Concurrency)
	f.fileRegex = regexp.MustCompile(f.FilePattern)

	var wg sync.WaitGroup

	for i := 0; i < f.Concurrency; i++ {
		wg.Add(1)
		go f.CreateConnection(&wg)
	}
	wg.Wait()

	for i, ftpconn := range f.conn {
		go f.Transferer(i, ftpconn)
	}

	return nil
}

func (f *Ftp) Close() error {
	return nil
}

func (f *Ftp) SampleConfig() string {
	return sampleConfig
}

func (f *Ftp) Description() string {
	return "Send telegraf metrics to ftp server(s)"
}

func (f *Ftp) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		fields := metric.Fields()
		tags := metric.Tags()
		if fields["relative"] == nil {
			continue
		}

		infile := f.NewFtpItem(fields["relative"].(string), tags)
		f.HandleFtpItem(infile)
	}

	return nil
}

func init() {
	outputs.Add("ftp", func() telegraf.Output {
		return &Ftp{}
	})
}
