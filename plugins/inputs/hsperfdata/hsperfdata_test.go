package hsperfdata

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const datadir = "hsperfdata_tokuhirom"

func TestGatherNoTags(t *testing.T) {
	setup()
	defer teardown()

	hs := &Hsperfdata{User: "tokuhirom"}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))

	assert.New(t).Equal(acc.NMetrics(), uint64(2))

	acc.Lock()
	defer acc.Unlock()
	for _, p := range acc.Metrics {
		if reflect.DeepEqual(
			map[string]string{
				"pid": "13223",
			},
			p.Tags) {

			assert.Equal(
				t,
				212,
				len(p.Fields))

			// verify some of the fields (there's quite a lot!)
			assert.Equal(
				t,
				"367001600",
				p.Fields["sun.gc.generation.2.space.0.capacity"])
			assert.Equal(
				t,
				"/System/Library/Java/JavaVirtualMachines/1.6.0.jdk/Contents/Libraries",
				p.Fields["sun.property.sun.boot.library.path"])

		} else if reflect.DeepEqual(map[string]string{
			"pid":      "21916",
			"procname": "org.jetbrains.jps.cmdline.Launcher",
		}, p.Tags) {

			assert.Equal(
				t,
				253,
				len(p.Fields),
				fmt.Sprintf("wrong number of fields in %v", p))

			assert.Equal(
				t,
				"3313990237",
				p.Fields["java.ci.totalTime"])
			assert.Equal(
				t,
				"/Library/Java/JavaVirtualMachines/jdk1.8.0_31.jdk/Contents/Home/jre/lib",
				p.Fields["sun.property.sun.boot.library.path"])

		} else {
			msg := fmt.Sprintf("unknown with tags %v", p.Tags)
			assert.Fail(t, msg)
		}
	}
}

func TestGatherWithTags(t *testing.T) {
	setup()
	defer teardown()

	hs := &Hsperfdata{
		User: "tokuhirom",
		Tags: []string{"java.property.java.vm.specification.vendor", "sun.gc.policy.minorCollectionSlope"}}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))

	assert.New(t).Equal(acc.NMetrics(), uint64(2))

	acc.Lock()
	defer acc.Unlock()
	for _, p := range acc.Metrics {
		if reflect.DeepEqual(
			map[string]string{
				"pid": "13223",
				"java.property.java.vm.specification.vendor": "Sun Microsystems Inc.",
			},
			p.Tags) {

			assert.Equal(
				t,
				211,
				len(p.Fields))

			assert.NotContains(
				t,
				p.Fields,
				"java.property.java.vm.specification.vendor",
				"value promoted to tag")

			// verify some of the fields (there's quite a lot!)
			assert.Equal(
				t,
				"367001600",
				p.Fields["sun.gc.generation.2.space.0.capacity"])
			assert.Equal(
				t,
				"/System/Library/Java/JavaVirtualMachines/1.6.0.jdk/Contents/Libraries",
				p.Fields["sun.property.sun.boot.library.path"])

		} else if reflect.DeepEqual(map[string]string{
			"pid":      "21916",
			"procname": "org.jetbrains.jps.cmdline.Launcher",
			"java.property.java.vm.specification.vendor": "Oracle Corporation",
			"sun.gc.policy.minorCollectionSlope":         "0",
		}, p.Tags) {

			assert.Equal(
				t,
				251,
				len(p.Fields))

			assert.NotContains(
				t,
				p.Fields,
				"java.property.java.vm.specification.vendor",
				"value promoted to tag")

			assert.Equal(
				t,
				"3313990237",
				p.Fields["java.ci.totalTime"])
			assert.Equal(
				t,
				"/Library/Java/JavaVirtualMachines/jdk1.8.0_31.jdk/Contents/Home/jre/lib",
				p.Fields["sun.property.sun.boot.library.path"])

		} else {
			msg := fmt.Sprintf("unknown with tags %v", p.Tags)
			assert.Fail(t, msg)
		}
	}
}

func TestNoDirectoryNoMeasurements(t *testing.T) {
	hs := &Hsperfdata{User: "tokuhirom"}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))
	assert.New(t).Equal(acc.NMetrics(), uint64(0))
}

func setup() {
	_, filename, _, _ := runtime.Caller(1)
	src := filepath.Join(
		strings.Replace(filename, "hsperfdata_test.go", "testdata", 1),
		datadir)
	dest := filepath.Join(
		os.TempDir(),
		datadir)
	os.Symlink(src, dest)
}

func teardown() {
	os.Remove(filepath.Join(os.TempDir(), datadir))
}
