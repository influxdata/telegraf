package docker_cnt_logs

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/influxdata/telegraf"
	"github.com/pkg/errors"

	"io"
	"net/http"
	"sync"
	"time"

	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//Docker client wrapper
type Client interface {
	ContainerInspect(ctx context.Context, contID string) (types.ContainerJSON, error)
	ContainerLogs(ctx context.Context, contID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	Close() error
}

// DockerCNTLogs object
type DockerCNTLogs struct {
	//Passed from config
	Endpoint            string `toml:"endpoint"`
	tlsint.ClientConfig        //Parsing is handled in tlsint module

	InitialChunkSize  int                      `toml:"initial_chunk_size"`
	MaxChunkSize      int                      `toml:"max_chunk_size"`
	OffsetFlush       string                   `toml:"offset_flush"`
	OffsetStoragePath string                   `toml:"offset_storage_path"`
	ShutDownWhenEOF   bool                     `toml:"shutdown_when_eof"`
	TargetContainers  []map[string]interface{} `toml:"container"`

	//Internal
	context context.Context
	//client      *docker.Client
	client      Client
	wg          sync.WaitGroup
	checkerDone chan bool
	//checkerLock         *sync.Mutex
	offsetDone                 chan bool
	logReader                  map[string]*logReader //Log reader data...
	offsetFlushInterval        time.Duration
	disableTimeStampsStreaming bool //Used for simulating reading logs with or without TS (used in tests only)

}

type logReader struct {
	contID     string
	contStream io.ReadCloser

	msgHeaderExamined     bool
	dockerTimeStamps      bool
	interval              time.Duration
	initialChunkSize      int
	currentChunkSize      int
	maxChunkSize          int
	outputMsgStartIndex   uint
	dockerTimeStampLength uint
	buffer                []byte
	leftoverBuffer        []byte
	length                int
	endOfLineIndex        int
	tags                  map[string]string
	done                  chan bool
	eofReceived           bool
	currentOffset         int64
	lock                  *sync.Mutex
}

const defaultInitialChunkSize = 1000
const defaultMaxChunkSize = 5000

const dockerLogHeaderSize = 8
const dockerTimeStampLength = 30

const defaultPolingIntervalNS = 500 * time.Millisecond

const defaultFlushInterval = 3 * time.Second

const sampleConfig = `
  ## Interval to gather data from docker sock.
  ## the longer the interval the fewer request is made towards docker API (less CPU utilization on dockerd).
  ## On the other hand, this increase the delay between producing logs and delivering it. Reasonable trade off
  ## should be chosen
  interval = "2000ms"
  
  ## Docker Endpoint
  ##  To use unix, set endpoint = "unix:///var/run/docker.sock" (/var/run/docker.sock is default mount path)
  ##  To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##  To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Log streaming settings
  ## Set initial chunk size (length of []byte buffer to read from docker socket)
  ## If not set, default value of 'defaultInitialChunkSize = 1000' will be used
  # initial_chunk_size = 1000 # 1K symbols (half of 80x25 screen)

  ## Max chunk size (length of []byte buffer to read from docker socket)
  ## Buffer can grow in capacity adjusting to volume of data received from docker sock
  ## to the maximum volume limited by this parameter. The bigger buffer is set
  ## the more data potentially it can read during 1 API call to docker.
  ## And all of this data will be processed before sending, that increase CPU utilization.
  ## This parameter should be set carefully.
  # max_chunk_size = 5000 # 5K symbols

  ## Offset flush interval. How often the offset pointer (see below) in the
  ## log stream is flashed to file.Offset pointer represents the unix time stamp
  ## in nano seconds for the last message read from log stream (default - 3 sec)
  # offset_flush = "3s"

  ## Offset storage path (mandatory), make sure the user on behalf 
  ## of which the telegraf is started has appropriate rights to read and write to chosen path.
  offset_storage_path = "/var/run/collector_offset"

  ## Shutdown telegraf if all log streaming containers stopped/killed, default - false
  ## This option make sense when telegraf started especially for streaming logs
  ## in a form of sidecar container in k8s. In case primary container exited,
  ## side-car should be terminated also.
  # shutdown_when_eof = false

  ## Settings per container (specify as many sections as needed)
  [[inputs.docker_cnt_logs.container]]
    ## Set container id (long or short from), or container name
    ## to stream logs from, this attribute is mandatory
    id = "dc23d3ea534b3a6ec3934ae21e2dd4955fdbf61106b32fa19b831a6040a7feef"

    ## Override common settings
    ## input interval (specified or inherited from agent section)
    # interval = "500ms"

    ## Initial chunk size
    initial_chunk_size = 2000 # 2K symbols

    ## Max chunk size
    max_chunk_size = 6000 # 6K symbols

    #Set additional tags that will be tagged to the stream from the current container:
    tags = [
        "tag1=value1",
        "tag2=value2"
    ]
  ##Another container to stream logs from  
  [[inputs.docker_cnt_logs.container]]
    id = "009d82030745c9994e2f5c2280571e8b9f95681793a8f7073210759c74c1ea36"
    interval = "600ms"
`

var (
	version        = "1.21" // Support as old version as possible
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

//Service functions
func isContainsHeader(str *[]byte, length int) bool {

	//Docker inject headers when running in detached mode, to distinguish stdout, stderr, etc.
	//Header structure:
	//header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
	//STREAM_TYPE can be:
	//
	//0: stdin (is written on stdout)
	//1: stdout
	//2: stderr
	//SIZE1, SIZE2, SIZE3, SIZE4 are the four bytes of the uint32 size encoded as big endian.
	//
	//Following the header is the payload, which is the specified number of bytes of STREAM_TYPE.

	var result bool
	if length <= 0 || /*garbage*/
		length < dockerLogHeaderSize /*No header*/ {
		return false
	}

	strLength := len(*str)
	if strLength > 100 {
		strLength = 100
	}

	log.Printf("D! [inputs.docker_cnt_logs] Raw string for detecting headers (first 100 symbols):\n%s...\n",
		(*str)[:strLength-1])
	log.Printf("D! [inputs.docker_cnt_logs] First 4 bytes: '%v,%v,%v,%v', string representation: '%s'",
		(*str)[0], (*str)[1], (*str)[2], (*str)[3], (*str)[0:4])
	log.Printf("D! [inputs.docker_cnt_logs] Big endian value: %d",
		binary.BigEndian.Uint32((*str)[4:dockerLogHeaderSize]))

	//Examine first 4 bytes to detect if they match to header structure (see above)
	if ((*str)[0] == 0x0 || (*str)[0] == 0x1 || (*str)[0] == 0x2) &&
		((*str)[1] == 0x0 && (*str)[2] == 0x0 && (*str)[3] == 0x0) &&
		binary.BigEndian.Uint32((*str)[4:dockerLogHeaderSize]) >= 2 /*Encoding big endian*/ {
		//binary.BigEndian.Uint32((*str)[4:dockerLogHeaderSize]) - calculates message length.
		//Minimum message length with timestamp is 32 (timestamp (30 symbols) + space + '\n' = 32.
		//But in case you switch timestamp off it will be 2 (space + '\n')

		log.Printf("I! [inputs.docker_cnt_logs] Detected: log messages from docker API streamed WITH headers...")
		result = true

	} else {
		log.Printf("I! [inputs.docker_cnt_logs] Detected: log messages from docker API streamed WITHOUT headers...")
		result = false
	}

	return result
}

//If there is no new line in interval [eolIndex-HeaderSize,eolIndex+HeaderSize],
//then we are definitely not in the middle of header, otherwise, we are.
func isNewLineInMsgHeader(str *[]byte, eolIndex int) bool {
	//Edge case:
	if eolIndex == dockerLogHeaderSize {
		return false
	}

	//If in the frame there is the following sequence '\n, 0|1|2, 0,0,0',
	// then we are somewhere in the header. First '\n means that there is another
	// srting that ends before this, and we actually need to find this particular '\n'
	for i := eolIndex - dockerLogHeaderSize; i < eolIndex; i++ {
		if ((*str)[i] == '\n') &&
			((*str)[i+1] == 0x1 || (*str)[i+1] == 0x2 || (*str)[i+1] == 0x0) &&
			((*str)[i+2] == 0x0 && (*str)[i+3] == 0x0 && (*str)[i+4] == 0x0) {
			return true
		}
	}

	return false
}

func getInputIntervalDuration(acc telegraf.Accumulator) (dur time.Duration) {
	// As agent.accumulator type is not exported, we need to use reflect, for getting the value of
	// 'interval' for this input
	// The code below should be revised, in case agent.accumulator type would be changed.
	// Anyway, all possible checks on the data structure are made to handle types change.
	emptyValue := reflect.Value{}
	agentAccumulator := reflect.ValueOf(acc).Elem()
	if agentAccumulator.Kind() == reflect.Struct {
		if agentAccumulator.FieldByName("maker") == emptyValue { //Empty value
			log.Printf("W! [inputs.docker_cnt_logs] Error while parsing agent.accumulator type, filed 'maker'"+
				" is not found.\nDefault pooling duration '%d' nano sec. will be used", defaultPolingIntervalNS)
			dur = defaultPolingIntervalNS
		} else {
			runningInput := reflect.Indirect(agentAccumulator.FieldByName("maker").Elem())
			if reflect.Indirect(runningInput).FieldByName("Config") == emptyValue {
				log.Printf("W! [inputs.docker_cnt_logs] Error while parsing models.RunningInput type, filed "+
					"'Config' is not found.\nDefault pooling duration '%d' nano sec. will be used", defaultPolingIntervalNS)
				dur = defaultPolingIntervalNS
			} else {
				if reflect.Indirect(reflect.Indirect(runningInput).FieldByName("Config").Elem()).FieldByName("Interval") == emptyValue {
					log.Printf("W! [inputs.docker_cnt_logs] Error while parsing models.InputConfig type, filed 'Interval'"+
						" is not found.\nDefault pooling duration '%d' nano sec. will be used", defaultPolingIntervalNS)
					dur = defaultPolingIntervalNS

				} else {
					interval := reflect.Indirect(reflect.Indirect(runningInput).FieldByName("Config").Elem()).FieldByName("Interval")
					if interval.Kind() == reflect.Int64 {
						dur = time.Duration(interval.Int())
					} else {
						log.Printf("W! [inputs.docker_cnt_logs] Error while parsing models.RunningInput.Interval type, filed "+
							" is not of type 'int'.\nDefault pooling duration '%d' nano sec. will be used", defaultPolingIntervalNS)
						dur = defaultPolingIntervalNS
					}
				}

			}

		}

	}

	return
}

func getOffset(offsetFile string) (string, int64) {

	if _, err := os.Stat(offsetFile); !os.IsNotExist(err) {
		data, errRead := ioutil.ReadFile(offsetFile)
		if errRead != nil {
			log.Printf("E! [inputs.docker_cnt_logs] Error reading offset file '%s', reason: %s",
				offsetFile, errRead.Error())
		} else {
			timeString := ""
			timeInt, err := strconv.ParseInt(string(data), 10, 64)
			if err == nil {
				timeString = time.Unix(0, timeInt).UTC().Format(time.RFC3339Nano)
			}

			log.Printf("D! [inputs.docker_cnt_logs] Parsed offset from '%s'\nvalue: %s, %s",
				offsetFile, string(data), timeString)
			return timeString, timeInt
		}
	}

	return "", 0
}

//Primary plugin interface
func (dl *DockerCNTLogs) Description() string {
	return "Read logs from docker containers via Docker API"
}

func (dl *DockerCNTLogs) SampleConfig() string { return sampleConfig }

func (dl *DockerCNTLogs) Gather(acc telegraf.Accumulator) error {

	return nil
}

func (dl *DockerCNTLogs) goGather(done <-chan bool, acc telegraf.Accumulator, lr *logReader) {

	var err error

	dl.wg.Add(1)
	defer dl.wg.Done()

	eofReceived := false
	for {
		select {
		case <-done:
			return
		default:
			//Iterative reads by chunks
			// While reading in chunks, there are 2 general cases:
			// 1. Either full buffer (it means that the message either fit to chunkSize or exceed it.
			// To figure out if it exceed we need to check if the buffer ends with "\r\n"

			// 2. Or partially filled buffer. In this case the rest of the buffer is '\0'

			// Read from docker API
			lr.length, err = lr.contStream.Read(lr.buffer) //Can be a case when API returns lr.length==0, and err==nil

			if err != nil {
				if err.Error() == "EOF" { //Special case, need to flush data and exit
					eofReceived = true
				} else {
					select {
					case <-done: //In case the goroutine was signaled, the stream would be closed, to unblock
						//the read operation. That's why we don't need to print error, as it is expected behaviour...
						return
					default:
						acc.AddError(fmt.Errorf("Read error from container '%s': %v", lr.contID, err))
						return
					}

				}
			}

			if !lr.msgHeaderExamined {
				if isContainsHeader(&lr.buffer, lr.length) {
					lr.outputMsgStartIndex = dockerLogHeaderSize //Header is in the string, need to strip it out...
				} else {
					lr.outputMsgStartIndex = 0 //No header in the output, start from the 1st letter.
				}
				lr.msgHeaderExamined = true
			}

			if len(lr.leftoverBuffer) > 0 { //append leftover from previous iteration
				lr.buffer = append(lr.leftoverBuffer, lr.buffer...)
				lr.length += len(lr.leftoverBuffer)

				//Erasing leftover buffer once used:
				lr.leftoverBuffer = nil
			}

			if lr.length != 0 {
				//Docker API fills buffer with '\0' until the end even if there is no data at all,
				//In this case, lr.length == 0 as it shows the amount of actually read data, but len(lr.buffer)
				// will be equal to cap(lr.buffer), as the buffer will be filled out with '\0'
				lr.endOfLineIndex = lr.length - 1
			} else {
				lr.endOfLineIndex = 0
			}

			//1st case
			if lr.length == len(lr.buffer) && lr.length > 0 {
				//Seek last line end (from the end), ignoring the case when this line end is in the message header
				//for ; lr.endOfLineIndex >= 0; lr.endOfLineIndex-- {
				for ; lr.endOfLineIndex >= int(lr.outputMsgStartIndex); lr.endOfLineIndex-- {
					if lr.buffer[lr.endOfLineIndex] == '\n' {

						//Skip '\n' if there are headers and '\n' is inside header
						if lr.outputMsgStartIndex > 0 && isNewLineInMsgHeader(&lr.buffer, lr.endOfLineIndex) {
							continue
						}

						if lr.endOfLineIndex != lr.length-1 {
							// Moving everything that is after lr.endOfLineIndex to leftover buffer (2nd case)
							lr.leftoverBuffer = nil
							lr.leftoverBuffer = make([]byte, (lr.length-1)-lr.endOfLineIndex)
							copy(lr.leftoverBuffer, lr.buffer[lr.endOfLineIndex+1:])
						}
						break
					}
				}

				//Check if line end is not found
				if lr.endOfLineIndex == int(lr.outputMsgStartIndex-1) { //This is 1st case -
					// buffer holds one string that is not terminated
					//We need simply to move it into leftover buffer
					//and grow current chunk size if limit is not exceeded
					lr.leftoverBuffer = nil
					lr.leftoverBuffer = make([]byte, len(lr.buffer))
					copy(lr.leftoverBuffer, lr.buffer)

					//Grow chunk size
					if lr.currentChunkSize*2 < lr.maxChunkSize {
						lr.currentChunkSize = lr.currentChunkSize * 2
						lr.buffer = nil
						lr.buffer = make([]byte, lr.currentChunkSize)
						runtime.GC()
					}

					continue
				}

			}

			//Parsing the buffer line by line and passing data to accumulator
			//Since read from API can return lr.length==0, and err==nil, we need to additionally check the boundaries
			if len(lr.buffer) > 0 && lr.endOfLineIndex > 0 {

				totalLineLength := 0
				var timeStamp time.Time
				var field map[string]interface{}
				//var tags = map[string]string{}
				var tags = lr.tags

				for i := 0; i <= lr.endOfLineIndex; i = i + totalLineLength {
					field = make(map[string]interface{})
					//Checking boundaries:
					if i+int(lr.outputMsgStartIndex) > lr.endOfLineIndex { //sort of garbage
						timeStamp = time.Now()
						field["value"] = fmt.Sprintf("%s", lr.buffer[i:lr.endOfLineIndex])
						acc.AddFields("stream", field, tags, timeStamp)
						break
					}

					//Looking for the end of the line (skipping index)
					totalLineLength = 0
					for j := i + int(lr.outputMsgStartIndex); j <= lr.endOfLineIndex; j++ {
						if lr.buffer[j] == '\n' {
							totalLineLength = j - i + 1 //Include '\n'
							break
						}
					}
					if totalLineLength == 0 {
						totalLineLength = (lr.endOfLineIndex + 1) - i
					}

					//Getting stream type (if header persist)
					if lr.outputMsgStartIndex > 0 {
						if lr.buffer[i] == 0x1 {
							tags["stream"] = "stdout"
						} else if lr.buffer[i] == 0x2 {
							tags["stream"] = "stderr"
						} else if lr.buffer[i] == 0x0 {
							tags["stream"] = "stdin"
						}
					} else {
						tags["stream"] = "interactive"
					}

					if uint(totalLineLength) < lr.outputMsgStartIndex+lr.dockerTimeStampLength+1 || !lr.dockerTimeStamps {
						//no time stamp
						timeStamp = time.Now()
						field["value"] = fmt.Sprintf("%s", lr.buffer[i+int(lr.outputMsgStartIndex):i+totalLineLength])
					} else {
						timeStamp, err = time.Parse(time.RFC3339Nano,
							fmt.Sprintf("%s", lr.buffer[i+int(lr.outputMsgStartIndex):i+int(lr.outputMsgStartIndex)+int(lr.dockerTimeStampLength)]))
						if err != nil {
							acc.AddError(fmt.Errorf("Can't parse time stamp from string, container '%s': "+
								"%v. Raw message string:\n%s\nOutput msg start index: %d",
								lr.contID, err, lr.buffer[i:i+totalLineLength], lr.outputMsgStartIndex))
							log.Printf("E! [inputs.docker_cnt_logs]\n=========== buffer[:lr.endOfLineIndex] ===========\n"+
								"%s\n=========== ====== ===========\n", lr.buffer[:lr.endOfLineIndex])
						}
						field["value"] = fmt.Sprintf("%s",
							lr.buffer[i+int(lr.outputMsgStartIndex)+int(lr.dockerTimeStampLength)+1:i+totalLineLength])
					}

					acc.AddFields("stream", field, tags, timeStamp)
					field = nil

					//Saving offset
					currentOffset := atomic.LoadInt64(&lr.currentOffset)
					atomic.AddInt64(&lr.currentOffset, timeStamp.UTC().UnixNano()-currentOffset+1)
				}
			}

			//Control the size of buffer`
			if len(lr.buffer) > lr.maxChunkSize {
				lr.buffer = nil
				lr.buffer = make([]byte, lr.currentChunkSize)
				runtime.GC()
			}

			if eofReceived {
				log.Printf("E! [inputs.docker_cnt_logs] Container '%s': 'EOF' received.",
					lr.contID)

				lr.lock.Lock()
				lr.eofReceived = eofReceived
				lr.lock.Unlock()

				return
			}

		}

		time.Sleep(lr.interval)
	}

}

func (dl *DockerCNTLogs) Start(acc telegraf.Accumulator) error {
	var err error
	var tlsConfig *tls.Config

	dl.context = context.Background()
	switch dl.Endpoint {
	case "ENV":
		{
			dl.client, err = docker.NewClientWithOpts(docker.FromEnv)
		}
	case "MOCK":
		{
			log.Printf("W! [inputs.docker_cnt_logs] Starting with mock docker client...")
		}
	default:
		{
			tlsConfig, err = dl.ClientConfig.TLSConfig()
			if err != nil {
				return err
			}

			transport := &http.Transport{
				TLSClientConfig: tlsConfig,
			}
			httpClient := &http.Client{Transport: transport}

			dl.client, err = docker.NewClientWithOpts(
				docker.WithHTTPHeaders(defaultHeaders),
				docker.WithHTTPClient(httpClient),
				docker.WithVersion(version),
				docker.WithHost(dl.Endpoint))
		}
	}

	if err != nil {
		return err
	}

	if dl.InitialChunkSize == 0 {
		dl.InitialChunkSize = defaultInitialChunkSize
	} else {
		if dl.InitialChunkSize <= dockerLogHeaderSize {
			dl.InitialChunkSize = 2 * dockerLogHeaderSize
		}
	}

	if dl.MaxChunkSize == 0 {
		dl.MaxChunkSize = defaultMaxChunkSize
	} else {
		if dl.MaxChunkSize <= dl.InitialChunkSize {
			dl.MaxChunkSize = 5 * dl.InitialChunkSize
		}
	}

	//Parsing flush offset
	if dl.OffsetFlush == "" {
		dl.offsetFlushInterval = defaultFlushInterval
	} else {
		dl.offsetFlushInterval, err = time.ParseDuration(dl.OffsetFlush)
		if err != nil {
			dl.offsetFlushInterval = defaultFlushInterval
			log.Printf("W! [inputs.docker_cnt_logs] Can't parse '%s' duration, default value will be used.", dl.OffsetFlush)
		}
	}

	//Create storage path
	if src, err := os.Stat(dl.OffsetStoragePath); os.IsNotExist(err) {
		errDir := os.MkdirAll(dl.OffsetStoragePath, 0755)
		if errDir != nil {
			return errors.Errorf("Can't create directory '%s' to store offset, reason: %s", dl.OffsetStoragePath, errDir.Error())
		}

	} else if src != nil && src.Mode().IsRegular() {
		return errors.Errorf("'%s' already exist as a file!", dl.OffsetStoragePath)
	}

	//Prepare data for running log streaming from containers
	dl.logReader = map[string]*logReader{}

	for _, container := range dl.TargetContainers {

		if _, ok := container["id"]; !ok { //id is not specified
			return errors.Errorf("Mandatory attribute 'id' is not specified for '[[inputs.docker_cnt_logs.container]]' section!")
		}
		logReader := logReader{}
		logReader.contID = container["id"].(string)

		if _, ok := container["interval"]; ok { //inetrval is specified
			logReader.interval, err = time.ParseDuration(container["interval"].(string))
			if err != nil {
				return errors.Errorf("Can't parse interval from string '%s', reason: %s", container["interval"].(string), err.Error())
			}
		} else {
			logReader.interval = getInputIntervalDuration(acc)
		}

		logReader.dockerTimeStamps = !dl.disableTimeStampsStreaming //Default behaviour - stream logs with time-stamps
		logReader.dockerTimeStampLength = dockerTimeStampLength

		//intitial chunk size
		if _, ok := container["initial_chunk_size"]; ok { //initial_chunk_size specified

			if int(container["initial_chunk_size"].(int64)) <= dockerLogHeaderSize {
				logReader.initialChunkSize = 2 * dockerLogHeaderSize
			} else {
				logReader.initialChunkSize = int(container["initial_chunk_size"].(int64))
			}
		} else {
			logReader.initialChunkSize = dl.InitialChunkSize
		}

		//max chunk size
		if _, ok := container["max_chunk_size"]; ok { //max_chunk_size specified

			if int(container["max_chunk_size"].(int64)) <= logReader.initialChunkSize {
				logReader.maxChunkSize = 5 * logReader.initialChunkSize
			} else {
				logReader.maxChunkSize = int(container["max_chunk_size"].(int64))
			}
		} else {
			logReader.maxChunkSize = dl.MaxChunkSize
		}

		logReader.currentChunkSize = logReader.initialChunkSize

		//Gettings target container status (It can be a case when we can attempt
		//to read from the container that already stopped/crashed)
		contStatus, err := dl.client.ContainerInspect(dl.context, logReader.contID)
		if err != nil {
			return err
		}
		getLogsSince := ""
		getLogsSince, logReader.currentOffset = getOffset(path.Join(dl.OffsetStoragePath, logReader.contID))

		if contStatus.State.Status == "removing" ||
			contStatus.State.Status == "exited" || contStatus.State.Status == "dead" {
			log.Printf("W! [inputs.docker_cnt_logs] container '%s' is not running!", logReader.contID)
		}

		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: logReader.dockerTimeStamps,
			Since:      getLogsSince}

		logReader.contStream, err = dl.client.ContainerLogs(dl.context, logReader.contID, options)
		if err != nil {
			return err
		}

		//Parse tags if any
		logReader.tags = map[string]string{}
		if _, ok := container["tags"]; ok { //tags specified
			for _, tag := range container["tags"].([]interface{}) {
				arr := strings.Split(tag.(string), "=")
				if len(arr) != 2 {
					return errors.Errorf("Can't parse tags from string '%s', valid format is <tag_name>=<tag_value>", tag.(string))
				}
				logReader.tags[arr[0]] = arr[1]
			}
		}
		//set container ID tag:
		logReader.tags["container_id"] = logReader.contID

		//Allocate buffer for reading logs
		logReader.buffer = make([]byte, logReader.initialChunkSize)
		logReader.msgHeaderExamined = false

		//Init channel to manage go routine
		logReader.done = make(chan bool)
		logReader.lock = &sync.Mutex{}

		//Store
		dl.logReader[container["id"].(string)] = &logReader
	}

	//Starting log streaming (only afeter full initialization of logger settings performed)
	for _, logReader := range dl.logReader {
		go dl.goGather(logReader.done, acc, logReader)
	}

	//Start checker
	dl.checkerDone = make(chan bool)
	go dl.checkStreamersStatus(dl.checkerDone)

	//Start offset flusher
	dl.offsetDone = make(chan bool)
	go dl.flushOffset(dl.offsetDone)

	return nil
}

func (dl *DockerCNTLogs) shutdownTelegraf() {
	var err error
	var p *os.Process
	p, err = os.FindProcess(os.Getpid())
	if err != nil {
		log.Printf("E! [inputs.docker_cnt_logs] Can't get current process PID "+
			"to initiate graceful shutdown: %v.\nHave to panic for shutdown...", err)
	} else {
		if runtime.GOOS == "windows" {
			err = p.Signal(os.Kill) //Interrupt is not supported on windows
		} else {
			err = p.Signal(os.Interrupt)
		}
		if err != nil {
			log.Printf("W! [inputs.docker_cnt_logs] Can't send signal to main process "+
				"for initiating Telegraf shutdown, reason: %v\nHave to panic for shutdown...", err)
		} else {
			return
		}
	}

	panic(errors.New("Graceful shutdown is not possible, force panic."))
}

func (dl *DockerCNTLogs) checkStreamersStatus(done <-chan bool) {

	dl.wg.Add(1)
	defer dl.wg.Done()

	for {
		select {
		case <-done:
			return
		default:
			closed := 0
			for _, logReader := range dl.logReader {

				logReader.lock.Lock()
				if logReader.eofReceived {
					closed++
				}
				logReader.lock.Unlock()
			}
			if closed == len(dl.logReader) {
				log.Printf("I! [inputs.docker_cnt_logs] All target containers are stopped/killed!")
				if dl.ShutDownWhenEOF {
					log.Printf("I! [inputs.docker_cnt_logs] Telegraf shutdown is requested...")
					dl.shutdownTelegraf()
					return
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (dl *DockerCNTLogs) flushOffset(done <-chan bool) {

	dl.wg.Add(1)
	defer dl.wg.Done()

	for {
		select {
		case <-done:
			return
		default:

			for _, logReader := range dl.logReader {
				filename := path.Join(dl.OffsetStoragePath, logReader.contID)
				offset := []byte(strconv.FormatInt(atomic.LoadInt64(&logReader.currentOffset), 10))
				err := ioutil.WriteFile(filename, offset, 0777)
				if err != nil {
					log.Printf("E! [inputs.docker_cnt_logs] Can't write logger offset to file '%s', reason: %v",
						filename, err)
				}
			}

		}
		time.Sleep(dl.offsetFlushInterval)
	}
}

func (dl *DockerCNTLogs) Stop() {

	log.Printf("D! [inputs.docker_cnt_logs] Shutting down streams checkers...")

	//Stop check streamers status
	close(dl.checkerDone)

	//Stop log streaming
	log.Printf("D! [inputs.docker_cnt_logs] Shutting down log streamers & closing docker streams...")
	for _, logReader := range dl.logReader {
		close(logReader.done) //Signaling go routine to close
		//Unblock goroutine if it waits for the data from stream
		if logReader.contStream != nil {
			if err := logReader.contStream.Close(); err != nil {
				log.Printf("D! [inputs.docker_cnt_logs] Can't close container logs stream, reason: %v", err)
			}

		}
	}

	//Stop offset flushing
	log.Printf("D! [inputs.docker_cnt_logs] Waiting for shutting down offset flusher...")
	time.Sleep(dl.offsetFlushInterval) //This sleep needed to guarantee that offset will be flushed
	close(dl.offsetDone)

	//Wait for all go routines to complete
	dl.wg.Wait()

	if dl.client != nil {
		if err := dl.client.Close(); err != nil {
			log.Printf("D! [inputs.docker_cnt_logs] Can't close docker client, reason: %v", err)
		}
	}

}

func init() {
	inputs.Add("docker_cnt_logs", func() telegraf.Input { return &DockerCNTLogs{} })
}
