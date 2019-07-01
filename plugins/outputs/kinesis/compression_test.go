package kinesis

import (
	"testing"
	"time"
)

func TestGoodCompression(t *testing.T) {
	tests := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		time.Now().String(),
		`abcdefghijklmnopqrstuvwzyz1234567890@~|\/?><@~#+=!"£$%^&_*(){}[]`,
	}

	for _, test := range tests {
		_, err := gzipMetrics([]byte(test))
		if err != nil {
			t.Logf("Failed to gzip test data")
			t.Fail()
		}

		// Snappy doesn't error, so we can only look for panics
		snappyMetrics([]byte(test))
	}
}

func TestBadGzipCompressionLevel(t *testing.T) {
	oldlevel := gzipCompressionLevel
	gzipCompressionLevel = 11
	defer func() { gzipCompressionLevel = oldlevel }()

	_, err := gzipMetrics([]byte(time.Now().String()))
	if err == nil {
		t.Logf("Expect gzip to fail because of a bad compression level")
		t.Fail()
	}

}
