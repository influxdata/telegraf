package syslog

import (
	"math/rand"
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
var message7681 = getRandomString(7681)

func TestListenError(t *testing.T) {
	receiver := &Syslog{
		Address: "wrong address",
	}
	require.Error(t, receiver.Start(&testutil.Accumulator{}))
}

func getRandomString(n int) string {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // Number of letter indices fitting in 63 bits
	)

	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
