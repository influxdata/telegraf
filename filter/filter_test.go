package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileFilter(t *testing.T) {
	f, err := CompileFilter([]string{})
	assert.NoError(t, err)
	assert.Nil(t, f)

	f, err = CompileFilter([]string{"cpu"})
	assert.NoError(t, err)
	assert.True(t, f.Match("cpu"))
	assert.False(t, f.Match("cpu0"))
	assert.False(t, f.Match("mem"))

	f, err = CompileFilter([]string{"cpu*"})
	assert.NoError(t, err)
	assert.True(t, f.Match("cpu"))
	assert.True(t, f.Match("cpu0"))
	assert.False(t, f.Match("mem"))

	f, err = CompileFilter([]string{"cpu", "mem"})
	assert.NoError(t, err)
	assert.True(t, f.Match("cpu"))
	assert.False(t, f.Match("cpu0"))
	assert.True(t, f.Match("mem"))

	f, err = CompileFilter([]string{"cpu", "mem", "net*"})
	assert.NoError(t, err)
	assert.True(t, f.Match("cpu"))
	assert.False(t, f.Match("cpu0"))
	assert.True(t, f.Match("mem"))
	assert.True(t, f.Match("network"))
}

var benchbool bool

func BenchmarkFilterSingleNoGlobFalse(b *testing.B) {
	f, _ := CompileFilter([]string{"cpu"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("network")
	}
	benchbool = tmp
}

func BenchmarkFilterSingleNoGlobTrue(b *testing.B) {
	f, _ := CompileFilter([]string{"cpu"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("cpu")
	}
	benchbool = tmp
}

func BenchmarkFilter(b *testing.B) {
	f, _ := CompileFilter([]string{"cpu", "mem", "net*"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("network")
	}
	benchbool = tmp
}

func BenchmarkFilterNoGlob(b *testing.B) {
	f, _ := CompileFilter([]string{"cpu", "mem", "net"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("net")
	}
	benchbool = tmp
}

func BenchmarkFilter2(b *testing.B) {
	f, _ := CompileFilter([]string{"aa", "bb", "c", "ad", "ar", "at", "aq",
		"aw", "az", "axxx", "ab", "cpu", "mem", "net*"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("network")
	}
	benchbool = tmp
}

func BenchmarkFilter2NoGlob(b *testing.B) {
	f, _ := CompileFilter([]string{"aa", "bb", "c", "ad", "ar", "at", "aq",
		"aw", "az", "axxx", "ab", "cpu", "mem", "net"})
	var tmp bool
	for n := 0; n < b.N; n++ {
		tmp = f.Match("net")
	}
	benchbool = tmp
}
