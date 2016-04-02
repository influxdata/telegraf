package agent

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	a := accumulator{}
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", float64(101), map[string]string{})
	a.Add("acctest", float64(101), map[string]string{"acc": "test"})
	a.Add("acctest", float64(101), map[string]string{"acc": "test"}, now)

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

func TestAddDefaultTags(t *testing.T) {
	a := accumulator{}
	a.addDefaultTag("default", "tag")
	now := time.Now()
	a.metrics = make(chan telegraf.Metric, 10)
	defer close(a.metrics)
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", float64(101), map[string]string{})
	a.Add("acctest", float64(101), map[string]string{"acc": "test"})
	a.Add("acctest", float64(101), map[string]string{"acc": "test"}, now)

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
	a.inputConfig = &internal_models.InputConfig{}

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
	a.inputConfig = &internal_models.InputConfig{}

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
	a.inputConfig = &internal_models.InputConfig{}

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
	a.inputConfig = &internal_models.InputConfig{}

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
	a.inputConfig = &internal_models.InputConfig{}

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
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", int(101), map[string]string{})
	a.Add("acctest", int32(101), map[string]string{"acc": "test"})
	a.Add("acctest", int64(101), map[string]string{"acc": "test"}, now)

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
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", float32(101), map[string]string{"acc": "test"})
	a.Add("acctest", float64(101), map[string]string{"acc": "test"}, now)

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
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", "test", map[string]string{"acc": "test"})
	a.Add("acctest", "foo", map[string]string{"acc": "test"}, now)

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
	a.inputConfig = &internal_models.InputConfig{}

	a.Add("acctest", true, map[string]string{"acc": "test"})
	a.Add("acctest", false, map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest,acc=test,default=tag value=true")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test,default=tag value=false %d", now.UnixNano()),
		actual)
}
