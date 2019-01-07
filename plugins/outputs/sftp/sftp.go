package sftp

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
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpItem struct {
	Source   string
	Relative string
	Dest     string
	Temp     string
	Meta     map[string]string
	Data     []byte
	Files    []string
	Created  time.Time
}

type Sftp struct {
	Incoming      string
	Outgoing      string
	Error         string
	FilePattern   string
	Destination   string
	Username      string
	Password      string
	DataDir       string
	TempExtension string
	Concurrency   int
	Concatenate   int
	ForceMax      bool
	MinSize       int
	MaxConcat     int
	MaxDuration   time.Duration

	pq        *goque.PrefixQueue
	queue     chan SftpItem
	fileRegex *regexp.Regexp
	outdir    string
	conn      []*sftp.Client
	writer    io.Writer
	closers   []io.Closer
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

func (s *Sftp) Transferer(id int, conn *sftp.Client) {
	for true {
		item := <-s.queue

		// Create the destination file
		log.Printf("Sending (%d)[%d]: %s\n", id, len(item.Data), item.Dest)
		dstFile, err := conn.Create(item.Temp)
		if err != nil {
			// We could try to create the dest dir, but for now... just throw the file away
			log.Printf("ERROR [sftp.create] [%s]: %s", item.Dest, err)
			item.Move(s.Incoming, s.Outgoing, s.Error, false)
			continue
		}

		// Move the file
		_, err = dstFile.Write(item.Data)
		if err != nil {
			log.Println("ERROR [sftp.write]: ", err)
			item.Move(s.Incoming, s.Outgoing, s.Error, false)
			continue
		}

		item.Move(s.Incoming, s.Outgoing, s.Error, true)

		dstFile.Close()
		// Rename the file if we are using a temporary extension
		if item.Temp != item.Dest {
			err = conn.Rename(item.Temp, item.Dest)
			if err != nil {
				log.Println("ERROR [sftp.rename]: ", err)
				continue
			}
		}
	}
}

func (s *SftpItem) Move(inDir string, outDir string, errorDir string, success bool) {
	if len(inDir) > 0 && len(outDir) > 0 && len(errorDir) > 0 {
		files := append(s.Files, s.Relative)
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

func (s *Sftp) HandleSftpItem(item *SftpItem) {
	id := []byte(item.Meta["id"])
	queueItem, err := s.pq.Peek(id)
	if err == nil {
		var masterItem SftpItem
		err := queueItem.ToObject(&masterItem)
		if err != nil {
			log.Println("ERROR [peek]: ", err)
		}

		masterItem.Data = append(masterItem.Data, item.Data...)
		masterItem.Files = append(masterItem.Files, item.Relative)
		if len(masterItem.Data) > s.MinSize || len(masterItem.Files) > s.MaxConcat {
			s.pq.Dequeue(id)
			s.queue <- masterItem
		} else {
			_, err = s.pq.UpdateObject(id, queueItem.ID, masterItem)
			if err != nil {
				log.Println("ERROR [update]: ", err)
			}
		}
	} else {
		if len(item.Data) > s.MinSize {
			s.queue <- *item
		} else {
			s.pq.EnqueueObject([]byte(id), item)
		}
	}
}

func (s *Sftp) GetDestName(tags map[string]string) string {
	outdir := s.outdir
	if strings.Contains(outdir, "{{") {
		split := strings.Split(outdir, "{{")
		for i := 1; i < len(split); i++ {
			tag := split[i][0:strings.Index(split[i], "}}")]
			outdir = strings.Replace(outdir, "{{"+tag+"}}", tags[tag], -1)
		}
	}

	return outdir
}

func (s *Sftp) NewSftpItem(relative_path string, tags map[string]string) *SftpItem {
	var err error
	item := new(SftpItem)

	item.Relative = relative_path
	name := s.Incoming + relative_path
	inbase := filepath.Base(name)
	match := s.fileRegex.FindStringSubmatch(inbase)
	item.Meta = make(map[string]string)
	for i, name := range s.fileRegex.SubexpNames() {
		if i != 0 && name != "" {
			item.Meta[name] = match[i]
		}
	}

	item.Data, err = ioutil.ReadFile(name)
	if err != nil {
		log.Println("ERROR [sftp.ReadFile]: ", err)
	}

	item.Source = name
	item.Dest = fmt.Sprintf("%s/%s", s.GetDestName(tags), inbase)
	item.Temp = item.Dest

	if len(s.TempExtension) > 0 {
		item.Temp += s.TempExtension
	}

	return item
}

func (s *Sftp) CreateConnection(pwg *sync.WaitGroup) {
	defer pwg.Done()
	sshCommonConfig := ssh.Config{
		Ciphers: []string{
			"3des-cbc",
			"blowfish-cbc",
			"aes128-cbc",
			"aes128-ctr",
			"aes256-ctr",
		},
	}
	sshConfig := &ssh.ClientConfig{
		User: s.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		Config:          sshCommonConfig,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	token := strings.Index(s.Destination, "/")
	sshdest := s.Destination[0:token]
	s.outdir = s.Destination[token:]

	for i := 0; i < 10; i++ {
		sshConn, err := ssh.Dial("tcp", sshdest, sshConfig)
		if err != nil {
			log.Println("ERROR [ssh]: ", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}
		// Connect to the sftp server
		sftpconn, err := sftp.NewClient(sshConn)
		if err != nil {
			log.Println("ERROR [sftp]: ", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}

		s.conn = append(s.conn, sftpconn)
		break
	}

}

func (s *Sftp) Connect() error {
	var err error

	if os.RemoveAll(s.DataDir) != nil {
		log.Fatal("Remove Prefix Queue Data Directory", err)
	}

	s.pq, err = goque.OpenPrefixQueue(s.DataDir)
	if err != nil {
		log.Fatalln("ERROR [prefixQueue]: ", err)
	}
	s.queue = make(chan SftpItem, s.Concurrency)
	s.fileRegex = regexp.MustCompile(s.FilePattern)

	var wg sync.WaitGroup

	for i := 0; i < s.Concurrency; i++ {
		wg.Add(1)
		go s.CreateConnection(&wg)
	}
	wg.Wait()

	for i, sftpconn := range s.conn {
		go s.Transferer(i, sftpconn)
	}

	return nil
}

func (s *Sftp) Close() error {
	return nil
}

func (s *Sftp) SampleConfig() string {
	return sampleConfig
}

func (s *Sftp) Description() string {
	return "Send telegraf metrics to sftp server(s)"
}

func (s *Sftp) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		fields := metric.Fields()
		tags := metric.Tags()
		if fields["relative"] == nil {
			continue
		}

		infile := s.NewSftpItem(fields["relative"].(string), tags)
		s.HandleSftpItem(infile)
	}

	return nil
}

func init() {
	outputs.Add("sftp", func() telegraf.Output {
		return &Sftp{}
	})
}
