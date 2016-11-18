package hsperfdata

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The code paths exercised depend on the contents of the

func TestGatherNoTags(t *testing.T) {
	hs := &Hsperfdata{Directory: GetDirectory("testdata/good")}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))

	assert := assert.New(t)
	assert.Equal(acc.NMetrics(), uint64(2))

	acc.Lock()
	defer acc.Unlock()
	for _, p := range acc.Metrics {
		if reflect.DeepEqual(
			map[string]string{
				"pid": "13223",
			},
			p.Tags) {

			assert.Equal(
				212,
				len(p.Fields))

			// verify some of the fields (there's quite a lot!)
			assert.Equal(
				int64(367001600),
				p.Fields["sun.gc.generation.2.space.0.capacity"])
			assert.Equal(
				"/System/Library/Java/JavaVirtualMachines/1.6.0.jdk/Contents/Libraries",
				p.Fields["sun.property.sun.boot.library.path"])

			assert.Equal(
				time.Date(2015, time.June, 10, 9, 40, 29, 493350542, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else if reflect.DeepEqual(map[string]string{
			"pid":      "21916",
			"procname": "org.jetbrains.jps.cmdline.Launcher",
		}, p.Tags) {

			assert.Equal(
				253,
				len(p.Fields),
				fmt.Sprintf("wrong number of fields in %v", p))

			assert.Equal(
				int64(3313990237),
				p.Fields["java.ci.totalTime"])
			assert.Equal(
				"/Library/Java/JavaVirtualMachines/jdk1.8.0_31.jdk/Contents/Home/jre/lib",
				p.Fields["sun.property.sun.boot.library.path"])

			assert.Equal(
				time.Date(2015, time.June, 11, 6, 48, 1, 661669176, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else {
			assert.Fail(fmt.Sprintf("unknown with tags %v", p.Tags))
		}
	}
}

func TestGatherWithTags(t *testing.T) {
	hs := &Hsperfdata{
		Directory: GetDirectory("testdata/good"),
		Tags: []string{
			"java.property.java.vm.specification.vendor", // a string-type
			"sun.gc.policy.minorCollectionSlope",         // a long-type
		}}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))

	assert := assert.New(t)
	assert.Equal(acc.NMetrics(), uint64(2), "num metrics gathered")

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
				211,
				len(p.Fields))

			assert.NotContains(
				p.Fields,
				"java.property.java.vm.specification.vendor",
				"value promoted to tag")

			// verify some of the fields (there's quite a lot!)
			assert.Equal(
				int64(367001600),
				p.Fields["sun.gc.generation.2.space.0.capacity"])
			assert.Equal(
				"/System/Library/Java/JavaVirtualMachines/1.6.0.jdk/Contents/Libraries",
				p.Fields["sun.property.sun.boot.library.path"])

			assert.Equal(
				time.Date(2015, time.June, 10, 9, 40, 29, 493350542, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else if reflect.DeepEqual(map[string]string{
			"pid":      "21916",
			"procname": "org.jetbrains.jps.cmdline.Launcher",
			"java.property.java.vm.specification.vendor": "Oracle Corporation",
			"sun.gc.policy.minorCollectionSlope":         "0",
		}, p.Tags) {

			assert.Equal(
				251,
				len(p.Fields))

			assert.NotContains(
				p.Fields,
				"java.property.java.vm.specification.vendor",
				"value promoted to tag")

			assert.Equal(
				int64(3313990237),
				p.Fields["java.ci.totalTime"])
			assert.Equal(
				"/Library/Java/JavaVirtualMachines/jdk1.8.0_31.jdk/Contents/Home/jre/lib",
				p.Fields["sun.property.sun.boot.library.path"])

			assert.Equal(
				time.Date(2015, time.June, 11, 6, 48, 1, 661669176, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else {
			assert.Fail(fmt.Sprintf("unknown with tags %v", p.Tags))
		}
	}
}

func TestGatherWithFilter(t *testing.T) {
	hs := &Hsperfdata{
		Directory: GetDirectory("testdata/good"),
		Filter:    `^java\.`}

	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))

	assert := assert.New(t)
	assert.Equal(acc.NMetrics(), uint64(2), "num metrics gathered")

	acc.Lock()
	defer acc.Unlock()

	for _, p := range acc.Metrics {
		if reflect.DeepEqual(
			map[string]string{
				"pid": "13223",
			},
			p.Tags) {

			assert.Equal(
				24,
				len(p.Fields))

			// verify some of the fields (there's quite a lot!)
			assert.Equal(
				int64(39846),
				p.Fields["java.cls.loadedClasses"])
			assert.NotContains(
				p.Fields,
				"sun.property.sun.boot.library.path",
				"value excluded by filter")

			assert.Equal(
				time.Date(2015, time.June, 10, 9, 40, 29, 493350542, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else if reflect.DeepEqual(map[string]string{
			"pid":      "21916",
			"procname": "org.jetbrains.jps.cmdline.Launcher",
		}, p.Tags) {

			assert.Equal(
				24,
				len(p.Fields))

			assert.Equal(
				int64(3313990237),
				p.Fields["java.ci.totalTime"])
			assert.NotContains(
				p.Fields,
				"sun.property.sun.boot.library.path",
				"value excluded by filter")

			assert.Equal(
				time.Date(2015, time.June, 11, 6, 48, 1, 661669176, time.UTC),
				p.Time.In(time.UTC),
				fmt.Sprintf("unexpected time %v", p.Time))

		} else {
			assert.Fail(fmt.Sprintf("unknown with tags %v", p.Tags))
		}
	}
}

func TestNoDirectoryNoMeasurements(t *testing.T) {
	hs := &Hsperfdata{Directory: "hathathat"}
	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))
	assert.New(t).Equal(acc.NMetrics(), uint64(0), "no metrics gathered")
}

func TestErrorsOnInvalidFormat(t *testing.T) {
	hs := &Hsperfdata{Directory: "testdata/bad"}
	acc := testutil.Accumulator{}
	assert := assert.New(t)

	err := hs.Gather(&acc)
	require.Error(t, err)

	assert.Contains(err.Error(), "EOF", "file '0' is zero-length")
	assert.Contains(err.Error(), "illegal magic 3800001267", "file '1' is random data")
	assert.Contains(err.Error(), "invalid binary: <nil>", "file '2' is truncated")

	assert.Equal(acc.NMetrics(), uint64(0), "no metrics gathered")
}

func TestNoPidFilesNoMeasurements(t *testing.T) {
	// the directory with this file has subdirectories and go files,
	// but no valid pid files
	hs := &Hsperfdata{Directory: GetDirectory(".")}
	acc := testutil.Accumulator{}
	require.NoError(t, hs.Gather(&acc))
	assert.New(t).Equal(acc.NMetrics(), uint64(0), "no metrics gathered")
}

func GetDirectory(subdir string) string {
	_, filename, _, _ := runtime.Caller(1)
	return filepath.Join(
		strings.Replace(filename, "hsperfdata_test.go", "", 1),
		subdir)
}
