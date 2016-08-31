package agent

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", now.UnixNano()),
		actual)
}

func TestAddNoPrecisionWithInterval(t *testing.T) {
	a := accumulator{}
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.SetPrecision(0, time.Second)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800000000000)),
		actual)
}

func TestAddNoIntervalWithPrecision(t *testing.T) {
	a := accumulator{}
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.SetPrecision(time.Second, time.Millisecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800000000000)),
		actual)
}

func TestAddDisablePrecision(t *testing.T) {
	a := accumulator{}
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.SetPrecision(time.Second, time.Millisecond)
	a.DisablePrecision()
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800082912748)),
		actual)
}

func TestDifferentPrecisions(t *testing.T) {
	a := accumulator{}
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.SetPrecision(0, time.Second)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm := <-a.metrics
	actual := testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800000000000)),
		actual)

	a.SetPrecision(0, time.Millisecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800083000000)),
		actual)

	a.SetPrecision(0, time.Microsecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800082913000)),
		actual)

	a.SetPrecision(0, time.Nanosecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", int64(1139572800082912748)),
		actual)
}

func TestAddDefaultTags(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,default=tag value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=101 %d", now.UnixNano()),
		actual)
}

func TestAddFields(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	fields := map[string]interface{}{
		"usage": float64(99),
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest usage=99")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test usage=99")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test usage=99 %d", now.UnixNano()),
		actual)
}

// Test that all Inf fields get dropped, and not added to metrics channel
func TestAddInfFields(t *testing.T) {
	inf := math.Inf(1)
	ninf := math.Inf(-1)

	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	fields := map[string]interface{}{
		"usage":  inf,
		"nusage": ninf,
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"}, now)

	assert.Len(t, a.metrics, 0)

	// test that non-inf fields are kept and not dropped
	fields["notinf"] = float64(100)
	a.AddFields("acctest", fields, map[string]string{})
	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest notinf=100")
}

// Test that nan fields are dropped and not added
func TestAddNaNFields(t *testing.T) {
	nan := math.NaN()

	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	fields := map[string]interface{}{
		"usage": nan,
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"}, now)

	assert.Len(t, a.metrics, 0)

	// test that non-nan fields are kept and not dropped
	fields["notnan"] = float64(100)
	a.AddFields("acctest", fields, map[string]string{})
	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest notnan=100")
}

func TestAddUint64Fields(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	fields := map[string]interface{}{
		"usage": uint64(99),
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest usage=99i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test usage=99i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test usage=99i %d", now.UnixNano()),
		actual)
}

func TestAddUint64Overflow(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	fields := map[string]interface{}{
		"usage": uint64(9223372036854775808),
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"})
	a.AddFields("acctest", fields, map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest usage=9223372036854775807i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test usage=9223372036854775807i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test usage=9223372036854775807i %d", now.UnixNano()),
		actual)
}

func TestAddInts(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": int(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": int32(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": int64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,default=tag value=101i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=101i")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=101i %d", now.UnixNano()),
		actual)
}

func TestAddFloats(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": float32(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=101 %d", now.UnixNano()),
		actual)
}

func TestAddStrings(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": "test"},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": "foo"},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=\"test\"")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=\"foo\" %d", now.UnixNano()),
		actual)
}

func TestAddBools(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &models.InputConfig{}

	a.AddFields("acctest",
		map[string]interface{}{"value": true}, map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": false}, map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=true")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=false %d", now.UnixNano()),
		actual)
}

// Test that tag filters get applied to metrics.
func TestAccFilterTags(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	filter := models.Filter{
		TagExclude: []string{"acc"},
	}
	assert.NoError(t, filter.CompileFilter())
	a.inputConfig = &models.InputConfig{}
	a.inputConfig.Filter = filter

	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest value=101 %d", now.UnixNano()),
		actual)
}

func TestAccAddError(t *testing.T) {
	errBuf := bytes.NewBuffer(nil)
	log.SetOutput(errBuf)
	defer log.SetOutput(os.Stderr)

	a := accumulator{}
	a.inputConfig = &models.InputConfig{}
	a.inputConfig.Name = "mock_plugin"

	a.AddError(fmt.Errorf("foo"))
	a.AddError(fmt.Errorf("bar"))
	a.AddError(fmt.Errorf("baz"))

	errs := bytes.Split(errBuf.Bytes(), []byte{'\n'})
	assert.EqualValues(t, 3, a.errCount)
	require.Len(t, errs, 4) // 4 because of trailing newline
	assert.Contains(t, string(errs[0]), "mock_plugin")
	assert.Contains(t, string(errs[0]), "foo")
	assert.Contains(t, string(errs[1]), "mock_plugin")
	assert.Contains(t, string(errs[1]), "bar")
	assert.Contains(t, string(errs[2]), "mock_plugin")
	assert.Contains(t, string(errs[2]), "baz")
}
