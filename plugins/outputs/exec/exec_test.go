package exec

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExecBasic(t *testing.T) {
	outFile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(outFile.Name())

	scriptFile := fmt.Sprintf("testscript%ctestscript.cmd", os.PathSeparator)
	exec := getBasicExec(scriptFile, outFile.Name())

	err = exec.Connect()
	assert.NoError(t, err)

	err = exec.Write(testutil.MockMetrics())
	assert.NoError(t, err)

	validateFile(outFile.Name(), "executed\n", t)

	err = exec.Close()
	assert.NoError(t, err)
}

func TestExecNonExistingFile(t *testing.T) {
	outFile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(outFile.Name())

	exec := getBasicExec("nonExistingScript.sh", outFile.Name())

	err = exec.Connect()
	assert.NoError(t, err)

	err = exec.Write(testutil.MockMetrics())
	assert.Error(t, err)

	err = exec.Close()
	assert.NoError(t, err)
}

func getBasicExec(scriptFile, outFile string) *Exec {
	s, _ := serializers.NewInfluxSerializer()
	exec := &Exec{
		Command:    getTestScriptCmd(scriptFile, outFile),
		Timeout:    internal.Duration{Duration: time.Second * 5},
		runner:     CommandRunner{},
		serializer: s,
	}

	return exec
}

func getTestScriptCmd(scriptFile, outFile string) []string {
	_, filename, _, _ := runtime.Caller(1)
	script := strings.Replace(filename, "exec_test.go", scriptFile, 1)

	if runtime.GOOS != "windows" {
		script = "/bin/sh " + script
	}

	return append(strings.Split(script, " "), outFile)
}

func validateFile(filename, expected string, t *testing.T) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expected, string(buf))
}
