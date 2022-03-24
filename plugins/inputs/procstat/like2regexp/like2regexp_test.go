// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package like2regexp

import (
	"regexp"
	"testing"
)

var tests = []struct {
	input, expected   string
	matches, excludes []string
}{
	{`abc`, `(?i:^abc$)`,
		[]string{`abc`, `ABC`, `AbC`},
		[]string{`def`, `DEF`},
	},
	{`a.c`, `(?i:^a\.c$)`,
		[]string{`a.c`},
		[]string{`abc`, `abbc`, `xcc`, `abb`, `a`, `ac`},
	},
	{`%watch%`, `(?i:^.*watch.*$)`,
		[]string{`amazon-cloudwatch-agent.exe`, `WatchSomething`, `watch`, `amazoncloudwatch`},
		[]string{`amazon-ssm-agent.exe`, `wach`, `amazoncloudwatc`},
	},
	{`[a-z]`, `(?i:^[a-z]$)`,
		[]string{`a`, `c`, `x`},
		[]string{`1`, `2`, `3`, `ab`},
	},
	{`[^a-z]`, `(?i:^[^a-z]$)`,
		[]string{`1`, `2`},
		[]string{`a`, `c`, `x`, `abc`},
	},
	{`abc^def[^a-z]`, `(?i:^abc\^def[^a-z]$)`,
		[]string{`abc^def1`, `abc^def_`},
		[]string{`abc^defa`, `abc^defz`, `abcdef`},
	},
	{`a_c`, `(?i:^a.c$)`,
		[]string{`a.c`, `abc`, `acc`},
		[]string{`bbc`, `abb`, `ac`},
	},
	{`%a_c%`, `(?i:^.*a.c.*$)`,
		[]string{`aa.cc`, `abc`, `acc`, `xxxabcxxx`},
		[]string{`abbc`, `abbb`, `bbac`},
	},
	{`%[_][_]%`, `(?i:^.*[_][_].*$)`,
		[]string{`__`, `a__cc`},
		[]string{`_x_`, `acc`},
	},
	{`he[[]ll]o.exe`, `(?i:^he[[]ll]o\.exe$)`,
		[]string{`he[ll]o.exe`},
		[]string{`hello.exe`},
	},
}

func TestWMILikeToRegexp(t *testing.T) {
	for _, test := range tests {
		output := WMILikeToRegexp(test.input)
		if test.expected != output {
			t.Errorf("translated regexp does not match expected result:\n\tExpected: %v\n\tOutput  : %v", test.expected, output)
		}

		re := regexp.MustCompile(output)
		for _, m := range test.matches {
			if !re.MatchString(m) {
				t.Errorf("case '%v': generated regexp %v does not match value '%v'", test.input, output, m)
			}
		}
		for _, m := range test.excludes {
			if re.MatchString(m) {
				t.Errorf("case '%v': generated regexp %v should not match value '%v'", test.input, output, m)
			}
		}
	}
}