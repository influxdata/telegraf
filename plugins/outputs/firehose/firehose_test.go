package firehose

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	//"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	//uuid "github.com/satori/go.uuid"
)

// primary key spoof
var recordID int64 = 0

// mockFirehose extends the firehose interface API for testing
//
// The type adds a test interface from the testing package as well as
// values to be used during test runs and subsequent validation
type mockFirehose struct {
	firehoseiface.FirehoseAPI

	// a link to the test interface to quickly fail tests
	t *testing.T

	// test values
	expectedLines      int64
	numErrorsRemaining int64

	// reaction values
	numOfPuts int64 // tracks the number of times PutRecordBatch was called
}

// PutRecordBatch is an override of a function in the firehose package of the
// AWS SDK in order to circumvent utilizing the actual API endpoints during
// testing
func (m *mockFirehose) PutRecordBatch(input *firehose.PutRecordBatchInput) (output *firehose.PutRecordBatchOutput, err error) {
	if len(input.Records) > 500 {
		m.t.Log("E! firehose_test: got more than 500 in a batch")
		m.t.Fail()
	}

	code := "42"
	message := "Deliberate Error!"
	errCount := int64(0)
	numInputRecords := int64(len(input.Records))

	// insert errors
	var entry firehose.PutRecordBatchResponseEntry
	var responses []*firehose.PutRecordBatchResponseEntry

	// if we have more errors to insert than total number of records
	// simulate a total failure of the write
	if numInputRecords <= m.numErrorsRemaining {
		// simulated error, eaten by writeToFirehose (calling function)
		err = errors.New("Total Failure Simulated")
		errCount = errCount + numInputRecords
		m.numErrorsRemaining = m.numErrorsRemaining - numInputRecords

		// create responses for output and process errors
	} else {
		m.numOfPuts++

		for index, _ := range input.Records {
			recordID++
			if int64(index) < m.numErrorsRemaining {
				// inserting errors at first for specified error count
				entry = firehose.PutRecordBatchResponseEntry{
					ErrorCode:    &code,
					ErrorMessage: &message}
				// once the error is inserted, decrement the error count remaining
				m.numErrorsRemaining = m.numErrorsRemaining - 1
				errCount++
			} else {
				idString := strconv.FormatInt(recordID, 10)
				entry = firehose.PutRecordBatchResponseEntry{RecordId: &idString}
			}
			responses = append(responses, &entry)
		}
	}

	batchOutput := firehose.PutRecordBatchOutput{FailedPutCount: &errCount, RequestResponses: responses}
	return &batchOutput, err
}

// generateLines simply uses the built-in telegraf testing functions to
// create some metrics in the correct format and structure.  We did this
// little wrapper to quickly spin up any number of metrics.
func generateLines(numLines int64) (lines []telegraf.Metric, err error) {
	err = nil
	lines = testutil.MockMetrics()

	// generate 1 less line then specified since
	// the MockMetrics line returns a line when generated
	for i := int64(0); i < (numLines - 1); i++ {
		lines = append(lines, testutil.TestMetric(1.0))
	}
	return
}

// attachSerializer quickly creates a new serializer for handling
// metrics from telegraf. Nothing fancy here, just a new serializer
// and added it into our firehose output struct.
func attachSerializer(t *testing.T, f *FirehoseOutput) (err error) {
	s, err := serializers.NewInfluxSerializer()
	if err != nil {
		t.Fail()
		return
	}
	f.SetSerializer(s)
	return
}

// checkBuffer takes a quick look at how many firehose record errors were
// handled by the code.  It will error out if the number of errors found
// are not equal to the number of errors expected.
func checkBuffer(t *testing.T, f *FirehoseOutput, numErrors int64) (err error) {
	count := int64(len(f.errorBuffer))
	if count != numErrors {
		err = errors.New(fmt.Sprintf("Got buffer length of %d, expected %d", count, numErrors))
		t.Fail()
	}
	return
}

// initFirehose is the initializer for FirehoseOutput but with a mockFirehose
// as the svc variable
func initFirehose(t *testing.T, m *mockFirehose) (f *FirehoseOutput) {
	f = &FirehoseOutput{}
	f.svc = m
	attachSerializer(t, f)
	return f
}

// TestWriteToFirehoseAllSuccess tests the case when no error is returned
// from our mock AWS firehose function.
func TestWriteToFirehoseAllSuccess(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 0,
		t:                  t,
	}
	f := initFirehose(t, m)

	generatedLines, err := generateLines(10)
	if err != nil {
		t.FailNow()
	}

	err = f.Write(generatedLines)
	assert.Equal(t, int64(1), m.numOfPuts)
}

// TestWriteToFirehoseAllSuccess tests the case when no error is returned
// from our mock AWS firehose function when submitting 500 metrics in one
// write.
func TestWriteToFirehose500AllSuccess(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 0,
		t:                  t,
	}
	f := initFirehose(t, m)

	generatedLines, err := generateLines(500)
	if err != nil {
		t.FailNow()
	}

	err = f.Write(generatedLines)
	assert.Equal(t, int64(1), m.numOfPuts)
}

// TestWriteToFirehoseAllSuccess tests the case when no error is returned
// from our mock AWS firehose function when we send more metrics than
// a single write to Firehose can accomodate.
func TestWriteToFirehose550AllSuccess(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 0,
		t:                  t,
	}
	f := initFirehose(t, m)

	generatedLines, err := generateLines(550)
	if err != nil {
		t.FailNow()
	}

	err = f.Write(generatedLines)
	assert.Equal(t, int64(2), m.numOfPuts)
}

// TestWriteToFirehoseOneSuccessOneError tests the case when we do see
// a total failure and are capable of retry
func TestWriteToFirehoseTotalFail(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 10,
		t:                  t,
	}
	f := initFirehose(t, m)
	f.MaxSubmitAttempts = 1

	// generate 10 lines, all 10 will have an error
	generatedLines, err := generateLines(10)
	if err != nil {
		t.FailNow()
	}

	// write attempt, should see one error, and that
	// single metric should be re-queued for submission
	// upon next write
	err = f.Write(generatedLines)
	assert.Equal(t, int64(0), m.numOfPuts)

	// do another write. We should see 2 puts. 0 for the
	// original attempt, 1 for the retry of the earlier error
	// and then 1 for the new data.
	err = f.Write(generatedLines)
	assert.Equal(t, int64(2), m.numOfPuts)
}

// TestWriteToFirehoseOneSuccessOneError tests the case when we do see
// an error when attempting to send to firehose.
func TestWriteToFirehoseOneSuccessOneError(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 1,
		t:                  t,
	}
	f := initFirehose(t, m)
	f.MaxSubmitAttempts = 1

	// generate 2 lines, but 1 will have an error
	generatedLines, err := generateLines(2)
	if err != nil {
		t.FailNow()
	}

	// write attempt, should see one error, and that
	// single metric should be re-queued for submission
	// upon next write
	err = f.Write(generatedLines)
	assert.Equal(t, int64(1), m.numOfPuts)

	// do another write. We should see 3 puts. 1 for the
	// original attempt, 1 for the retry of the earlier error,
	// and then 1 for the new data.
	err = f.Write(generatedLines)
	assert.Equal(t, int64(3), m.numOfPuts)
}

// TestWriteToFirehoseOneSuccessOneErrorNoRetry tests the case when we do see
// an error when attempting to send to firehose, but we hit the max retry
// count and the retry data is discarded.
func TestWriteToFirehoseOneSuccessOneErrorNoRetry(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 1,
		t:                  t,
	}
	f := initFirehose(t, m)
	f.MaxSubmitAttempts = 0

	// generate 2 lines, but 1 will have an error
	generatedLines, err := generateLines(2)
	if err != nil {
		t.FailNow()
	}

	// write attempt, should see one error, and that
	// single metric should be re-queued for submission
	// upon next write
	err = f.Write(generatedLines)
	assert.Equal(t, int64(1), m.numOfPuts)

	// do another write. We should see 2 puts. 1 for the
	// original attempt, 0 for the retry of the earlier error (as it is discarded),
	// and then 1 for the new data.
	err = f.Write(generatedLines)
	assert.Equal(t, int64(2), m.numOfPuts)
}

// TestWriteToFirehoseFullFailure tests the case when we simply cannot write to firehose
// whatsoever.  Desired output is a retry next time it writes.
func TestWriteToFirehoseTotalFailNoRetry(t *testing.T) {
	m := &mockFirehose{
		numErrorsRemaining: 500,
		t:                  t,
	}
	f := initFirehose(t, m)
	f.MaxSubmitAttempts = 0

	generatedLines, err := generateLines(500)
	if err != nil {
		t.FailNow()
	}

	// write attempt, should see a total failure
	err = f.Write(generatedLines)
	assert.Equal(t, int64(0), m.numOfPuts)
}
