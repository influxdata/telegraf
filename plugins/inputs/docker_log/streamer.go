package docker_log

import (
	"context"
	"fmt"
	"github.com/influxdata/telegraf/internal/docker"
	"io"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/influxdata/telegraf"
	"github.com/pkg/errors"
)

type streamer struct {
	contID     string
	acc        telegraf.Accumulator
	contStream io.ReadCloser

	dockerTimeStamps    bool
	headers             bool
	interval            time.Duration
	initialChunkSize    int
	currentChunkSize    int
	maxChunkSize        int
	outputMsgStartIndex int
	buffer              []byte
	leftoverBuffer      []byte
	length              int
	endOfLineIndex      int
	tags                map[string]string

	currentOffset int64
	offsetData    chan offsetData
}

func newStreamer(ctxStream context.Context, container types.Container, contStatus types.ContainerJSON, acc telegraf.Accumulator, d *DockerLogs) (*streamer, error) {
	var err error
	var getLogsSince string
	var currentOffset int64
	var fromBeginning bool
	var logRC io.ReadCloser
	var outputMsgStartIndex int
	var logGatherInterval time.Duration
	var initialChunkSize int
	var maxChunkSize int
	var tags = map[string]string{"container_name": contStatus.Name}

	fromBeginning = d.FromBeginning
	logGatherInterval = d.LogGatherInterval.Duration
	initialChunkSize = d.InitialChunkSize
	maxChunkSize = d.MaxChunkSize

	//Update streaming options respecting individual settings that can come from static container list:
	if staticContainer := d.getContainerFromStaticList(contStatus.ID); staticContainer != nil {
		//From beginning
		if fb, ok := staticContainer["from_beginning"].(bool); ok {
			fromBeginning = fb
		}

		//Interval
		interval, ok := staticContainer["log_gather_interval"].(string)
		if parsedInterval, err := time.ParseDuration(interval); ok && err == nil {
			logGatherInterval = parsedInterval
		} else if err != nil && ok {
			return nil, errors.Errorf("Can't parse interval from string '%s', reason: %s", staticContainer["interval"].(string), err.Error())
		}

		//initial chunk size
		if initialChSz, ok := staticContainer["initial_chunk_size"].(int); ok && initialChSz <= dockerLogHeaderSize {
			initialChunkSize = 2 * dockerLogHeaderSize
		} else if initialChSz > dockerLogHeaderSize {
			initialChunkSize = initialChSz
		}

		//max chunk size
		if maxChSz, ok := staticContainer["max_chunk_size"].(int); ok && maxChSz <= initialChunkSize {
			maxChunkSize = 5 * initialChunkSize
		} else if maxChSz > initialChunkSize {
			maxChunkSize = maxChSz
		}

		if _, ok := staticContainer["tags"]; ok { //tags specified
			for _, tag := range staticContainer["tags"].([]interface{}) {
				arr := strings.Split(tag.(string), "=")
				if len(arr) != 2 {
					lg.logE("Static container: '%s'.\n"+
						"Can't parse tags from string '%s', valid format is <tag_name>=<tag_value>",
						trimId(contStatus.ID),
						tag.(string))
					continue
				}
				tags[arr[0]] = arr[1]
			}
		}

	}

	//Detecting headers:
	if !contStatus.Config.Tty { //Non tty containers mutliplex stdout, stderr into single channel
		outputMsgStartIndex = dockerLogHeaderSize //Header is in the string, need to strip it out...
		lg.logD("Container '%s', logs are streamed WITH headers!", trimId(contStatus.ID))
	} else {
		outputMsgStartIndex = 0 //No header in the output, start from the 1st letter.
		lg.logD("Container '%s', logs are streamed WITHOUT headers!", trimId(contStatus.ID))
	}

	//Detecting offset
	getLogsSince, currentOffset = d.getOffset(path.Join(d.OffsetStoragePath, contStatus.ID))

	if getLogsSince != "" && fromBeginning { // If there is an offset, then it means that we already deliver logs until the offset
		//In this case we can ignore 'fromBeginning' ,and continue from the offset
		lg.logD("Container '%s', 'from_beginning' option ignored, since there is an offset: %s", trimId(contStatus.ID), getLogsSince)

	} else if getLogsSince == "" && !fromBeginning { //No offset and not from beginning, means that we ship logs since now.
		getLogsSince = time.Now().UTC().Format(time.RFC3339Nano)
	}

	//Setup logs reader
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: !d.disableTimeStampsStreaming,
		Since:      getLogsSince}

	logRC, err = d.client.ContainerLogs(ctxStream, contStatus.ID, options)
	if err != nil {
		return nil, errors.Errorf("Container '%s', can't get log ReadCloser from Docker API, reason:\n%s", trimId(contStatus.ID), err.Error())
	}

	tags["container_image"], tags["container_version"] = docker.ParseImage(contStatus.Config.Image)
	if d.IncludeSourceTag {
		tags["source"] = trimId(contStatus.ID)
	}

	// Add matching container labels as tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	return &streamer{
			acc:                 acc,
			contID:              contStatus.ID,
			offsetData:          d.offsetData,
			interval:            logGatherInterval,
			dockerTimeStamps:    !d.disableTimeStampsStreaming, //Default behaviour - stream logs with time-stamps
			headers:             !contStatus.Config.Tty,
			initialChunkSize:    initialChunkSize,
			maxChunkSize:        maxChunkSize,
			currentChunkSize:    initialChunkSize,
			outputMsgStartIndex: outputMsgStartIndex,
			contStream:          logRC,
			buffer:              make([]byte, initialChunkSize),
			currentOffset:       currentOffset,
			tags:                tags},
		nil
}

//If there is no new line in interval [eolIndex-HeaderSize,eolIndex+HeaderSize],
//then we are definitely not in the middle of header, otherwise, we are.
func (s *streamer) isNewLineInMsgHeader(eolIndex int) bool {
	str := &s.buffer
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

func (s *streamer) stream() error {

	var err error
	var eofReceived = false
	var contextCancelled = false
	defer s.contStream.Close()

	for {
		//Iterative reads by chunks
		// While reading in chunks, there are 2 general cases:
		// 1. Either full buffer (it means that the message either fit to chunkSize or exceed it.
		// To figure out if it exceed we need to check if the buffer ends with "\r\n"

		// 2. Or partially filled buffer. In this case the rest of the buffer is '\0'

		// Read from docker API
		//Can be a case when API returns length==0, and err==nil
		s.length, err = s.contStream.Read(s.buffer)

		switch err {
		case nil: //No error, proceed!
		case io.EOF:
			eofReceived = true
			lg.logW("Container '%s', EOF received.", trimId(s.contID))
		case context.Canceled:
			contextCancelled = true
			lg.logW("Container '%s', streamer shutdown requested.", trimId(s.contID))
		default:
			return fmt.Errorf("Read error from container '%s': %v", trimId(s.contID), err)
		}

		if needMoreData := s.prepareDataForSending(); needMoreData {
			continue
		}

		//Check if there smth. to be sent and send.
		if len(s.buffer) > 0 && s.endOfLineIndex > 0 {
			s.sendData()
		}

		//Send offset to flusher (in a non blocking manner)
		select {
		case s.offsetData <- offsetData{s.contID, s.currentOffset}:
		default:
		}

		//Control the size of buffer. Buffer can grow because of:
		//1. Appending leftover (most often case)
		//2. Increase in chunksize
		if len(s.buffer) > s.currentChunkSize+s.currentChunkSize/2 ||
			len(s.buffer) > s.maxChunkSize {
			s.buffer = nil
			s.buffer = make([]byte, s.currentChunkSize)
		}

		if eofReceived || contextCancelled {
			return err
		}

		time.Sleep(s.interval)
	}

}

func (s *streamer) prepareDataForSending() (readMore bool) {

	// s.length stores count of symbols returned by last call to s.contStream.Read(s.buffer)
	// if s.length==0 - no data was returned.
	newDataFromStreamExist := s.length != 0

	//Append leftover from previous iteration (if it exists)
	if len(s.leftoverBuffer) > 0 {
		s.buffer = append(s.leftoverBuffer, s.buffer...)
		s.length += len(s.leftoverBuffer)

		//Erasing leftover buffer once used:
		s.leftoverBuffer = nil
	}

	//Now we have buffer to work with
	//The purpose of the code below is to find position (index) where the last complete line ends (last '\n' symbol).
	//If buffer represents unterminated line, then we move all the buffer to leftover buffer, and request more data from s.contStream
	//The edge case here is when s.contStream returns some data and EOF, in this case buffer won't be terminated with \n

	//Setting s.endOfLineIndex to the last position in buffer (as we perform search of '\n' from the end of the buffer)
	s.endOfLineIndex = s.length - 1

	if !newDataFromStreamExist { //No new data to process
		return false
	}

	//Seek last '\n' (from the end), ignoring the edge case when '\n' is in the message header
	for ; s.endOfLineIndex >= s.outputMsgStartIndex; s.endOfLineIndex-- {
		if s.buffer[s.endOfLineIndex] == '\n' {

			//Skip '\n' if there are headers and '\n' is inside header (edge case when stream has headers)
			if s.headers && s.isNewLineInMsgHeader(s.endOfLineIndex) {
				continue
			}

			if s.endOfLineIndex != s.length-1 { //Something is in buffer after last '\n'
				// Moving this to leftover buffer
				s.leftoverBuffer = nil
				s.leftoverBuffer = make([]byte, (s.length-1)-s.endOfLineIndex)
				copy(s.leftoverBuffer, s.buffer[s.endOfLineIndex+1:])
			}
			return false
		}
	}

	//End of line is not found! We either got very long message or buffer is really small
	//We need to:
	//1. move it into leftover buffer
	//2. grow current chunk size if limit is not exceeded
	s.leftoverBuffer = nil
	s.leftoverBuffer = make([]byte, s.length)
	copy(s.leftoverBuffer, s.buffer[:s.length])

	//Grow chunk size
	if s.currentChunkSize*2 <= s.maxChunkSize {
		s.currentChunkSize = s.currentChunkSize * 2
		s.buffer = nil
		s.buffer = make([]byte, s.currentChunkSize)
	}

	return true //Means that we have incomplete portion of data

}

func (s *streamer) sendData() {
	var totalLineLength int
	var timeStamp time.Time
	var field = map[string]interface{}{"container_id": s.contID}
	var tags = s.tags
	var err error

	tags["stream"] = "tty" //default stream tag

	//Parsing the buffer line by line and send data to accumulator
	for i := 0; i <= s.endOfLineIndex; i = i + totalLineLength {
		field["message"] = ""

		if i+s.outputMsgStartIndex < s.endOfLineIndex { //Checking boundaries, if the data consistent
			//Looking for the end of the line (skipping header, as header can contain '\n')
			totalLineLength = 0
			for j := i + s.outputMsgStartIndex; j <= s.endOfLineIndex; j++ {
				if s.buffer[j] == '\n' {
					totalLineLength = j - i + 1 //Include '\n'
					break
				}
			}
			if totalLineLength == 0 {
				totalLineLength = (s.endOfLineIndex + 1) - i
			}

			//Getting stream type (if header persist)
			if s.headers {
				switch s.buffer[i] {
				case 0x1:
					tags["stream"] = "stdout"
				case 0x2:
					tags["stream"] = "stderr"
				case 0x0:
					tags["stream"] = "stdin"
				}
			}

			if !s.dockerTimeStamps { //no time stamp
				timeStamp = time.Now()
				field["message"] = string(s.buffer[i+s.outputMsgStartIndex : i+totalLineLength])
			} else {

				timeStamp, err = time.Parse(time.RFC3339Nano,
					string(s.buffer[i+s.outputMsgStartIndex:i+s.outputMsgStartIndex+dockerTimeStampLength]))

				if err != nil {
					s.acc.AddError(fmt.Errorf("Can't parse time stamp from string, container '%s': "+
						"%v. Raw message string:\n%s\nOutput msg start index: %d",
						trimId(s.contID), err, s.buffer[i:i+totalLineLength], s.outputMsgStartIndex))
					//lg.logE("\n=========== buffer[:lr.endOfLineIndex] ===========\n"+
					//	"%s\n=========== ====== ===========\n", s.buffer[:s.endOfLineIndex])
					timeStamp = time.Now()
				}
				field["message"] = string(s.buffer[i+s.outputMsgStartIndex+dockerTimeStampLength+1 : i+totalLineLength])
			}

			//Saving offset
			s.currentOffset += timeStamp.UTC().UnixNano() - s.currentOffset + 1 //+1 (nanosecond) here prevents to include current message

		} else { //sort of garbage
			timeStamp = time.Now()
			field["message"] = string(s.buffer[i:s.endOfLineIndex])
		}

		// Keep any leading space, but remove whitespace from end of line.
		// This preserves space in, for example, stacktraces, while removing
		// annoying end of line characters and is similar to how other logging
		// plugins such as syslog behave.
		field["message"] = strings.TrimRightFunc(field["message"].(string), unicode.IsSpace)
		s.acc.AddFields(title, field, tags, timeStamp)
	}
}
