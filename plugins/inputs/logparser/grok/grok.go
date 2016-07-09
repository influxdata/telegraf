package grok

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vjeantet/grok"

	"github.com/influxdata/telegraf"
)

var timeFormats = map[string]string{
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
	"ts-epoch":       "EPOCH",
	"ts-epochnano":   "EPOCH_NANO",
}

const (
	INT      = "int"
	TAG      = "tag"
	FLOAT    = "float"
	STRING   = "string"
	DURATION = "duration"
	DROP     = "drop"
)

var (
	// matches named captures that contain a type.
	//   ie,
	//     %{NUMBER:bytes:int}
	//     %{IPORHOST:clientip:tag}
	//     %{HTTPDATE:ts1:ts-http}
	//     %{HTTPDATE:ts2:ts-"02 Jan 06 15:04"}
	typedRe = regexp.MustCompile(`%{\w+:(\w+):(ts-".+"|t?s?-?\w+)}`)
	// matches a plain pattern name. ie, %{NUMBER}
	patternOnlyRe = regexp.MustCompile(`%{(\w+)}`)
)

type Parser struct {
	Patterns           []string
	CustomPatterns     string
	CustomPatternFiles []string

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
	// patterns is a map of all of the parsed patterns from CustomPatterns
	// and CustomPatternFiles.
	//   ie, {
	//          "DURATION":      "%{NUMBER}[nuµm]?s"
	//          "RESPONSE_CODE": "%{NUMBER:rc:tag}"
	//       }
	patterns map[string]string

	g        *grok.Grok
	tsModder *tsModder
}

func (p *Parser) Compile() error {
	p.typeMap = make(map[string]map[string]string)
	p.tsMap = make(map[string]map[string]string)
	p.patterns = make(map[string]string)
	p.tsModder = &tsModder{}
	var err error
	p.g, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		return err
	}

	p.CustomPatterns = DEFAULT_PATTERNS + p.CustomPatterns

	if len(p.CustomPatterns) != 0 {
		scanner := bufio.NewScanner(strings.NewReader(p.CustomPatterns))
		p.addCustomPatterns(scanner)
	}

	for _, filename := range p.CustomPatternFiles {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))
		p.addCustomPatterns(scanner)
	}

	return p.compileCustomPatterns()
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	var err error
	var values map[string]string
	// the matching pattern string
	var patternName string
	for _, pattern := range p.Patterns {
		if values, err = p.g.Parse(pattern, line); err != nil {
			return nil, err
		}
		if len(values) != 0 {
			patternName = pattern
			break
		}
	}

	if len(values) == 0 {
		return nil, nil
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)
	timestamp := time.Now()
	for k, v := range values {
		if k == "" || v == "" {
			continue
		}

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
			t = STRING
		}

		switch t {
		case INT:
			iv, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Printf("ERROR parsing %s to int: %s", v, err)
			} else {
				fields[k] = iv
			}
		case FLOAT:
			fv, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Printf("ERROR parsing %s to float: %s", v, err)
			} else {
				fields[k] = fv
			}
		case DURATION:
			d, err := time.ParseDuration(v)
			if err != nil {
				log.Printf("ERROR parsing %s to duration: %s", v, err)
			} else {
				fields[k] = int64(d)
			}
		case TAG:
			tags[k] = v
		case STRING:
			fields[k] = strings.Trim(v, `"`)
		case "EPOCH":
			iv, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Printf("ERROR parsing %s to int: %s", v, err)
			} else {
				timestamp = time.Unix(iv, 0)
			}
		case "EPOCH_NANO":
			iv, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Printf("ERROR parsing %s to int: %s", v, err)
			} else {
				timestamp = time.Unix(0, iv)
			}
		case DROP:
		// goodbye!
		default:
			ts, err := time.Parse(t, v)
			if err == nil {
				timestamp = ts
			} else {
				log.Printf("ERROR parsing %s to time layout [%s]: %s", v, t, err)
			}
		}
	}

	return telegraf.NewMetric("logparser_grok", tags, fields, p.tsModder.tsMod(timestamp))
}

func (p *Parser) addCustomPatterns(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && line[0] != '#' {
			names := strings.SplitN(line, " ", 2)
			p.patterns[names[0]] = names[1]
		}
	}
}

func (p *Parser) compileCustomPatterns() error {
	var err error
	// check if the pattern contains a subpattern that is already defined
	// replace it with the subpattern for modifier inheritance.
	for i := 0; i < 2; i++ {
		for name, pattern := range p.patterns {
			subNames := patternOnlyRe.FindAllStringSubmatch(pattern, -1)
			for _, subName := range subNames {
				if subPattern, ok := p.patterns[subName[1]]; ok {
					pattern = strings.Replace(pattern, subName[0], subPattern, 1)
				}
			}
			p.patterns[name] = pattern
		}
	}

	// check if pattern contains modifiers. Parse them out if it does.
	for name, pattern := range p.patterns {
		if typedRe.MatchString(pattern) {
			// this pattern has modifiers, so parse out the modifiers
			pattern, err = p.parseTypedCaptures(name, pattern)
			if err != nil {
				return err
			}
			p.patterns[name] = pattern
		}
	}

	return p.g.AddPatternsFromMap(p.patterns)
}

// parseTypedCaptures parses the capture types, and then deletes the type from
// the line so that it is a valid "grok" pattern again.
//   ie,
//     %{NUMBER:bytes:int}      => %{NUMBER:bytes}      (stores %{NUMBER}->bytes->int)
//     %{IPORHOST:clientip:tag} => %{IPORHOST:clientip} (stores %{IPORHOST}->clientip->tag)
func (p *Parser) parseTypedCaptures(name, pattern string) (string, error) {
	matches := typedRe.FindAllStringSubmatch(pattern, -1)

	// grab the name of the capture pattern
	patternName := "%{" + name + "}"
	// create type map for this pattern
	p.typeMap[patternName] = make(map[string]string)
	p.tsMap[patternName] = make(map[string]string)

	// boolean to verify that each pattern only has a single ts- data type.
	hasTimestamp := false
	for _, match := range matches {
		// regex capture 1 is the name of the capture
		// regex capture 2 is the type of the capture
		if strings.HasPrefix(match[2], "ts-") {
			if hasTimestamp {
				return pattern, fmt.Errorf("logparser pattern compile error: "+
					"Each pattern is allowed only one named "+
					"timestamp data type. pattern: %s", pattern)
			}
			if f, ok := timeFormats[match[2]]; ok {
				p.tsMap[patternName][match[1]] = f
			} else {
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
