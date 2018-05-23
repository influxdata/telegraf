package syslog

import (
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var defaultTime = time.Unix(0, 0)
var maxP = uint8(191)
var maxV = uint16(999)
var maxTS = "2017-12-31T23:59:59.999999+00:00"
var maxH = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"
var maxA = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"
var maxPID = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"
var maxMID = "abcdefghilmnopqrstuvzabcdefghilm"
var message7681 = strings.Repeat("l", 7681)

func TestListenError(t *testing.T) {
	receiver := &Syslog{
		Address: "wrong address",
	}
	require.Error(t, receiver.Start(&testutil.Accumulator{}))
}
