package ts3

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanLines(t *testing.T) {
	str := "line1\n\rline2\n\rline3"

	s := bufio.NewScanner(strings.NewReader(str))
	s.Split(ScanLines)

	lines := []string{
		"line1",
		"line2",
		"line3",
	}

	var i int
	for s.Scan() {
		if !assert.True(t, len(lines) > i) {
			return
		}
		assert.Equal(t, lines[i], s.Text())
		i++
	}

	assert.Equal(t, len(lines), i)
}
