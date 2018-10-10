package dirmon

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/fileinfo"
	"github.com/rjeczalik/notify"
)

type DirDefObject struct {
	Name            string
	Incoming        string
	Outgoing        string
	Error           string
	DirInclude      []string
	DirExclude      []string
	FileInclude     []string
	FileExclude     []string
	DataFormat      string
	Tags            []string
	ParseByGroup    bool
	NumProcessors   int
	ConcurrentTasks int
	MetricName      string
	FieldReplace    map[string]string
	FileTagRegex    map[string]string
	TempExtension   string
	Timezone        string

	histQueue    chan string
	rtQueue      chan string
	location     *time.Location
	metricMatch  *regexp.Regexp
	fileTagMatch map[string]*regexp.Regexp
	parser       parsers.Parser
	fiParser     *fileinfo.FileInfoParser
	acc          telegraf.Accumulator
}

type DirMon struct {
	Directory    []DirDefObject
	FieldReplace map[string]string

	currDir DirDefObject
	acc     telegraf.Accumulator
}

const sampleConfig = `
	## Directories to monitor
	directories = ["D:\Data\InputData\DCInputData\Incoming"]
	## Data format to consume. Only influx is supported
	data_format = "influx"
`

func (dm *DirMon) SampleConfig() string {
	return sampleConfig
}

func (dm *DirMon) Description() string {
	return "Monitor a directory for DL Files"
}

func (dm *DirMon) Gather(acc telegraf.Accumulator) error {
	return nil
}

func fileHandlerMultiGzip(fileName string) ([]string, error) {
	var buf bytes.Buffer
	var f *os.File
	var err error

	for i := 0; i < 10; i++ {
		f, err = os.Open(fileName)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	defer f.Close()
	if err != nil {
		log.Println("ERROR: [os.open]:", err)
		// If we can't open the file... ignore and move on
		return nil, err
	}

	extension := filepath.Ext(fileName)
	bw := bufio.NewWriter(&buf)
	br := bufio.NewReader(f)
	var content []byte

	switch extension {
	case ".gz":
		zr, err := gzip.NewReader(br)
		if err != nil {
			log.Println("Error opening gz", fileName, err)
			return nil, err
		}
		defer zr.Close()

		for {
			zr.Multistream(false)
			if _, err := io.Copy(bw, zr); err != nil {
				log.Println("ERROR: [io.copy]:", err)
				return nil, err
			}
			content = append(content, buf.Bytes()...)
			err = zr.Reset(br)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("ERROR: [zr.reset]:", err)
				return nil, err
			}
		}
		break
	default:
		content, err = ioutil.ReadAll(br)
		if err != nil {
			log.Printf("ERROR [read %s]: %s", fileName, err)
			return nil, err
		}
		break
	}

	lines := strings.Split(string(content), "\n")
	return lines, nil
}

func fileHandler(fileName string) ([]string, error) {
	var f *os.File
	var err error
	var r io.Reader

	for i := 0; i < 10; i++ {
		f, err = os.Open(fileName)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	defer f.Close()
	if err != nil {
		log.Println("file open error", err)
		// If we can't open the file... ignore and move on
		return nil, nil
	}

	extension := filepath.Ext(fileName)
	r = bufio.NewReader(f)

	switch extension {
	case ".gz":
		r, err = gzip.NewReader(r)
		if err != nil {
			log.Println("ERROR [gzip.NewReader]:", fileName, err)
			return nil, err
		}
		break
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		log.Printf("ERROR [read %s]: %s", fileName, err)
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	return lines, nil
}

func getFileScanner(fileName string) (*bufio.Scanner, error) {
	var f *os.File
	var err error
	var r io.Reader

	for i := 0; i < 10; i++ {
		f, err = os.Open(fileName)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	defer f.Close()
	if err != nil {
		log.Println("file open error", err)
		// If we can't open the file... ignore and move on
		return nil, nil
	}

	extension := filepath.Ext(fileName)
	r = bufio.NewReader(f)

	switch extension {
	case ".gz":
		r, err = gzip.NewReader(r)
		if err != nil {
			log.Println("ERROR [gzip.NewReader]:", fileName, err)
			return nil, err
		}
		break
	}

	s := bufio.NewScanner(r)

	return s, nil
}

func HashID(metric telegraf.Metric) uint64 {
	h := fnv.New64a()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(metric.Time().UnixNano()))
	h.Write(b)

	tags := metric.Tags()
	tmp := make([]string, len(tags))
	i := 0
	for k, v := range tags {
		tmp[i] = k + v
		i++
	}
	sort.Strings(tmp)

	for _, s := range tmp {
		h.Write([]byte(s))
	}

	return h.Sum64()
}

func MoveFile(from string, to string) error {
	// open files r and w
	var r *os.File
	var err error

	for i := 0; i < 10; i++ {
		r, err = os.Open(from)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	// do the actual work
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	err = w.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (ddo *DirDefObject) MoveFile(id int, filename string, success bool) {
	relativePath := strings.TrimPrefix(filename, ddo.Incoming)
	from := filename
	to := ddo.Outgoing + "/" + relativePath
	bad := ddo.Error + "/" + relativePath
	var final string
	var err error

	if len(ddo.TempExtension) > 0 {
		final = to
		to = to + ddo.TempExtension
	}

	if success {
		// Move to Archive dir
		err := MoveFile(from, to)
		if err != nil {
			log.Println("ERROR [outgoing.rename]", err)
		}
	} else {
		// Move to Bad dir
		err := MoveFile(from, bad)
		if err != nil {
			log.Println("ERROR [error.rename]", err)
		}
	}

	// Delete the original file
	for i := 0; i < 10; i++ {
		err = os.Remove(from)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}
	if err != nil {
		log.Println("ERROR [remove.from]", err)
	}

	// Rename the file to the proper extension
	if len(ddo.TempExtension) > 0 {
		err = os.Rename(to, final)
		if err != nil {
			log.Println("ERROR [rename.to]", err)
		}
	}
}

func (ddo *DirDefObject) ProcessFile(id int, fileName string, acc telegraf.Accumulator) error {
	ddo.fiParser.SetIncomingDir(ddo.Incoming)
	fiMetrics, err := ddo.fiParser.Parse([]byte(fileName))
	if err != nil {
		log.Printf("ERROR [%s]: %s", fileName, err)
		return err
	}

	if fiMetrics != nil {
		if len(fiMetrics) > 1 {
			log.Printf("ERROR [%s]: Expected 1 set of metrics. Found [%d]", fileName, len(fiMetrics))
			return err
		}

		for _, m := range fiMetrics {
			acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}

	// If we are just doing fileinfo... end here.
	if ddo.DataFormat != "fileinfo" {
		s, err := getFileScanner(fileName)
		if err != nil {
			log.Println("ERROR [getFileScanner]", err)
			return err
		}

		for s.Scan() {
			//fileLines, err := fileHandler(fileName)
			//if err != nil {
			//	return err
			//}
			line := s.Text()
			log.Println("line", line)
			groupedMetrics := make(map[uint64][]telegraf.Metric)
			//for _, line := range fileLines {
			if len(line) == 0 {
				continue
			}
			m, err := ddo.parser.ParseLine(line)
			if err != nil {
				log.Printf("ERROR [%s]: %s", fileName, err)
				continue
			}

			if m != nil {
				id := HashID(m)
				groupedMetrics[id] = append(groupedMetrics[id], m)
			}
			//}

			for _, metrics := range groupedMetrics {
				metric := metrics[0]
				for i := 1; i < len(metrics); i++ {
					m := metrics[i]
					for fieldkey, fieldval := range m.Fields() {
						metric.AddField(fieldkey, fieldval)
					}
				}

				name := metric.Name()
				if ddo.metricMatch != nil {
					match := ddo.metricMatch.FindStringSubmatch(fileName)
					if len(match) > 1 {
						name = match[1]
					}
				}

				for key, regex := range ddo.fileTagMatch {
					match := regex.FindStringSubmatch(fileName)
					if len(match) > 1 {
						metric.AddTag(key, match[1])
					}
				}

				acc.AddFields(name, metric.Fields(), metric.Tags(), metric.Time())
			}
		}

	}

	return nil
}

func (ddo *DirDefObject) IsDirMatch(strMatch string) bool {
	return ddo.IsDirInclude(strMatch) && !ddo.IsDirExclude(strMatch)
}

func (ddo *DirDefObject) IsDirInclude(strMatch string) bool {
	isInclude := 0
	for _, r := range ddo.DirInclude {
		b, err := regexp.MatchString(r, strMatch)
		if b && err == nil {
			isInclude++
		}
	}

	return isInclude > 0
}

func (ddo *DirDefObject) IsDirExclude(strMatch string) bool {
	isExclude := 0
	for _, r := range ddo.DirExclude {
		b, err := regexp.MatchString(r, strMatch)
		if b && err == nil {
			isExclude++
		}
	}

	return isExclude > 0
}

func (ddo *DirDefObject) IsFileMatch(strMatch string) bool {
	return ddo.IsFileInclude(strMatch) && !ddo.IsFileExclude(strMatch)
}

func (ddo *DirDefObject) IsFileInclude(strMatch string) bool {
	isInclude := 0
	for _, r := range ddo.FileInclude {
		b, err := regexp.MatchString(r, strMatch)
		if b && err == nil {
			isInclude++
		}
	}

	return isInclude > 0
}

func (ddo *DirDefObject) IsFileExclude(strMatch string) bool {
	isExclude := 0
	for _, r := range ddo.FileExclude {
		b, err := regexp.MatchString(r, strMatch)
		if b && err == nil {
			isExclude++
		}
	}

	return isExclude > 0
}

func (ddo *DirDefObject) OSReadDir(root string) (map[string][]string, error) {
	files := make(map[string][]string)

	if ddo.IsDirInclude(root) {
		files[root] = nil
	}

	if ddo.IsDirExclude(root) {
		return nil, nil
	}

	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileInfo, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range fileInfo {
		filename := root + "/" + file.Name()
		if file.IsDir() {
			f, err := ddo.OSReadDir(filename)
			if err != nil {
				return files, err
			}

			for d, af := range f {
				files[d] = append(files[d], af...)
			}
		} else {
			if ddo.IsDirInclude(root) {
				dirname := path.Dir(filename)
				if ddo.IsFileMatch(filename) {
					files[dirname] = append(files[dirname], filename)
				}
			}
		}
	}
	return files, nil
}

func (ddo DirDefObject) FileProcessor(id int) {
	var filename string

	for true {
		select {
		case filename = <-ddo.rtQueue:
		case filename = <-ddo.histQueue:
		}

		err := ddo.ProcessFile(id, filename, ddo.acc)
		if len(ddo.Outgoing) > 0 {
			if err != nil {
				ddo.MoveFile(id, filename, false)
				continue
			}

			ddo.MoveFile(id, filename, true)
		}
	}
}

func (ddo DirDefObject) HistoryHandler(dir string, files []string) {
	log.Printf("[DIR](%d): %s\n", len(files), dir)

	for _, file := range files {
		ddo.histQueue <- file
	}
	log.Printf("Backlog completed [%s]", dir)
}

func (ddo DirDefObject) AddToRtQueue(fileName string) {
	time.Sleep(5 * time.Second)
	ddo.rtQueue <- fileName
}

func (ddo DirDefObject) RealtimeHandler(dir string) {
	var eventChan = make(chan notify.EventInfo, 10)

	if err := notify.Watch(dir, eventChan, notify.Rename|notify.Create); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(eventChan)

	// Handle event channel. Queue up items if we are not ready.
	for true {
		eventName := <-eventChan
		fileName := strings.Replace(eventName.Path(), "\\", "/", -1)

		if ddo.IsFileMatch(fileName) {
			go ddo.AddToRtQueue(fileName)
		}
	}
}

func (ddo DirDefObject) Start(acc telegraf.Accumulator, gFieldReplace map[string]string) error {
	var err error

	ddo.histQueue = make(chan string, ddo.ConcurrentTasks*ddo.NumProcessors)
	ddo.rtQueue = make(chan string, 2000)
	ddo.acc = acc

	for key, value := range ddo.FieldReplace {
		gFieldReplace[key] = value
	}
	ddo.FieldReplace = gFieldReplace

	ddo.location, err = time.LoadLocation(ddo.Timezone)
	if err != nil {
		log.Fatalln("FATAL [timezone]: ", err)
	}

	args := make(map[string]interface{})
	args["acc"] = ddo.acc
	args["fieldreplace"] = ddo.FieldReplace
	args["location"] = ddo.location

	ddo.fiParser, err = fileinfo.NewFileInfoParser()
	if err != nil {
		return err
	}

	p, err := parsers.NewParser(&parsers.Config{
		DataFormat: ddo.DataFormat,
		MetricName: ddo.MetricName,
		Args:       args,
	})

	if err != nil {
		log.Fatalf("ERROR [parser:%s]: %s", p, err)
	}
	ddo.parser = p

	if strings.HasPrefix(ddo.MetricName, "/") && strings.HasSuffix(ddo.MetricName, "/") {
		// This is a regular expression. Parse it and use it as the name.
		name := strings.Trim(ddo.MetricName, "/")
		ddo.metricMatch = regexp.MustCompile(name)
	}

	ddo.fileTagMatch = make(map[string]*regexp.Regexp)
	for key, sRegex := range ddo.FileTagRegex {
		if strings.HasPrefix(sRegex, "/") && strings.HasSuffix(sRegex, "/") {
			// This is a regular expression. Parse it and use it as the name.
			name := strings.Trim(sRegex, "/")
			ddo.fileTagMatch[key] = regexp.MustCompile(name)
		}
	}

	results, err := ddo.OSReadDir(ddo.Incoming)
	if err != nil {
		log.Fatalln("ERROR [receiver]: ", err)
	}
	if results == nil || len(results) == 0 {
		log.Fatalln("ERROR [results]: No directory found to monitor")
	}

	for dir, files := range results {
		go ddo.HistoryHandler(dir, files)
		go ddo.RealtimeHandler(dir)
	}

	// Main processing loop
	for i := 0; i < ddo.NumProcessors; i++ {
		go ddo.FileProcessor(i)
	}

	return nil
}

func (dm *DirMon) Start(acc telegraf.Accumulator) error {
	dm.acc = acc
	// Create a monitor for each directory
	for _, d := range dm.Directory {
		if err := d.Start(acc, dm.FieldReplace); err != nil {
			log.Println("Error starting", d.Name)
			return err
		}

	}

	return nil
}

func (dm *DirMon) Stop() {
}

func init() {
	fmt.Println("dirmon init...")
	inputs.Add("dirmon", func() telegraf.Input {
		dm := DirMon{}
		return &dm
	})
	fmt.Println("dirmon init done...")
}
