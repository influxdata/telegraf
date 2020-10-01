package procstat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPattern(b *testing.B) {
	f, err := NewNativeFinder()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		_, err := f.Pattern(".*")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFullPattern(b *testing.B) {
	f, err := NewNativeFinder()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		_, err := f.FullPattern(".*")
		if err != nil {
			panic(err)
		}
	}
}
