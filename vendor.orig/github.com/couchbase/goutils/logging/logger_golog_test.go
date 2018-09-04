//  Copyright (c) 2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var buffer *bytes.Buffer

func setLogWriter(w io.Writer, lvl Level, fmtLogging LogEntryFormatter) {
	logger = NewLogger(w, lvl, fmtLogging)
	SetLogger(logger)
}

func TestStub(t *testing.T) {
	logger := NewLogger(os.Stdout, DEBUG, KVFORMATTER)
	SetLogger(logger)

	logger.Infof("This is a message from %s", "test")
	Infof("This is a message from %s", "test")
	logger.Infop("This is a message from ", Pair{"name", "test"}, Pair{"Queue Size", 10}, Pair{"Debug Mode", false})
	Infop("This is a message from ", Pair{"name", "test"})

	logger.Infom("This is a message from ", Map{"name": "test", "Queue Size": 10, "Debug Mode": false})
	Infom("This is a message from ", Map{"name": "test"})

	logger.Requestf(WARN, "This is a Request from %s", "test")
	Requestf(INFO, "This is a Request from %s", "test")
	logger.Requestp(DEBUG, "This is a Request from ", Pair{"name", "test"})
	Requestp(ERROR, "This is a Request from ", Pair{"name", "test"})

	logger.SetLevel(WARN)
	fmt.Printf("Log level is %s\n", logger.Level())

	logger.Requestf(WARN, "This is a Request from %s", "test")
	Requestf(INFO, "This is a Request from %s", "test")
	logger.Requestp(DEBUG, "This is a Request from ", Pair{"name", "test"})
	Requestp(ERROR, "This is a Request from ", Pair{"name", "test"})

	logger.Warnf("This is a message from %s", "test")
	Infof("This is a message from %s", "test")
	logger.Debugp("This is a message from ", Pair{"name", "test"})
	Errorp("This is a message from ", Pair{"name", "test"})

	fmt.Printf("Changing to json formatter\n")
	logger.entryFormatter = &jsonFormatter{}
	logger.SetLevel(DEBUG)

	logger.Infof("This is a message from %s", "test")
	Infof("This is a message from %s", "test")
	logger.Infop("This is a message from ", Pair{"name", "test"}, Pair{"Queue Size", 10}, Pair{"Debug Mode", false})
	Infop("This is a message from ", Pair{"name", "test"})

	logger.Infom("This is a message from ", Map{"name": "test", "Queue Size": 10, "Debug Mode": false})
	Infom("This is a message from ", Map{"name": "test"})

	logger.Requestf(WARN, "This is a Request from %s", "test")
	Requestf(INFO, "This is a Request from %s", "test")
	logger.Requestp(DEBUG, "This is a Request from ", Pair{"name", "test"})
	Requestp(ERROR, "This is a Request from ", Pair{"name", "test"})

	fmt.Printf("Changing to Text formatter\n")
	logger.entryFormatter = &textFormatter{}
	logger.SetLevel(DEBUG)

	logger.Infof("This is a message from %s", "test")
	Infof("This is a message from %s", "test")
	logger.Infop("This is a message from ", Pair{"name", "test"}, Pair{"Queue Size", 10}, Pair{"Debug Mode", false})
	Infop("This is a message from ", Pair{"name", "test"})

	logger.Infom("This is a message from ", Map{"name": "test", "Queue Size": 10, "Debug Mode": false})
	Infom("This is a message from ", Map{"name": "test"})

	logger.Requestf(WARN, "This is a Request from %s", "test")
	Requestf(INFO, "This is a Request from %s", "test")
	logger.Requestp(DEBUG, "This is a Request from ", Pair{"name", "test"})
	Requestp(ERROR, "This is a Request from ", Pair{"name", "test"})

	buffer.Reset()
	logger = NewLogger(buffer, DEBUG, KVFORMATTER)
	logger.Infof("This is a message from test in key-value format")
	if s := string(buffer.Bytes()); strings.Contains(s, "_msg=This is a message from test in key-value format") == false {
		t.Errorf("Infof() failed %v", s)
	}
	buffer.Reset()
	logger.entryFormatter = &jsonFormatter{}
	logger.Infof("This is a message from test in jason format")
	if s := string(buffer.Bytes()); strings.Contains(s, "\"_msg\":\"This is a message from test in jason format\"") == false {
		t.Errorf("Infof() failed %v", s)
	}
	buffer.Reset()
	logger.entryFormatter = &textFormatter{}
	logger.Infof("This is a message from test in text format")
	if s := string(buffer.Bytes()); strings.Contains(s, "[INFO] This is a message from test in text format") == false {
		t.Errorf("Infof() failed %v", s)
	}
}

func init() {
	buffer = bytes.NewBuffer([]byte{})
	buffer.Reset()
}

func TestLogNone(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, DEBUG, KVFORMATTER)
	Warnf("%s", "test")
	if s := string(buffer.Bytes()); strings.Contains(s, "test") == false {
		t.Errorf("Warnf() failed %v", s)
	}
	SetLevel(NONE)
	Warnf("test")
	if s := string(buffer.Bytes()); s == "" {
		t.Errorf("Warnf() failed %v", s)
	}
}
func TestLogLevelDefault(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, INFO, KVFORMATTER)
	SetLevel(INFO)
	Warnf("%s", "warn")
	Errorf("error")
	Severef("severe")
	Infof("info")
	Debugf("debug")
	Tracef("trace")
	s := string(buffer.Bytes())
	if strings.Contains(s, "warn") == false {
		t.Errorf("Warnf() failed %v", s)
	} else if strings.Contains(s, "error") == false {
		t.Errorf("Errorf() failed %v", s)
	} else if strings.Contains(s, "severe") == false {
		t.Errorf("Severef() failed %v", s)
	} else if strings.Contains(s, "info") == false {
		t.Errorf("Infof() failed %v", s)
	} else if strings.Contains(s, "debug") == true {
		t.Errorf("Debugf() failed %v", s)
	} else if strings.Contains(s, "trace") == true {
		t.Errorf("Tracef() failed %v", s)
	}
	setLogWriter(os.Stdout, INFO, KVFORMATTER)
}

func TestLogLevelInfo(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, INFO, KVFORMATTER)
	Warnf("warn")
	Infof("info")
	Debugf("debug")
	Tracef("trace")
	s := string(buffer.Bytes())
	if strings.Contains(s, "warn") == false {
		t.Errorf("Warnf() failed %v", s)
	} else if strings.Contains(s, "info") == false {
		t.Errorf("Infof() failed %v", s)
	} else if strings.Contains(s, "debug") == true {
		t.Errorf("Debugf() failed %v", s)
	} else if strings.Contains(s, "trace") == true {
		t.Errorf("Tracef() failed %v", s)
	}
	setLogWriter(os.Stdout, INFO, KVFORMATTER)
}

func TestLogLevelDebug(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, DEBUG, KVFORMATTER)
	Warnf("warn")
	Infof("info")
	Debugf("debug")
	Tracef("trace")
	s := string(buffer.Bytes())
	if strings.Contains(s, "warn") == false {
		t.Errorf("Warnf() failed %v", s)
	} else if strings.Contains(s, "info") == false {
		t.Errorf("Infof() failed %v", s)
	} else if strings.Contains(s, "debug") == false {
		t.Errorf("Debugf() failed %v", s)
	} else if strings.Contains(s, "trace") == false {
		t.Errorf("Tracef() failed %v", s)
	}
	setLogWriter(os.Stdout, INFO, KVFORMATTER)
}

func TestLogLevelTrace(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, TRACE, KVFORMATTER)
	Warnf("warn")
	Infof("info")
	Debugf("debug")
	Tracef("trace")
	s := string(buffer.Bytes())
	if strings.Contains(s, "warn") == false {
		t.Errorf("Warnf() failed %v", s)
	} else if strings.Contains(s, "info") == false {
		t.Errorf("Infof() failed %v", s)
	} else if strings.Contains(s, "debug") == true {
		t.Errorf("Debugf() failed %v", s)
	} else if strings.Contains(s, "trace") == false {
		t.Errorf("Tracef() failed %v", s)
	}
	setLogWriter(os.Stdout, INFO, KVFORMATTER)
}

func TestDefaultLog(t *testing.T) {
	buffer.Reset()
	setLogWriter(buffer, TRACE, KVFORMATTER)
	sl := logger
	sl.Warnf("warn")
	sl.Infof("info")
	sl.Debugf("debug")
	sl.Tracef("trace")
	s := string(buffer.Bytes())
	if strings.Contains(s, "warn") == false {
		t.Errorf("Warnf() failed %v", s)
	} else if strings.Contains(s, "info") == false {
		t.Errorf("Infof() failed %v", s)
	} else if strings.Contains(s, "trace") == false {
		t.Errorf("Tracef() failed %v", s)
	} else if strings.Contains(s, "debug") == true {
		t.Errorf("Debugf() failed %v", s)
	}
	setLogWriter(os.Stdout, INFO, KVFORMATTER)
}
