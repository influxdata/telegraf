package dirmon

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
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
	"github.com/rjeczalik/notify"
	"github.com/srclosson/telegraf-plugins/plugins/parsers/fileinfo"
)

//DirDefObject structure
type DirDefObject struct {
	Directory      string
	DirInclude     []string
	DirExclude     []string
	FileInclude    []string
	FileExclude    []string
	NumProcessors  int
	FieldReplace   map[string]string
	FileRegex      map[string]string
	FileTagRegex   map[string]string
	MetricTagRegex map[string]string
	TempExtension  string
	Timezone       string `toml:"vqtcsv_timezone"` //newly added

	histQueue      chan string
	rtQueue        chan string
	location       *time.Location
	metricTagMatch map[string]*regexp.Regexp
	fiParser       *fileinfo.FileInfoParser
	parser         parsers.Parser
	acc            telegraf.Accumulator
}

//DirMon structure
type DirMon struct {
	Directory    []DirDefObject
	FieldReplace map[string]string

	currDir DirDefObject
	parsers.Parser
	acc telegraf.Accumulator
	//vqtcsvTimeZone string `toml:"vqtcsv_timezone"` //newly added
}

const sampleConfig = `
	## Directories to monitor
	directories = ["D:\Data\FAInputData\Suseela_Enterprise"] 
	// ["D:\Data\InputData\DCInputData\Incoming"]
	## Data format to consume. Only influx is supported
	data_format = "influx"
	
`

//SampleConfig method
func (dm *DirMon) SampleConfig() string {
	return sampleConfig
}

//Description method
func (dm *DirMon) Description() string {
	return "Monitor a directory for DL Files"
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
	log.Println("I am here 9: ", fileName)
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

	strcontent := string(content)
	if strings.Contains(strcontent, "\r\n") {
		return strings.Split(string(content), "\r\n"), nil
	} else if strings.Contains(strcontent, "\n") {
		return strings.Split(string(content), "\n"), nil
	}

	return nil, nil
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

//HashID method
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

//ProcessFile method
func (ddo *DirDefObject) ProcessFile(id int, fileName string, acc telegraf.Accumulator) error {
	ddo.fiParser.SetRelativeDir(ddo.Directory)
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
	if ddo.parser != nil {
		//s, err := getFileScanner(fileName)
		//if err != nil {
		//	log.Println("ERROR [getFileScanner]", err)
		//	return err
		//}

		//for s.Scan() {
		fileLines, err := fileHandler(fileName)
		log.Println("filelines: ", fileLines)
		if err != nil {
			return err
		}
		//line := s.Text()
		groupedMetrics := make(map[uint64][]telegraf.Metric)
		for _, line := range fileLines {
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
		}

		for _, metrics := range groupedMetrics {
			metric := metrics[0]
			for i := 1; i < len(metrics); i++ {
				m := metrics[i]
				for fieldkey, fieldval := range m.Fields() {
					metric.AddField(fieldkey, fieldval)
				}
			}

			for key, regex := range ddo.metricTagMatch {
				match := regex.FindStringSubmatch(fileName)
				if len(match) > 1 {
					metric.AddTag(key, match[1])
				}
			}

			if len(metric.Name()) == 0 {
				metric.SetName("dirmon")
			}
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}

	return nil
}

//IsDirMatch method
func (ddo *DirDefObject) IsDirMatch(strMatch string) bool {
	return ddo.IsDirInclude(strMatch) && !ddo.IsDirExclude(strMatch)
}

//IsDirInclude method
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

//IsDirExclude method
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

//IsFileMatch method
func (ddo *DirDefObject) IsFileMatch(strMatch string) bool {
	return ddo.IsFileInclude(strMatch) && !ddo.IsFileExclude(strMatch)
}

//IsFileInclude method
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

//IsFileExclude method
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

//OSReadDir method
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

//FileProcessor method
func (ddo DirDefObject) FileProcessor(id int) {
	var filename string
	log.Println("File Process i value ", id)
	log.Println("File in rt queue : ", len(ddo.rtQueue))
	log.Println("File in hist queue : ", len(ddo.histQueue))
	//for true {
	if len(ddo.rtQueue) <= 20 || len(ddo.histQueue) <= 20 {
		for i := 0; i < 20; i++ {
			select {
			case filename = <-ddo.rtQueue:
			case filename = <-ddo.histQueue:
			}

			log.Println("In file processor, filename is : ", filename)
			err := ddo.ProcessFile(id, filename, ddo.acc)
			if err != nil {
				log.Printf("E!: Error processing file [%i] [%s]", id, filename) //this is correct
			}
		}
	}
}

//HistoryHandler method
func (ddo DirDefObject) HistoryHandler(dir string, files []string) {
	log.Printf("[DIR](%d): %s\n", len(files), dir)

	for _, file := range files {
		ddo.histQueue <- file
	}
	log.Printf("Backlog completed [%s]", dir)
	log.Println("Hist queue : ", len(ddo.histQueue))
}

//AddToRtQueue method
func (ddo DirDefObject) AddToRtQueue(fileName string) {
	time.Sleep(5 * time.Second)
	log.Println("In AddtoRtQueue")
	ddo.rtQueue <- fileName

}

//RealtimeHandler method
func (ddo DirDefObject) RealtimeHandler(dir string) {
	var eventChan = make(chan notify.EventInfo, 10)
	log.Println("I am in the Realtime handler : ", len(eventChan))

	if err := notify.Watch(dir, eventChan, notify.Rename|notify.Create); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(eventChan)

	// Handle event channel. Queue up items if we are not ready.
	for true {
		eventName := <-eventChan

		fileName := strings.Replace(eventName.Path(), "\\", "/", -1)

		if ddo.IsFileMatch(fileName) {
			log.Println("In file match: ", fileName)
			go ddo.AddToRtQueue(fileName)
		}
	}
}

//Start method
func (ddo DirDefObject) Start(acc telegraf.Accumulator, parser parsers.Parser, gFieldReplace map[string]string) error {
	var err error

	ddo.histQueue = make(chan string, ddo.NumProcessors)
	ddo.rtQueue = make(chan string, ddo.NumProcessors)
	ddo.acc = acc

	for key, value := range ddo.FieldReplace {
		gFieldReplace[key] = value
	}
	ddo.FieldReplace = gFieldReplace

	// ddo.location, err = time.LoadLocation(ddo.Timezone)
	// if err != nil {
	// 	log.Fatalln("FATAL [timezone]: ", err)
	// }

	ddo.parser = parser
	ddo.fiParser, err = fileinfo.NewFileInfoParser(ddo.FileRegex, ddo.FileTagRegex)
	if err != nil {
		return err
	}

	ddo.metricTagMatch = make(map[string]*regexp.Regexp)
	for key, sRegex := range ddo.MetricTagRegex {
		ddo.metricTagMatch[key] = regexp.MustCompile(sRegex)
	}
	results, err := ddo.OSReadDir(ddo.Directory)
	log.Println("results : ", len(results))

	if err != nil {
		log.Fatalln("ERROR [receiver]: ", err)
	}
	if results == nil {
		log.Fatalln("ERROR [results]: No directory found to monitor")
	}

	var totalFiles = 0

	for dir, files := range results {
		go ddo.HistoryHandler(dir, files)
		go ddo.RealtimeHandler(dir)
		log.Println("The directory in results is : ", dir)
		log.Println("The # of files is : ", len(files))
		totalFiles += len(files)

	}
	log.Println("Total files : ", totalFiles)

	// Main processing loop
	for i := 0; i < ddo.NumProcessors; i++ {
		log.Println("value of i is : ", i)
		go ddo.FileProcessor(i)
	}

	return nil
}

//Start method
func (dm *DirMon) Start(acc telegraf.Accumulator) error {
	dm.acc = acc
	// Create a monitor for each directory
	for _, d := range dm.Directory {
		if err := d.Start(acc, dm.Parser, dm.FieldReplace); err != nil {
			log.Println("Error starting", d.Directory)
			return err
		}

	}

	return nil
}

//SetParser method
func (dm *DirMon) SetParser(p parsers.Parser) {
	dm.Parser = p
}

//Gather method
func (dm *DirMon) Gather(_ telegraf.Accumulator) error {
	return nil
}

//Stop method
func (dm *DirMon) Stop() {
}

func init() {
	inputs.Add("dirmon", func() telegraf.Input {
		return &DirMon{}
	})
}
