package grok

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vjeantet/grok"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var timeLayouts = map[string]string{
	"ts-ansic":       "Mon Jan _2 15:04:05 2006",
	"ts-unix":        "Mon Jan _2 15:04:05 MST 2006",
	"ts-ruby":        "Mon Jan 02 15:04:05 -0700 2006",
	"ts-rfc822":      "02 Jan 06 15:04 MST",
	"ts-rfc822z":     "02 Jan 06 15:04 -0700", // RFC822 with numeric zone
	"ts-rfc850":      "Monday, 02-Jan-06 15:04:05 MST",
	"ts-rfc1123":     "Mon, 02 Jan 2006 15:04:05 MST",
	"ts-rfc1123z":    "Mon, 02 Jan 2006 15:04:05 -0700", // RFC1123 with numeric zone
	"ts-rfc3339":     "2006-01-02T15:04:05Z07:00",
	"ts-rfc3339nano": "2006-01-02T15:04:05.999999999Z07:00",
	"ts-httpd":       "02/Jan/2006:15:04:05 -0700",
	// These four are not exactly "layouts", but they are special cases that
	// will get handled in the ParseLine function.
	"ts-epoch":      "EPOCH",
	"ts-epochnano":  "EPOCH_NANO",
	"ts-epochmilli": "EPOCH_MILLI",
	"ts-syslog":     "SYSLOG_TIMESTAMP",
	"ts":            "GENERIC_TIMESTAMP", // try parsing all known timestamp layouts.
}

const (
	Measurement      = "measurement"
	Int              = "int"
	Tag              = "tag"
	Float            = "float"
	String           = "string"
	Duration         = "duration"
	Drop             = "drop"
	Epoch            = "EPOCH"
	EpochMilli       = "EPOCH_MILLI"
	EpochNano        = "EPOCH_NANO"
	SyslogTimestamp  = "SYSLOG_TIMESTAMP"
	GenericTimestamp = "GENERIC_TIMESTAMP"
)

var (
	// matches named captures that contain a modifier.
	//   ie,
	//     %{NUMBER:bytes:int}
	//     %{IPORHOST:clientip:tag}
	//     %{HTTPDATE:ts1:ts-http}
	//     %{HTTPDATE:ts2:ts-"02 Jan 06 15:04"}
	modifierRe = regexp.MustCompile(`%{\w+:(\w+):(ts-".+"|t?s?-?\w+)}`)
	// matches a plain pattern name. ie, %{NUMBER}
	patternOnlyRe = regexp.MustCompile(`%{(\w+)}`)
)

// Parser is the primary struct to handle and grok-patterns defined in the config toml
type Parser struct {
	Patterns []string
	// namedPatterns is a list of internally-assigned names to the patterns
	// specified by the user in Patterns.
	// They will look like:
	//   GROK_INTERNAL_PATTERN_0, GROK_INTERNAL_PATTERN_1, etc.
	NamedPatterns      []string
	CustomPatterns     string
	CustomPatternFiles []string
	Measurement        string
	DefaultTags        map[string]string
	Log                telegraf.Logger `toml:"-"`

	// Timezone is an optional component to help render log dates to
	// your chosen zone.
	// Default: "" which renders UTC
	// Options are as follows:
	// 1. Local             -- interpret based on machine localtime
	// 2. "America/Chicago" -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	// 3. UTC               -- or blank/unspecified, will return timestamp in UTC
	Timezone string
	loc      *time.Location

	// UniqueTimestamp when set to "disable", timestamp will not incremented if there is a duplicate.
	UniqueTimestamp string

	// typeMap is a map of patterns -> capture name -> modifier,
	//   ie, {
	//          "%{TESTLOG}":
	//             {
	//                "bytes": "int",
	//                "clientip": "tag"
	//             }
	//       }
	typeMap map[string]map[string]string
	// tsMap is a map of patterns -> capture name -> timestamp layout.
	//   ie, {
	//          "%{TESTLOG}":
	//             {
	//                "httptime": "02/Jan/2006:15:04:05 -0700"
	//             }
	//       }
	tsMap map[string]map[string]string
	// patternsMap is a map of all of the parsed patterns from CustomPatterns
	// and CustomPatternFiles.
	//   ie, {
	//          "DURATION":      "%{NUMBER}[nuµm]?s"
	//          "RESPONSE_CODE": "%{NUMBER:rc:tag}"
	//       }
	patternsMap map[string]string
	// foundTsLayouts is a slice of timestamp patterns that have been found
	// in the log lines. This slice gets updated if the user uses the generic
	// 'ts' modifier for timestamps. This slice is checked first for matches,
	// so that previously-matched layouts get priority over all other timestamp
	// layouts.
	foundTsLayouts []string

	timeFunc func() time.Time
	g        *grok.Grok
	tsModder *tsModder
}

// Compile is a bound method to Parser which will process the options for our parser
func (p *Parser) Compile() error {
	p.typeMap = make(map[string]map[string]string)
	p.tsMap = make(map[string]map[string]string)
	p.patternsMap = make(map[string]string)
	p.tsModder = &tsModder{}
	var err error
	p.g, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		return err
	}

	if p.UniqueTimestamp == "" {
		p.UniqueTimestamp = "auto"
	}

	// Give Patterns fake names so that they can be treated as named
	// "custom patterns"
	p.NamedPatterns = make([]string, 0, len(p.Patterns))
	for i, pattern := range p.Patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		name := fmt.Sprintf("GROK_INTERNAL_PATTERN_%d", i)
		p.CustomPatterns += "\n" + name + " " + pattern + "\n"
		p.NamedPatterns = append(p.NamedPatterns, "%{"+name+"}")
	}

	if len(p.NamedPatterns) == 0 {
		return fmt.Errorf("pattern required")
	}

	// Combine user-supplied CustomPatterns with DEFAULT_PATTERNS and parse
	// them together as the same type of pattern.
	p.CustomPatterns = DefaultPatterns + p.CustomPatterns
	if len(p.CustomPatterns) != 0 {
		scanner := bufio.NewScanner(strings.NewReader(p.CustomPatterns))
		p.addCustomPatterns(scanner)
	}

	// Parse any custom pattern files supplied.
	for _, filename := range p.CustomPatternFiles {
		file, fileErr := os.Open(filename)
		if fileErr != nil {
			return fileErr
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))
		p.addCustomPatterns(scanner)
	}

	p.loc, err = time.LoadLocation(p.Timezone)
	if err != nil {
		p.Log.Warnf("Improper timezone supplied (%s), setting loc to UTC", p.Timezone)
		p.loc, _ = time.LoadLocation("UTC")
	}

	if p.timeFunc == nil {
		p.timeFunc = time.Now
	}

	return p.compileCustomPatterns()
}

// ParseLine is the primary function to process individual lines, returning the metrics
func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	var err error
	// values are the parsed fields from the log line
	var values map[string]string
	// the matching pattern string
	var patternName string
	for _, pattern := range p.NamedPatterns {
		if values, err = p.g.Parse(pattern, line); err != nil {
			return nil, err
		}
		if len(values) != 0 {
			patternName = pattern
			break
		}
	}

	if len(values) == 0 {
		p.Log.Debugf("Grok no match found for: %q", line)
		return nil, nil
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	//add default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	timestamp := time.Now()
	for k, v := range values {
		if k == "" || v == "" {
			continue
		}
		// t is the modifier of the field
		var t string
		// check if pattern has some modifiers
		if types, ok := p.typeMap[patternName]; ok {
			t = types[k]
		}
		// if we didn't find a modifier, check if we have a timestamp layout
		if t == "" {
			if ts, ok := p.tsMap[patternName]; ok {
				// check if the modifier is a timestamp layout
				if layout, ok := ts[k]; ok {
					t = layout
				}
			}
		}
		// if we didn't find a type OR timestamp modifier, assume string
		if t == "" {
			t = String
		}

		switch t {
		case Measurement:
			p.Measurement = v
		case Int:
			iv, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				p.Log.Errorf("Error parsing %s to int: %s", v, err)
			} else {
				fields[k] = iv
			}
		case Float:
			fv, err := strconv.ParseFloat(v, 64)
			if err != nil {
				p.Log.Errorf("Error parsing %s to float: %s", v, err)
			} else {
				fields[k] = fv
			}
		case Duration:
			d, err := time.ParseDuration(v)
			if err != nil {
				p.Log.Errorf("Error parsing %s to duration: %s", v, err)
			} else {
				fields[k] = int64(d)
			}
		case Tag:
			tags[k] = v
		case String:
			fields[k] = v
		case Epoch:
			parts := strings.SplitN(v, ".", 2)
			if len(parts) == 0 {
				p.Log.Errorf("Error parsing %s to timestamp: %s", v, err)
				break
			}

			sec, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				p.Log.Errorf("Error parsing %s to timestamp: %s", v, err)
				break
			}
			ts := time.Unix(sec, 0)

			if len(parts) == 2 {
				padded := fmt.Sprintf("%-9s", parts[1])
				nsString := strings.Replace(padded[:9], " ", "0", -1)
				nanosec, err := strconv.ParseInt(nsString, 10, 64)
				if err != nil {
					p.Log.Errorf("Error parsing %s to timestamp: %s", v, err)
					break
				}
				ts = ts.Add(time.Duration(nanosec) * time.Nanosecond)
			}
			timestamp = ts
		case EpochMilli:
			ms, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				p.Log.Errorf("Error parsing %s to int: %s", v, err)
			} else {
				timestamp = time.Unix(0, ms*int64(time.Millisecond))
			}
		case EpochNano:
			iv, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				p.Log.Errorf("Error parsing %s to int: %s", v, err)
			} else {
				timestamp = time.Unix(0, iv)
			}
		case SyslogTimestamp:
			ts, err := time.ParseInLocation(time.Stamp, v, p.loc)
			if err == nil {
				if ts.Year() == 0 {
					ts = ts.AddDate(timestamp.Year(), 0, 0)
				}
				timestamp = ts
			} else {
				p.Log.Errorf("Error parsing %s to time layout [%s]: %s", v, t, err)
			}
		case GenericTimestamp:
			var foundTs bool
			// first try timestamp layouts that we've already found
			for _, layout := range p.foundTsLayouts {
				ts, err := time.ParseInLocation(layout, v, p.loc)
				if err == nil {
					timestamp = ts
					foundTs = true
					break
				}
			}
			// if we haven't found a timestamp layout yet, try all timestamp
			// layouts.
			if !foundTs {
				for _, layout := range timeLayouts {
					ts, err := time.ParseInLocation(layout, v, p.loc)
					if err == nil {
						timestamp = ts
						foundTs = true
						p.foundTsLayouts = append(p.foundTsLayouts, layout)
						break
					}
				}
			}
			// if we still haven't found a timestamp layout, log it and we will
			// just use time.Now()
			if !foundTs {
				p.Log.Errorf("Error parsing timestamp [%s], could not find any "+
					"suitable time layouts.", v)
			}
		case Drop:
		// goodbye!
		default:
			v = strings.Replace(v, ",", ".", -1)
			ts, err := time.ParseInLocation(t, v, p.loc)
			if err == nil {
				if ts.Year() == 0 {
					ts = ts.AddDate(timestamp.Year(), 0, 0)
				}
				timestamp = ts
			} else {
				p.Log.Errorf("Error parsing %s to time layout [%s]: %s", v, t, err)
			}
		}
	}

	if p.UniqueTimestamp != "auto" {
		return metric.New(p.Measurement, tags, fields, timestamp), nil
	}

	return metric.New(p.Measurement, tags, fields, p.tsModder.tsMod(timestamp)), nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		line := scanner.Text()
		m, err := p.ParseLine(line)
		if err != nil {
			return nil, err
		}

		if m == nil {
			continue
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) addCustomPatterns(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && line[0] != '#' {
			names := strings.SplitN(line, " ", 2)
			p.patternsMap[names[0]] = names[1]
		}
	}
}

func (p *Parser) compileCustomPatterns() error {
	var err error
	// check if the pattern contains a subpattern that is already defined
	// replace it with the subpattern for modifier inheritance.
	for i := 0; i < 2; i++ {
		for name, pattern := range p.patternsMap {
			subNames := patternOnlyRe.FindAllStringSubmatch(pattern, -1)
			for _, subName := range subNames {
				if subPattern, ok := p.patternsMap[subName[1]]; ok {
					pattern = strings.Replace(pattern, subName[0], subPattern, 1)
				}
			}
			p.patternsMap[name] = pattern
		}
	}

	// check if pattern contains modifiers. Parse them out if it does.
	for name, pattern := range p.patternsMap {
		if modifierRe.MatchString(pattern) {
			// this pattern has modifiers, so parse out the modifiers
			pattern, err = p.parseTypedCaptures(name, pattern)
			if err != nil {
				return err
			}
			p.patternsMap[name] = pattern
		}
	}

	return p.g.AddPatternsFromMap(p.patternsMap)
}

// parseTypedCaptures parses the capture modifiers, and then deletes the
// modifier from the line so that it is a valid "grok" pattern again.
//   ie,
//     %{NUMBER:bytes:int}      => %{NUMBER:bytes}      (stores %{NUMBER}->bytes->int)
//     %{IPORHOST:clientip:tag} => %{IPORHOST:clientip} (stores %{IPORHOST}->clientip->tag)
func (p *Parser) parseTypedCaptures(name, pattern string) (string, error) {
	matches := modifierRe.FindAllStringSubmatch(pattern, -1)

	// grab the name of the capture pattern
	patternName := "%{" + name + "}"
	// create type map for this pattern
	p.typeMap[patternName] = make(map[string]string)
	p.tsMap[patternName] = make(map[string]string)

	// boolean to verify that each pattern only has a single ts- data type.
	hasTimestamp := false
	for _, match := range matches {
		// regex capture 1 is the name of the capture
		// regex capture 2 is the modifier of the capture
		if strings.HasPrefix(match[2], "ts") {
			if hasTimestamp {
				return pattern, fmt.Errorf("logparser pattern compile error: "+
					"Each pattern is allowed only one named "+
					"timestamp data type. pattern: %s", pattern)
			}
			if layout, ok := timeLayouts[match[2]]; ok {
				// built-in time format
				p.tsMap[patternName][match[1]] = layout
			} else {
				// custom time format
				p.tsMap[patternName][match[1]] = strings.TrimSuffix(strings.TrimPrefix(match[2], `ts-"`), `"`)
			}
			hasTimestamp = true
		} else {
			p.typeMap[patternName][match[1]] = match[2]
		}

		// the modifier is not a valid part of a "grok" pattern, so remove it
		// from the pattern.
		pattern = strings.Replace(pattern, ":"+match[2]+"}", "}", 1)
	}

	return pattern, nil
}

// tsModder is a struct for incrementing identical timestamps of log lines
// so that we don't push identical metrics that will get overwritten.
type tsModder struct {
	dupe     time.Time
	last     time.Time
	incr     time.Duration
	incrn    time.Duration
	rollover time.Duration
}

// tsMod increments the given timestamp one unit more from the previous
// duplicate timestamp.
// the increment unit is determined as the next smallest time unit below the
// most significant time unit of ts.
//   ie, if the input is at ms precision, it will increment it 1µs.
func (t *tsModder) tsMod(ts time.Time) time.Time {
	if ts.IsZero() {
		return ts
	}
	defer func() { t.last = ts }()
	// don't mod the time if we don't need to
	if t.last.IsZero() || ts.IsZero() {
		t.incrn = 0
		t.rollover = 0
		return ts
	}
	if !ts.Equal(t.last) && !ts.Equal(t.dupe) {
		t.incr = 0
		t.incrn = 0
		t.rollover = 0
		return ts
	}
	if ts.Equal(t.last) {
		t.dupe = ts
	}

	if ts.Equal(t.dupe) && t.incr == time.Duration(0) {
		tsNano := ts.UnixNano()

		d := int64(10)
		counter := 1
		for {
			a := tsNano % d
			if a > 0 {
				break
			}
			d = d * 10
			counter++
		}

		switch {
		case counter <= 6:
			t.incr = time.Nanosecond
		case counter <= 9:
			t.incr = time.Microsecond
		case counter > 9:
			t.incr = time.Millisecond
		}
	}

	t.incrn++
	if t.incrn == 999 && t.incr > time.Nanosecond {
		t.rollover = t.incr * t.incrn
		t.incrn = 1
		t.incr = t.incr / 1000
		if t.incr < time.Nanosecond {
			t.incr = time.Nanosecond
		}
	}
	return ts.Add(t.incr*t.incrn + t.rollover)
}
