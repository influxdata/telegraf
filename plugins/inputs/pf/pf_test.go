// +build freebsd

package pf

import (
	"log"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

type measurementResult struct {
	tags   map[string]string
	fields map[string]interface{}
}

func fakeexecFunc(i int, t *testing.T, desiredCmd string, desiredArgs ...string) func(string, ...string) *exec.Cmd {
	return func(cmd string, args ...string) *exec.Cmd {
		if cmd != desiredCmd || !reflect.DeepEqual(args, desiredArgs) {
			t.Errorf("%d: not invoked correctly! %s - %#v vs %s - %#v", i, cmd, args, desiredCmd, desiredArgs)
		}
		return nil
	}
}

func TestPfctlInvocation(t *testing.T) {
	type pfctlInvocationTestCase struct {
		config PF
		cmd    string
		args   []string
	}

	var testCases = []pfctlInvocationTestCase{
		// 0: no sudo
		pfctlInvocationTestCase{
			config: PF{UseSudo: false},
			cmd:    "fakepfctl",
			args:   []string{"-s", "info"},
		},
		// 1: with sudo
		pfctlInvocationTestCase{
			config: PF{UseSudo: true},
			cmd:    "fakesudo",
			args:   []string{"fakepfctl", "-s", "info"},
		},
	}

	for i, tt := range testCases {
		execLookPath = func(cmd string) (string, error) { return "fake" + cmd, nil }
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			log.Printf("running #%d\n", i)
			execCommand = fakeexecFunc(i, t, tt.cmd, tt.args...)
			_, err := tt.config.buildPfctlCmd()
			if err != nil {
				t.Fatalf("error when running buildPfctlCmd: %s", err)
			}
		})
	}
}

func TestPfMeasurements(t *testing.T) {
	type pfTestCase struct {
		TestInput    string
		err          error
		measurements []measurementResult
	}

	testCases := []pfTestCase{
		// 0: nil input should raise an error
		pfTestCase{TestInput: "", err: errParseHeader},
		// 1: changes to pfctl output should raise an error
		pfTestCase{TestInput: `Status: Enabled for 161 days 21:24:45         Debug: Urgent

Interface Stats for re1               IPv4             IPv6
  Bytes In                   2585823744614    1059233657221
  Bytes Out                  1227266932673    3274698578875
  Packets In
    Passed                      2289953086       1945437219
    Blocked                      392835739            48609
  Packets Out
    Passed                      1649146326       2605569054
    Blocked                            107                0

State Table                          Total             Rate
  Current Entrys                       649
  searches                     18421725761         1317.0/s
  inserts                        156762508           11.2/s
  removals                       156761859           11.2/s
Counters
  match                          473002784           33.8/s
  bad-offset                             0            0.0/s
  fragment                            2729            0.0/s
  short                                107            0.0/s
  normalize                           1685            0.0/s
  memory                               101            0.0/s
  bad-timestamp                          0            0.0/s
  congestion                             0            0.0/s
  ip-option                         152301            0.0/s
  proto-cksum                          108            0.0/s
  state-mismatch                     24393            0.0/s
  state-insert                          92            0.0/s
  state-limit                            0            0.0/s
  src-limit                              0            0.0/s
  synproxy                               0            0.0/s
`,
			err: errMissingData("current entries"),
		},
		// 2: bad numbers should raise an error
		pfTestCase{TestInput: `Status: Enabled for 0 days 00:26:05           Debug: Urgent

State Table                          Total             Rate
  current entries                      -23               
  searches                           11325            7.2/s
  inserts                                5            0.0/s
  removals                               3            0.0/s
Counters
  match                              11226            7.2/s
  bad-offset                             0            0.0/s
  fragment                               0            0.0/s
  short                                  0            0.0/s
  normalize                              0            0.0/s
  memory                                 0            0.0/s
  bad-timestamp                          0            0.0/s
  congestion                             0            0.0/s
  ip-option                              0            0.0/s
  proto-cksum                            0            0.0/s
  state-mismatch                         0            0.0/s
  state-insert                           0            0.0/s
  state-limit                            0            0.0/s
  src-limit                              0            0.0/s
  synproxy                               0            0.0/s
`,
			err: errMissingData("current entries"),
		},
		pfTestCase{TestInput: `Status: Enabled for 0 days 00:26:05           Debug: Urgent

State Table                          Total             Rate
  current entries                        2               
  searches                           11325            7.2/s
  inserts                                5            0.0/s
  removals                               3            0.0/s
Counters
  match                              11226            7.2/s
  bad-offset                             0            0.0/s
  fragment                               0            0.0/s
  short                                  0            0.0/s
  normalize                              0            0.0/s
  memory                                 0            0.0/s
  bad-timestamp                          0            0.0/s
  congestion                             0            0.0/s
  ip-option                              0            0.0/s
  proto-cksum                            0            0.0/s
  state-mismatch                         0            0.0/s
  state-insert                           0            0.0/s
  state-limit                            0            0.0/s
  src-limit                              0            0.0/s
  synproxy                               0            0.0/s
`,
			measurements: []measurementResult{
				measurementResult{
					fields: map[string]interface{}{
						"entries":  uint32(2),
						"searches": uint64(11325),
						"inserts":  uint64(5),
						"removals": uint64(3)},
					tags: map[string]string{},
				},
			},
		},
		pfTestCase{TestInput: `Status: Enabled for 161 days 21:24:45         Debug: Urgent

Interface Stats for re1               IPv4             IPv6
  Bytes In                   2585823744614    1059233657221
  Bytes Out                  1227266932673    3274698578875
  Packets In
    Passed                      2289953086       1945437219
    Blocked                      392835739            48609
  Packets Out
    Passed                      1649146326       2605569054
    Blocked                            107                0

State Table                          Total             Rate
  current entries                      649
  searches                     18421725761         1317.0/s
  inserts                        156762508           11.2/s
  removals                       156761859           11.2/s
Counters
  match                          473002784           33.8/s
  bad-offset                             0            0.0/s
  fragment                            2729            0.0/s
  short                                107            0.0/s
  normalize                           1685            0.0/s
  memory                               101            0.0/s
  bad-timestamp                          0            0.0/s
  congestion                             0            0.0/s
  ip-option                         152301            0.0/s
  proto-cksum                          108            0.0/s
  state-mismatch                     24393            0.0/s
  state-insert                          92            0.0/s
  state-limit                            0            0.0/s
  src-limit                              0            0.0/s
  synproxy                               0            0.0/s
`,
			measurements: []measurementResult{
				measurementResult{
					fields: map[string]interface{}{
						"entries":  uint32(649),
						"searches": uint64(18421725761),
						"inserts":  uint64(156762508),
						"removals": uint64(156761859)},
					tags: map[string]string{},
				},
			},
		},
	}

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			log.Printf("running #%d\n", i)
			pf := &PF{
				infoFunc: func() (string, error) {
					return tt.TestInput, nil
				},
			}
			acc := new(testutil.Accumulator)
			err := acc.GatherError(pf.Gather)
			if !reflect.DeepEqual(tt.err, err) {
				t.Errorf("%d: expected error '%#v' got '%#v'", i, tt.err, err)
			}
			n := 0
			for j, v := range tt.measurements {
				if len(acc.Metrics) < n+1 {
					t.Errorf("%d: expected at least %d values got %d", i, n+1, len(acc.Metrics))
					break
				}
				m := acc.Metrics[n]
				if !reflect.DeepEqual(m.Measurement, measurement) {
					t.Errorf("%d %d: expected measurement '%#v' got '%#v'\n", i, j, measurement, m.Measurement)
				}
				if !reflect.DeepEqual(m.Tags, v.tags) {
					t.Errorf("%d %d: expected tags\n%#v got\n%#v\n", i, j, v.tags, m.Tags)
				}
				if !reflect.DeepEqual(m.Fields, v.fields) {
					t.Errorf("%d %d: expected fields\n%#v got\n%#v\n", i, j, v.fields, m.Fields)
				}
				n++
			}
		})
	}
}
