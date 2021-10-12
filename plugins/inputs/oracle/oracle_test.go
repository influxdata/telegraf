package oracle

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validResult = `
oracle_tablespaces,instance=XE,tbs_name=SYSAUX used_space_mb=605,free_space_mb=26812,percent_used=2.21,max_size_mb=27417
oracle_tablespaces,instance=XE,tbs_name=SYSTEM used_space_mb=353,free_space_mb=247,percent_used=58.78,max_size_mb=600
oracle_tablespaces,instance=XE,tbs_name=TEMP used_space_mb=0,free_space_mb=26797,percent_used=0,max_size_mb=26797
oracle_tablespaces,instance=XE,tbs_name=UNDOTBS1 used_space_mb=1,free_space_mb=26800,percent_used=0,max_size_mb=26802
oracle_tablespaces,instance=XE,tbs_name=USERS used_space_mb=3,free_space_mb=11261,percent_used=0.02,max_size_mb=11264
oracle_connectioncount,instance=XE,metric_name=ACTIVE metric_value=19
oracle_status,instance=XE,metric_name=database_status metric_value=1
oracle_status,instance=XE,metric_name=instance_status metric_value=1`

type runnerMock struct {
	out    []byte
	errout []byte
	err    error
}

func newRunnerMock(out []byte, errout []byte, err error) exec.Runner {
	return &runnerMock{
		out:    out,
		errout: errout,
		err:    err,
	}
}

func (r runnerMock) Run(_ string, _ []string, _ []string, _ []byte, _ time.Duration) ([]byte, []byte, error) {
	return r.out, r.errout, r.err
}

func TestOracleResults(t *testing.T) {
	parser, _ := parsers.NewInfluxParser()
	o := &Oracle{
		runner: newRunnerMock([]byte(validResult), nil, nil),
		parser: parser,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(o.Gather)
	require.NoError(t, err)
	assert.Equal(t, acc.NFields(), 23, "non-numeric measurements should be ignored")

	acc.AssertContainsFields(t, "oracle_status", map[string]interface{}{
		"metric_value": float64(1),
	})
	acc.AssertContainsFields(t, "oracle_connectioncount", map[string]interface{}{
		"metric_value": float64(19),
	})
}

func TestOracleConfiguration(t *testing.T) {
	creator := inputs.Inputs["oracle"]
	assert.NotNil(t, creator)
	input := creator()
	assert.NotNil(t, input)
	assert.NotEmpty(t, input.Description())
	assert.NotEmpty(t, input.SampleConfig())
	oracle, _ := input.(*Oracle)
	// python3 is used OOTB
	assert.Equal(t, oracle.Python, "python3")
	// username, password and dsn is required
	assert.Error(t, oracle.Init())
	oracle.Username = "system"
	assert.Error(t, oracle.Init())
	oracle.Password = "oracle"
	assert.Error(t, oracle.Init())
	oracle.DSN = "XE"
	oracle.Env = []string{"whatever=yes"}
	oracle.Init() //nolint:errcheck
	assert.Contains(t, oracle.scriptEnv, "whatever=yes")
	assert.Contains(t, oracle.scriptEnv, os.Environ()[0])
}

func TestOracleExecutionError(t *testing.T) {
	o := NewOracle()
	o.runner = newRunnerMock(nil, nil, fmt.Errorf("exit status code 1"))

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(o.Gather))
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}

func TestOracleInvalidResult(t *testing.T) {
	o := NewOracle()
	o.runner = newRunnerMock([]byte("wrongData"), []byte{}, nil)

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(o.Gather))
	assert.Equal(t, acc.NFields(), 0, "No new points should have been added")
}
