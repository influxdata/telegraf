package cpu

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestParseISAInfo(t *testing.T) {
	cases := []struct {
		filename string
		expected []string
	}{
		{
			"1cpu_1core_isainfo.txt",
			[]string{"rdseed", "adx", "avx2", "fma", "bmi2", "bmi1", "rdrand", "f16c", "vmx",
				"avx", "xsave", "pclmulqdq", "aes", "movbe", "sse4.2", "sse4.1", "ssse3", "popcnt",
				"tscp", "cx16", "sse3", "sse2", "sse", "fxsr", "mmx", "cmov", "amd_sysc", "cx8",
				"tsc", "fpu"},
		},
		{
			"2cpu_1core_isainfo.txt",
			[]string{"rdseed", "adx", "avx2", "fma", "bmi2", "bmi1", "rdrand", "f16c", "vmx",
				"avx", "xsave", "pclmulqdq", "aes", "movbe", "sse4.2", "sse4.1", "ssse3", "popcnt",
				"tscp", "cx16", "sse3", "sse2", "sse", "fxsr", "mmx", "cmov", "amd_sysc", "cx8",
				"tsc", "fpu"},
		},
		{
			"2cpu_8core_isainfo.txt",
			[]string{"vmx", "avx", "xsave", "pclmulqdq", "aes", "sse4.2", "sse4.1", "ssse3", "popcnt",
				"tscp", "cx16", "sse3", "sse2", "sse", "fxsr", "mmx", "cmov", "amd_sysc", "cx8",
				"tsc", "fpu"},
		},
	}

	for _, tc := range cases {
		content, err := ioutil.ReadFile(filepath.Join("testdata", "solaris", tc.filename))
		if err != nil {
			t.Errorf("cannot read test case: %s", err)
		}

		sort.Strings(tc.expected)

		flags, err := parseISAInfo(string(content))
		if err != nil {
			t.Fatalf("parseISAInfo: %s", err)
		}

		if !reflect.DeepEqual(tc.expected, flags) {
			t.Fatalf("Bad flags\nExpected: %v\n   Actual: %v", tc.expected, flags)
		}
	}
}

func TestParseProcessorInfo(t *testing.T) {
	cases := []struct {
		filename string
		expected []InfoStat
	}{
		{
			"1cpu_1core_psrinfo.txt",
			[]InfoStat{
				{CPU: 0, VendorID: "GenuineIntel", Family: "6", Model: "78", Stepping: 3, PhysicalID: "0", CoreID: "0", Cores: 1, ModelName: "Intel(r) Core(tm) i7-6567U CPU @ 3.30GHz", Mhz: 3312},
			},
		},
		{
			"2cpu_1core_psrinfo.txt",
			[]InfoStat{
				{CPU: 0, VendorID: "GenuineIntel", Family: "6", Model: "78", Stepping: 3, PhysicalID: "0", CoreID: "0", Cores: 1, ModelName: "Intel(r) Core(tm) i7-6567U CPU @ 3.30GHz", Mhz: 3312},
				{CPU: 1, VendorID: "GenuineIntel", Family: "6", Model: "78", Stepping: 3, PhysicalID: "1", CoreID: "0", Cores: 1, ModelName: "Intel(r) Core(tm) i7-6567U CPU @ 3.30GHz", Mhz: 3312},
			},
		},
		{
			"2cpu_8core_psrinfo.txt",
			[]InfoStat{
				{CPU: 0, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "0", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 1, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "1", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 2, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "2", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 3, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "3", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 4, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "4", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 5, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "5", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 6, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "6", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 7, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "0", CoreID: "7", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 8, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "0", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 9, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "1", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 10, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "2", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 11, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "3", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 12, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "4", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 13, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "5", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 14, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "6", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
				{CPU: 15, VendorID: "GenuineIntel", Family: "6", Model: "45", Stepping: 7, PhysicalID: "1", CoreID: "7", Cores: 2, ModelName: "Intel(r) Xeon(r) CPU E5-2670 0 @ 2.60GHz", Mhz: 2600},
			},
		},
	}

	for _, tc := range cases {
		content, err := ioutil.ReadFile(filepath.Join("testdata", "solaris", tc.filename))
		if err != nil {
			t.Errorf("cannot read test case: %s", err)
		}

		cpus, err := parseProcessorInfo(string(content))
		if err != nil {
			t.Errorf("cannot parse processor info: %s", err)
		}

		if !reflect.DeepEqual(tc.expected, cpus) {
			t.Fatalf("Bad Processor Info\nExpected: %v\n   Actual: %v", tc.expected, cpus)
		}
	}
}
