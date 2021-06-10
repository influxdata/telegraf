package metric

import (
	"hash/maphash"
	"testing"
	"time"
)

var m = New(
	"mymetric",
	map[string]string{
		"host":        "host.example.com",
		"mykey":       "myvalue",
		"another key": "another value",
	},
	map[string]interface{}{
		"f1": 1,
		"f2": 2,
		"f3": 3,
		"f4": 4,
		"f5": 5,
		"f6": 6,
		"f7": 7,
		"f8": 8,
	},
	time.Now(),
)

var result uint64

var hashSeed = maphash.MakeSeed()

func BenchmarkGroupID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		result = groupID(hashSeed, m.Name(), m.TagList(), m.Time())
	}
}
