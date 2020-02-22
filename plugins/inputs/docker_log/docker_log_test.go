package docker_log

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/influxdata/telegraf/internal"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//Mock data
var (
	//Ensure mockDockerClient implements dClient wrapper
	_ dClient = (*mockDockerClient)(nil)

	stdOutHeader = []byte{1, 0, 0, 0, 0, 0, 1, 0}
	stdErrHeader = []byte{2, 0, 0, 0, 0, 0, 1, 0}
	stdInHeader  = []byte{0, 0, 0, 0, 0, 0, 1, 0}

	testLogsOutputData = []map[string]interface{}{
		{
			"tss":    "2019-10-29T11:15:14.813957700Z",
			"data":   "0:first message\r\n",
			"header": stdOutHeader,
		},
		{
			"tss":    "2019-10-29T11:15:15.813957700Z",
			"data":   "1:intermediate message\r\n",
			"header": stdErrHeader,
		},
		{
			"tss":    "2019-10-29T11:15:17.813957700Z",
			"data":   "2:last message", //Important! When EOF, last message is not terminated...
			"header": stdInHeader,
		}}
	testLogsOutputDataWOHeaders = []map[string]interface{}{
		{
			"tss":    "2019-10-29T11:15:13.813957700Z",
			"data":   "0:first message\r\n",
			"header": nil,
		},
		{
			"tss":    "2019-10-29T11:15:18.813957700Z",
			"data":   "1:last message", //Important! When EOF, last message is not terminated...
			"header": nil,
		}}

	targetContainers = []map[string]interface{}{
		{
			"id":                  "container1",
			"log_gather_interval": "500ms",
			"rawLogEntries":       testLogsOutputData,
			"hasTTY":              false,
			"status":              "running",
			"tags":                []interface{}{"tag1=StreamWithHeaders", "tag2=value2"}},
		{
			"id":                  "container2",
			"log_gather_interval": "1000ms",
			"rawLogEntries":       testLogsOutputDataWOHeaders,
			"hasTTY":              true,
			"status":              "paused",
			"tags":                []interface{}{"tag1=StreamWithoutHeaders", "tag2=value2"}}}
	targetContainersWOTS = []map[string]interface{}{
		{
			"id":                  "container1",
			"log_gather_interval": "500ms",
			"rawLogEntries":       testLogsOutputDataWOHeaders,
			"hasTTY":              true,
			"status":              "running",
			"tags":                []interface{}{"tag1=StreamWithoutHeaders", "tag2=value2"}},
		{
			"id":                  "container2",
			"log_gather_interval": "1000ms",
			"rawLogEntries":       testLogsOutputData,
			"hasTTY":              false,
			"status":              "paused",
			"tags":                []interface{}{"tag1=StreamWithHeaders", "tag2=value2"}}}
)

type mockDockerClient struct {
	targetContainers []map[string]interface{}
}

func newMockDockerClient(targetContainers []map[string]interface{}) *mockDockerClient {
	return &mockDockerClient{targetContainers: targetContainers}
}

func (c *mockDockerClient) ContainerInspect(ctx context.Context, contID string) (types.ContainerJSON, error) {
	inspectData := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{ID: contID, State: &types.ContainerState{}},
		Config:            &container.Config{}}

	for _, container := range c.targetContainers {
		if container["id"].(string) == contID {
			inspectData.Config.Tty = container["hasTTY"].(bool)
			inspectData.State.Status = container["status"].(string)
		}
	}
	return inspectData, nil
}

func (c *mockDockerClient) ContainerLogs(ctx context.Context, contID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	var dockerLogDataEntry []byte
	var dockerLogMessage string
	var testDataRC mockReaderCloser
	var sinceTS int64
	var entryTS time.Time

	var err error

	for _, container := range c.targetContainers {
		if container["id"] == contID {

			if options.Since != "" {
				sinceTS, err = strconv.ParseInt(options.Since, 10, 64)
				if err != nil && err.Error() == "invalid syntax" {
					sinceTSTime := time.Time{}
					sinceTSTime, err = time.Parse(time.RFC3339Nano, options.Since)
					if err != nil {
						return nil, fmt.Errorf("MOCK dClient: Can't parse timestamp from docker options. Cont ID '%s', reason: %v", contID, err)
					} else {
						sinceTS = sinceTSTime.UnixNano()
					}
				}

				lg.logD("Container '%s', stream logs since: %s (%d)",
					contID, time.Unix(sinceTS, 0).Format(time.RFC3339Nano), sinceTS)
			}
			rawLogEntries := container["rawLogEntries"].([]map[string]interface{})
			logEntriesIncluded := 0
			for _, entry := range rawLogEntries {

				//Filtering log entries based on TS
				entryTS, err = time.Parse(time.RFC3339Nano, entry["tss"].(string))
				if err != nil {
					lg.logD("Container '%s', stream log entry '%s' can't parse ts '%s'\n",
						contID, entry["data"].(string), entry["tss"].(string))
					continue
				}
				if entryTS.Unix() < sinceTS {
					lg.logD("Container '%s', stream log entry '%s' filtered\n"+
						"based on it's ts: %s (%d)!", contID, entry["data"].(string), entry["tss"].(string), entryTS.Unix())
					continue
				}

				if options.Timestamps {
					dockerLogMessage = fmt.Sprintf("%s %s", entry["tss"], entry["data"])
				} else {
					dockerLogMessage = entry["data"].(string)
				}
				if entry["header"] != nil { //No header
					dockerLogDataEntry = append(entry["header"].([]byte),
						[]byte(dockerLogMessage)...)
				} else {
					dockerLogDataEntry = []byte(dockerLogMessage)
				}
				testDataRC.testData = append(
					testDataRC.testData,
					dockerLogDataEntry...)

				logEntriesIncluded++
			}
			container["msgCount"] = logEntriesIncluded

			return &testDataRC, nil

		}
	}
	return nil, fmt.Errorf("MOCK dClient: Can't find stream for container '%s'", contID)
}

func (c *mockDockerClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	containers := []types.Container{}
	for _, elem := range c.targetContainers {
		containerName, _ := elem["name"].(string)

		containers = append(containers,
			types.Container{ID: elem["id"].(string),
				Names: []string{containerName}})
	}
	return containers, nil
}

func (c *mockDockerClient) Close() error {
	return nil
}

func (c *mockDockerClient) getContainerByID(contID string) map[string]interface{} {
	for _, elem := range c.targetContainers {
		if elem["id"] == contID || trimId(elem["id"].(string)) == contID {
			return elem
		}
	}
	return nil
}

type mockReaderCloser struct {
	testData []byte
}

func (r *mockReaderCloser) eof() bool {
	return len(r.testData) == 0
}

func (r *mockReaderCloser) readByte() byte {
	// this function assumes that eof() check was done before
	b := r.testData[0]
	r.testData = r.testData[1:]
	return b
}

func (r *mockReaderCloser) Read(p []byte) (n int, err error) {
	if r.eof() {
		err = io.EOF
		return
	}

	if c := len(p); c > 0 {
		for n < c {
			p[n] = r.readByte()
			n++
			if r.eof() {
				break
			}
		}
	}
	return
}

func (r *mockReaderCloser) Close() (err error) {
	return nil
}

func parseDataFromRawLogEntry(rawLogEntry map[string]interface{}) (time.Time, string, error) {
	var ts time.Time
	var streamTag string
	var err error
	ts, err = time.Parse(time.RFC3339Nano, rawLogEntry["tss"].(string))
	if rawLogEntry["header"] == nil {
		streamTag = "tty"
	} else {
		switch rawLogEntry["header"].([]byte)[0] {
		case stdInHeader[0]:
			{
				streamTag = "stdin"
			}
		case stdOutHeader[0]:
			{
				streamTag = "stdout"
			}
		case stdErrHeader[0]:
			{
				streamTag = "stderr"
			}
		default:
			err = fmt.Errorf("Corrupted header in log entry. Can be one of"+
				" noHeader/stdInHeader/stdOutHeader/stdErrHeader, got: %s", rawLogEntry["header"])
		}
	}
	return ts, streamTag, err
}

func genericTest(t *testing.T, input *DockerLogs, waitStreamers time.Duration) {
	var acc testutil.Accumulator
	var err error
	var lastUnixTS = map[string]int64{}
	var msgConut = map[string]int{}
	var msgIndex = map[string][]int{}

	acc.SetDebug(true)

	err = input.Init()
	require.NoError(t, err)

	err = input.Gather(&acc)
	require.NoError(t, err)

	if waitStreamers == 0 { //wait until streamers done
		//Waiting until docker streamer receive EOF or canceled
		input.wgStreamers.Wait()

	} else {
		time.Sleep(waitStreamers)
	}

	input.Stop()
	input.wg.Wait()

	for _, metric := range acc.Metrics {
		containerElement := input.client.(*mockDockerClient).getContainerByID(metric.Fields["container_id"].(string))
		require.NotNil(t, containerElement, "Can't detect container by accumulator metric tag 'container_id'.\n"+
			"Check if tag populated and has proper value")

		msgSplitArray := strings.Split(metric.Fields["message"].(string), ":")
		if len(msgSplitArray) < 1 { //TODO: replace to require.Compare
			panic("Corrupted mock container log entries, each string should start with index of message in the raw log entries array.\n" +
				"Pattern: '<INDEX>:<MESSAGE>'")
		}
		rawLogEntriesIndex, err := strconv.Atoi(msgSplitArray[0])
		require.Nil(t, err, "Can't parse message index from the mock container message, is index an integer?...")

		rawLogEntry := containerElement["rawLogEntries"].([]map[string]interface{})[rawLogEntriesIndex]
		ts, streamTag, err := parseDataFromRawLogEntry(rawLogEntry)
		require.Nil(t, err, "Can't parse ts/streamTag from raw log entry")

		assert.Equal(t, "docker_log", metric.Measurement)

		assert.Equal(t,
			strings.TrimRightFunc(rawLogEntry["data"].(string), unicode.IsSpace),
			metric.Fields["message"].(string))

		if !input.disableTimeStampsStreaming {
			assert.Equal(t,
				ts,
				metric.Time)
		}

		assert.Equal(t,
			streamTag,
			metric.Tags["stream"])

		msgConut[metric.Fields["container_id"].(string)]++
		msgIndex[metric.Fields["container_id"].(string)] = append(msgIndex[metric.Fields["container_id"].(string)], []int{rawLogEntriesIndex}...)
		lastUnixTS[metric.Fields["container_id"].(string)] = metric.Time.UnixNano() + 1

	}

	//Check filtering of container log entries based on TS (in case all messages are received)
	if waitStreamers == 0 {
		for contID, msgCount := range msgConut {
			containerElement := input.client.(*mockDockerClient).getContainerByID(contID)
			assert.Equal(t, msgCount, containerElement["msgCount"], fmt.Sprintf("Container ID '%s'", contID))
		}
	}

	//Checking offset flusher
	for contID, lastTS := range lastUnixTS {
		_, tsFromOffsetFile := input.getOffset(path.Join(input.OffsetStoragePath, contID))
		assert.Equal(t, lastTS, tsFromOffsetFile, fmt.Sprintf("Container ID '%s'", contID))
	}
}

//Test log delivery with different headers and TS
func TestTS(t *testing.T) { //Mixed containers with time stamps

	input := DockerLogs{
		client:                         newMockDockerClient(targetContainers),
		OffsetStoragePath:              "./collector_offset",
		InitialChunkSize:               20, //to split the first string in 2 parts
		MaxChunkSize:                   80,
		StaticContainerList:            targetContainers,
		OffsetFlush:                    internal.Duration{Duration: defaultFlushInterval},
		LogGatherInterval:              internal.Duration{Duration: defaultLogGatherInterval},
		Timeout:                        internal.Duration{Duration: defaultAPICallTimeout},
		containerList:                  make(map[string]context.CancelFunc),
		processedContainerList:         make(map[string]interface{}),
		processedContainersCheckerDone: make(chan bool),
		processedContainersChan:        make(chan map[string]interface{}),
		offsetData:                     make(chan offsetData),
		offsetDone:                     make(chan bool)}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
	genericTest(t, &input, 0)
}

//Test filtering messages from container based on offset
func TestTSOffset(t *testing.T) {

	input := DockerLogs{
		client:                         newMockDockerClient(targetContainers),
		OffsetStoragePath:              "./collector_offset",
		InitialChunkSize:               20, //to split the first string in 2 parts
		MaxChunkSize:                   80,
		StaticContainerList:            targetContainers,
		OffsetFlush:                    internal.Duration{Duration: defaultFlushInterval},
		LogGatherInterval:              internal.Duration{Duration: defaultLogGatherInterval},
		Timeout:                        internal.Duration{Duration: defaultAPICallTimeout},
		containerList:                  make(map[string]context.CancelFunc),
		processedContainerList:         make(map[string]interface{}),
		processedContainersCheckerDone: make(chan bool),
		processedContainersChan:        make(chan map[string]interface{}),
		offsetData:                     make(chan offsetData),
		offsetDone:                     make(chan bool)}

	//Generating TS files:
	//Create storage path
	if src, err := os.Stat(input.OffsetStoragePath); os.IsNotExist(err) {
		errDir := os.MkdirAll(input.OffsetStoragePath, 0755)
		if errDir != nil {
			require.Nil(t, errDir, fmt.Sprintf("Can't create directory '%s' to store offset, reason: %s", input.OffsetStoragePath, errDir.Error()))
		}
	} else if src != nil && src.Mode().IsRegular() {
		require.Equal(t, false, src.Mode().IsRegular(), fmt.Sprintf("'%s' already exist as a file!", input.OffsetStoragePath))
	}

	for _, container := range input.client.(*mockDockerClient).targetContainers {
		//get ts from the lst log entry in container's rawLogEntries
		rawLogEntries := container["rawLogEntries"].([]map[string]interface{})
		entryTS, err := time.Parse(time.RFC3339Nano, rawLogEntries[len(rawLogEntries)-1]["tss"].(string))
		require.Nil(t, err, fmt.Sprintf("Container id '%s', can't parse timestamp from log entry: %s, ts: %s",
			container["id"],
			rawLogEntries[len(rawLogEntries)-1]["data"].(string),
			rawLogEntries[len(rawLogEntries)-1]["tss"].(string)))

		filename := path.Join(input.OffsetStoragePath, container["id"].(string))
		offset := []byte(strconv.FormatInt(entryTS.Unix(), 10))
		err = ioutil.WriteFile(filename, offset, 0777)
		require.Nil(t, err, fmt.Sprintf("Can't write lg offset to file '%s', reason: %v",
			filename, err))

	}
	genericTest(t, &input, 0)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}

//Test streaming of logs without TS
func TestWOTS(t *testing.T) { //Mixed containers without time stamps

	input := DockerLogs{
		client:                         newMockDockerClient(targetContainersWOTS),
		OffsetStoragePath:              "./collector_offset",
		InitialChunkSize:               20, //to split the first string in 2 parts
		MaxChunkSize:                   80,
		disableTimeStampsStreaming:     true,
		StaticContainerList:            targetContainersWOTS,
		OffsetFlush:                    internal.Duration{Duration: defaultFlushInterval},
		LogGatherInterval:              internal.Duration{Duration: defaultLogGatherInterval},
		Timeout:                        internal.Duration{Duration: defaultAPICallTimeout},
		containerList:                  make(map[string]context.CancelFunc),
		processedContainerList:         make(map[string]interface{}),
		processedContainersCheckerDone: make(chan bool),
		processedContainersChan:        make(chan map[string]interface{}),
		offsetData:                     make(chan offsetData),
		offsetDone:                     make(chan bool)}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	genericTest(t, &input, 0)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}

//Test race condition while interrupt receiving logs from containers...
//Check for possible deadlocks while stopping srteam readers, offset flusher, etc.
func TestRaceCondition(t *testing.T) {

	input := DockerLogs{
		client:                         newMockDockerClient(targetContainersWOTS),
		OffsetStoragePath:              "./collector_offset",
		InitialChunkSize:               20, //to split the first string in 2 parts
		MaxChunkSize:                   80,
		disableTimeStampsStreaming:     true,
		StaticContainerList:            targetContainersWOTS,
		OffsetFlush:                    internal.Duration{Duration: defaultFlushInterval},
		LogGatherInterval:              internal.Duration{Duration: defaultLogGatherInterval},
		Timeout:                        internal.Duration{Duration: defaultAPICallTimeout},
		containerList:                  make(map[string]context.CancelFunc),
		processedContainerList:         make(map[string]interface{}),
		processedContainersCheckerDone: make(chan bool),
		processedContainersChan:        make(chan map[string]interface{}),
		offsetData:                     make(chan offsetData),
		offsetDone:                     make(chan bool)}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	genericTest(t, &input, time.Millisecond*100)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}

//Test ShutDownWhenEOF...
//Check for possible deadlocks while checking the eof and shutdowning....
/*
func TestShutdownWhenEOF(t *testing.T) {

	input := DockerLogs{
		client: &mockDockerClient{
			containerInspectData: containerInspectRunning,
			targetContainers:     targetContainersWOTS},
		Endpoint:                   "MOCK",
		ShutDownWhenEOF:            true,
		OffsetStoragePath:          "./collector_offset",
		InitialChunkSize:           20, //to split the first string in 2 parts
		MaxChunkSize:               80,
		disableTimeStampsStreaming: true,
		StaticContainerList:           targetContainersWOTS}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	var acc testutil.Accumulator
	var err error
	var lastUnixTS = map[string]int64{}
	var msgConut = map[string]int{}
	var msgIndex = map[string][]int{}

	acc.SetDebug(true)

	err = input.Start(&acc)
	require.NoError(t, err)

	input.wgStreamers.Wait()

	for _, metric := range acc.Metrics {
		containerElement := input.client.(*mockDockerClient).getContainerByID(metric.Tags["container_id"])
		require.NotNil(t, containerElement, "Can't detect container by accumulator metric tag 'container_id'.\n"+
			"Check if tag populated and has proper value")

		msgSplitArray := strings.Split(metric.Fields["value"].(string), ":")
		if len(msgSplitArray) < 1 { //TODO: replace to require.Compare
			panic("Corrupted mock container log entries, each string should start with index of message in the raw log entries array.\n" +
				"Pattern: '<INDEX>:<MESSAGE>'")
		}
		rawLogEntriesIndex, err := strconv.Atoi(msgSplitArray[0])
		require.Nil(t, err, "Can't parse message index from the mock container message, is index an integer?...")

		rawLogEntry := containerElement["rawLogEntries"].([]map[string]interface{})[rawLogEntriesIndex]
		ts, streamTag, err := parseDataFromRawLogEntry(rawLogEntry)
		require.Nil(t, err, "Can't parse ts/streamTag from raw log entry")

		assert.Equal(t, "stream", metric.Measurement)

		assert.Equal(t,
			rawLogEntry["data"],
			metric.Fields["value"])

		if !input.disableTimeStampsStreaming {
			assert.Equal(t,
				ts,
				metric.Time)
		}

		assert.Equal(t,
			streamTag,
			metric.Tags["stream"])

		msgConut[metric.Tags["container_id"]]++
		msgIndex[metric.Tags["container_id"]] = append(msgIndex[metric.Tags["container_id"]], []int{rawLogEntriesIndex}...)
		lastUnixTS[metric.Tags["container_id"]] = metric.Time.UnixNano() + 1

	}

	//Check filtering of container log entries based on TS (in case all messages are received)
	for contID, msgCount := range msgConut {
		containerElement := input.client.(*mockDockerClient).getContainerByID(contID)
		assert.Equal(t, msgCount, containerElement["msgCount"], fmt.Sprintf("Container ID '%s'", contID))
	}

	input.wg.Wait()
	//Wait while input.Stop is called by telegraf itslef and all the stuff flushed
	time.Sleep(3 * time.Second)

	//Checking offset flusher
		for contID, lastTS := range lastUnixTS {
			_, tsFromOffsetFile := getOffset(path.Join(input.OffsetStoragePath, contID))
			assert.Equal(t, lastTS, tsFromOffsetFile, fmt.Sprintf("Container ID '%s'", contID))
		}
	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}
*/
