// +build linux

package pressure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var pressureData = []byte(`some avg10=0.00 avg60=0.01 avg300=0.04 total=13595686`)


func TestParsePressureData(t *testing.T) {
	expectedData := &pressureFields{
		avg10:	0,
		avg60:	0.01,
		avg300: 0.04,
		total:	uint64(13595686),
	}
	parsedData := parsePressureData(pressureData)
	if parsingErrors != 0 {
		t.Fatal("Parsing errors")
	}

	as := assert.New(t)
	as.Equal(expectedData, parsedData)
}


