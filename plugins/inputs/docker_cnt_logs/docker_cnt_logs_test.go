package docker_cnt_logs

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockDockerClient struct {
	containerInspectData types.ContainerJSON
	targetContainers     []map[string]interface{}
}

func (c *mockDockerClient) ContainerInspect(ctx context.Context, contID string) (types.ContainerJSON, error) {
	return c.containerInspectData, nil
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
						return nil, fmt.Errorf("MOCK Client: Can't parse timestamp from docker options. Cont ID '%s', reason: %v", contID, err)
					} else {
						sinceTS = sinceTSTime.UnixNano()
					}
				}

				log.Printf("D! [inputs.docker_cnt_logs (mock)] Container '%s', stream logs since: %s (%d)",
					contID, time.Unix(sinceTS, 0).Format(time.RFC3339Nano), sinceTS)
			}
			rawLogEntries := container["rawLogEntries"].([]map[string]interface{})
			logEntriesIncluded := 0
			for _, entry := range rawLogEntries {

				//Filtering log entries based on TS
				entryTS, err = time.Parse(time.RFC3339Nano, entry["tss"].(string))
				if err != nil {
					log.Printf("D! [inputs.docker_cnt_logs (mock)] Container '%s', stream log entry '%s' can't parse ts '%s'\n",
						contID, entry["data"].(string), entry["tss"].(string))
					continue
				}
				if entryTS.Unix() < sinceTS {
					log.Printf("D! [inputs.docker_cnt_logs (mock)] Container '%s', stream log entry '%s' filtered\n"+
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
	return nil, fmt.Errorf("MOCK Client: Can't find stream for container '%s'", contID)
}

func (c *mockDockerClient) Close() error {
	return nil
}

func (c *mockDockerClient) getContainerByID(contID string) map[string]interface{} {
	for _, elem := range c.targetContainers {
		if elem["id"] == contID {
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

var int64Dummy int64

var containerInspectRunning = types.ContainerJSON{
	ContainerJSONBase: &types.ContainerJSONBase{
		ID:              "dummyContainer",
		Created:         "",
		Path:            "",
		Args:            []string{""},
		State:           &types.ContainerState{Status: "running"},
		Image:           "",
		ResolvConfPath:  "",
		HostnamePath:    "",
		HostsPath:       "",
		LogPath:         "",
		Node:            &types.ContainerNode{},
		Name:            "",
		RestartCount:    0,
		Driver:          "",
		Platform:        "",
		MountLabel:      "",
		ProcessLabel:    "",
		AppArmorProfile: "",
		ExecIDs:         []string{""},
		HostConfig:      &container.HostConfig{},
		GraphDriver:     types.GraphDriverData{},
		SizeRw:          &int64Dummy,
		SizeRootFs:      &int64Dummy},
	Mounts:          []types.MountPoint{},
	Config:          &container.Config{},
	NetworkSettings: &types.NetworkSettings{}}

var stdOutHeader = []byte{1, 0, 0, 0, 0, 0, 1, 0}
var stdErrHeader = []byte{2, 0, 0, 0, 0, 0, 1, 0}
var stdInHeader = []byte{0, 0, 0, 0, 0, 0, 1, 0}

var testLogsOutputData = []map[string]interface{}{
	{
		"tss":    "2019-10-29T11:15:14.813957700Z",
		"data":   "0:first message\n",
		"header": stdOutHeader,
	},
	{
		"tss":    "2019-10-29T11:15:15.813957700Z",
		"data":   "1:intermediate message\n",
		"header": stdErrHeader,
	},
	{
		"tss":    "2019-10-29T11:15:17.813957700Z",
		"data":   "2:last message",
		"header": stdInHeader,
	}}
var testLogsOutputDataWOHeaders = []map[string]interface{}{
	{
		"tss":    "2019-10-29T11:15:13.813957700Z",
		"data":   "0:first message\n",
		"header": nil,
	},
	{
		"tss":    "2019-10-29T11:15:18.813957700Z",
		"data":   "1:last message",
		"header": nil,
	}}

func parseDataFromRawLogEntry(rawLogEntry map[string]interface{}) (time.Time, string, error) {
	var ts time.Time
	var streamTag string
	var err error
	ts, err = time.Parse(time.RFC3339Nano, rawLogEntry["tss"].(string))
	if rawLogEntry["header"] == nil {
		streamTag = "interactive"
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

var targetContainers = []map[string]interface{}{
	{
		"id":            "dummyContainer1",
		"interval":      "500ms",
		"rawLogEntries": testLogsOutputData,
		"tags":          []interface{}{"tag1=StreamWithHeaders", "tag2=value2"}},
	{
		"id":            "dummyContainer2",
		"interval":      "1000ms",
		"rawLogEntries": testLogsOutputDataWOHeaders,
		"tags":          []interface{}{"tag1=StreamWithoutHeaders", "tag2=value2"}}}

var targetContainersWOTS = []map[string]interface{}{
	{
		"id":            "dummyContainer1",
		"interval":      "500ms",
		"rawLogEntries": testLogsOutputDataWOHeaders,
		"tags":          []interface{}{"tag1=StreamWithoutHeaders", "tag2=value2"}},
	{
		"id":            "dummyContainer2",
		"interval":      "1000ms",
		"rawLogEntries": testLogsOutputData,
		"tags":          []interface{}{"tag1=StreamWithHeaders", "tag2=value2"}}}

func genericTest(t *testing.T, input *DockerCNTLogs, waitEof time.Duration) {
	var acc testutil.Accumulator
	var err error
	var lastUnixTS = map[string]int64{}
	var msgConut = map[string]int{}
	var msgIndex = map[string][]int{}

	acc.SetDebug(true)

	err = input.Start(&acc)
	require.NoError(t, err)
	if waitEof == 0 { //wait until EOF
		//Waiting until docker stream receive EOF
		closed := 0
		for {
			closed = 0
			for _, logReader := range input.logReader {

				logReader.lock.Lock()
				if logReader.eofReceived {
					closed++
				}
				logReader.lock.Unlock()
			}
			if closed == len(input.logReader) {
				break
			}
		}
	} else {
		time.Sleep(waitEof)
	}

	input.Stop()
	input.wg.Wait()

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
	if waitEof == 0 {
		for contID, msgCount := range msgConut {
			containerElement := input.client.(*mockDockerClient).getContainerByID(contID)
			assert.Equal(t, msgCount, containerElement["msgCount"], fmt.Sprintf("Container ID '%s'", contID))
		}
	}

	//Checking offset flusher
	for contID, lastTS := range lastUnixTS {
		_, tsFromOffsetFile := getOffset(path.Join(input.OffsetStoragePath, contID))
		assert.Equal(t, lastTS, tsFromOffsetFile, fmt.Sprintf("Container ID '%s'", contID))
	}
}

//Test log delivery with different headers and TS
func TestTS(t *testing.T) { //Mixed containers with time stamps

	input := DockerCNTLogs{
		client: &mockDockerClient{
			containerInspectData: containerInspectRunning,
			targetContainers:     targetContainers},
		Endpoint:          "MOCK",
		ShutDownWhenEOF:   false,
		OffsetStoragePath: "./collector_offset",
		InitialChunkSize:  20, //to split the first string in 2 parts
		MaxChunkSize:      80,
		TargetContainers:  targetContainers}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	genericTest(t, &input, 0)
}

//Test filtering messages from container based on offset
func TestTSOffset(t *testing.T) {

	input := DockerCNTLogs{
		client: &mockDockerClient{
			containerInspectData: containerInspectRunning,
			targetContainers:     targetContainers},
		Endpoint:          "MOCK",
		ShutDownWhenEOF:   false,
		OffsetStoragePath: "./collector_offset",
		InitialChunkSize:  20, //to split the first string in 2 parts
		MaxChunkSize:      80,
		TargetContainers:  targetContainers}

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
		require.Nil(t, err, fmt.Sprintf("Can't write logger offset to file '%s', reason: %v",
			filename, err))

	}
	genericTest(t, &input, 0)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}

//Test streaming of logs without TS
func TestWOTS(t *testing.T) { //Mixed containers without time stamps

	input := DockerCNTLogs{
		client: &mockDockerClient{
			containerInspectData: containerInspectRunning,
			targetContainers:     targetContainersWOTS},
		Endpoint:                   "MOCK",
		ShutDownWhenEOF:            false,
		OffsetStoragePath:          "./collector_offset",
		InitialChunkSize:           20, //to split the first string in 2 parts
		MaxChunkSize:               80,
		disableTimeStampsStreaming: true,
		TargetContainers:           targetContainersWOTS}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	genericTest(t, &input, 0)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}

//Test race condition while interrupt receiving logs from containers...
//Check for possible deadlocks while stopping srteam readers, offset flusher, etc.
func TestRaceCondition(t *testing.T) {

	input := DockerCNTLogs{
		client: &mockDockerClient{
			containerInspectData: containerInspectRunning,
			targetContainers:     targetContainersWOTS},
		Endpoint:                   "MOCK",
		ShutDownWhenEOF:            false,
		OffsetStoragePath:          "./collector_offset",
		InitialChunkSize:           20, //to split the first string in 2 parts
		MaxChunkSize:               80,
		disableTimeStampsStreaming: true,
		TargetContainers:           targetContainersWOTS}

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))

	genericTest(t, &input, time.Millisecond*500)

	//Removing offset files
	require.Nil(t, os.RemoveAll(input.OffsetStoragePath))
}
