package webdav

import (
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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/studio-b12/gowebdav"
)

type WebdavItem struct {
	Source   string
	Relative string
	Dest     string
	Meta     map[string]string
	Data     []byte
	Files    []string
	Created  time.Time
}

type Webdav struct {
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
	queue     chan WebdavItem
	fileRegex *regexp.Regexp
	outdir    string
	clients   []*gowebdav.Client
	writer    io.Writer
	closers   []io.Closer
}

var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metricw.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (w *Webdav) Transferer(id int, client *gowebdav.Client) {
	for true {
		item := <-w.queue

		// Create the destination file
		fmt.Printf("Sending (%d)[%d]: %s\n", id, len(item.Data), item.Dest)
		err := client.Write(item.Dest, item.Data, os.ModeExclusive)
		if err != nil {
			// We could try to create the dest dir, but for now... just throw the file away
			log.Printf("ERROR [webdav.write] [%s]: %s", item.Dest, err)
			item.Move(w.Incoming, w.Outgoing, w.Error, false)
			continue
		}

		// Move the file
		item.Move(w.Incoming, w.Outgoing, w.Error, true)
	}
}

func (wi *WebdavItem) Move(inDir string, outDir string, errorDir string, success bool) {
	if len(inDir) > 0 && len(outDir) > 0 && len(errorDir) > 0 {
		files := append(wi.Files, wi.Relative)
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

func (w *Webdav) HandleWebdavItem(item *WebdavItem) {
	id := []byte(item.Meta["id"])
	queueItem, err := w.pq.Peek(id)
	if err == nil {
		var masterItem WebdavItem
		err := queueItem.ToObject(&masterItem)
		if err != nil {
			fmt.Println("ERROR [peek]: ", err)
		}

		masterItem.Data = append(masterItem.Data, item.Data...)
		masterItem.Files = append(masterItem.Files, item.Relative)
		if len(masterItem.Data) > w.MinSize || len(masterItem.Files) > w.MaxConcat {
			w.pq.Dequeue(id)
			w.queue <- masterItem
		} else {
			_, err = w.pq.UpdateObject(id, queueItem.ID, masterItem)
			if err != nil {
				fmt.Println("ERROR [update]: ", err)
			}
		}
	} else {
		if len(item.Data) > w.MinSize {
			w.queue <- *item
		} else {
			w.pq.EnqueueObject([]byte(id), item)
		}
	}
}

func (w *Webdav) GetDestName(tags map[string]string) string {
	outdir := w.outdir
	if strings.Contains(outdir, "{{") {
		split := strings.Split(outdir, "{{")
		for i := 1; i < len(split); i++ {
			tag := split[i][0:strings.Index(split[i], "}}")]
			outdir = strings.Replace(outdir, "{{"+tag+"}}", tags[tag], -1)
		}
	}

	return outdir
}

func (w *Webdav) NewWebdavItem(relative_path string, tags map[string]string) *WebdavItem {
	var err error
	item := new(WebdavItem)

	item.Relative = relative_path
	name := w.Incoming + relative_path
	inbase := filepath.Base(name)
	match := w.fileRegex.FindStringSubmatch(inbase)
	item.Meta = make(map[string]string)
	for i, name := range w.fileRegex.SubexpNames() {
		if i != 0 && name != "" {
			item.Meta[name] = match[i]
		}
	}

	item.Data, err = ioutil.ReadFile(name)
	if err != nil {
		log.Println("ERROR [ReadFile]: ", err)
	}

	item.Source = name
	item.Dest = fmt.Sprintf("%s/%s", w.GetDestName(tags), inbase)

	return item
}

func (w *Webdav) CreateClient(pwg *sync.WaitGroup) {
	defer pwg.Done()
	c := gowebdav.NewClient(w.Destination, w.Username, w.Password)
	err := c.Connect()
	if err != nil {
		log.Println("Could not connect to destination:", w.Destination)
		return
	}
	w.clients = append(w.clients, c)
}

func (w *Webdav) Connect() error {
	var err error

	if os.RemoveAll(w.DataDir) != nil {
		log.Fatal("Remove Prefix Queue Data Directory", err)
	}

	w.pq, err = goque.OpenPrefixQueue(w.DataDir)
	if err != nil {
		log.Fatalln("ERROR [prefixQueue]: ", err)
	}
	w.queue = make(chan WebdavItem, w.Concurrency)
	w.fileRegex = regexp.MustCompile(w.FilePattern)

	var wg sync.WaitGroup

	for i := 0; i < w.Concurrency; i++ {
		wg.Add(1)
		go w.CreateClient(&wg)
	}
	wg.Wait()

	return nil
}

func (w *Webdav) Close() error {
	return nil
}

func (w *Webdav) SampleConfig() string {
	return sampleConfig
}

func (w *Webdav) Description() string {
	return "Send telegraf metrics to sftp server(s)"
}

func (w *Webdav) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		fields := metric.Fields()
		tags := metric.Tags()
		if fields["relative"] == nil {
			continue
		}

		infile := w.NewWebdavItem(fields["relative"].(string), tags)
		w.HandleWebdavItem(infile)
	}

	return nil
}

func init() {
	outputs.Add("webdav", func() telegraf.Output {
		return &Webdav{}
	})
}
