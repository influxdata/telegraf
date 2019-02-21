package kinesis

import (
	"strings"
	"testing"
)

func TestStringCompare(t *testing.T) {
	for i := 1; i <= 130; i++ {
		// check if sequence number with less than or equal to the number of digits
		//  is less than the max sequence number.
		if strings.Repeat("9", i) > maxSeq && i != 130 {
			t.Logf("Impossible; smaller sequence number greater than max. %d digits", i)
			t.FailNow()
		}
	}
}
