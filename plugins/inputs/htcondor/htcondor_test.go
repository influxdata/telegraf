package htcondor

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// var condorOutputRegex = regexp.MustCompile(`(?m)(?P<jobs>\d+\s*jobs);\s*(?P<completed>\d+\s*completed),\s*(?P<removed>\d+\s*removed),\s*(?P<idle>\d+\s*idle),\s*(?P<running>\d+\s*running),\s*(?P<held>\d+\s*held),\s*(?P<suspended>\d+\s*suspended)`)
var htcondorOutput = `

-- Schedd: rocks-186.sdsc.edu : <172.30.0.200:52641?...
 ID      OWNER            SUBMITTED     RUN_TIME ST PRI SIZE CMD

0 jobs; 0 completed, 0 removed, 0 idle, 0 running, 0 held, 0 suspended`

func TestHTCondorOutputGather(t *testing.T) {

	var regexGroupMatch = condorOutputRegex.FindAllStringSubmatch(string(htcondorOutput), -1)
	fields := make(map[string]interface{})

	for i := 1; i < len(regexGroupMatch[0]); i++ {
		var matched = strings.Split(regexGroupMatch[0][i], " ") // "1 jobs" --> ["1", "jobs"]
		var fieldKey = matched[1]
		var fieldvalue, _ = strconv.ParseInt(matched[0], 10, 64)
		fields[fieldKey] = fieldvalue
	}

	assert.Equal(t, fields["jobs"], int64(0), "Total job must be 0")
	assert.Equal(t, fields["completed"], int64(0), "Completed job must be 0")
	assert.Equal(t, fields["removed"], int64(0), "Removed job must be 0")
	assert.Equal(t, fields["idle"], int64(0), "Idle job must be 0")
	assert.Equal(t, fields["running"], int64(0), "Running job must be 0")
	assert.Equal(t, fields["held"], int64(0), "Held job must be 0")
	assert.Equal(t, fields["suspended"], int64(0), "Suspended job must be 0")
}
