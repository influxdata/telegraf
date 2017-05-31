package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const trues = `booltest b=T
booltest b=t
booltest b=True
booltest b=TRUE
booltest b=true
`

const falses = `booltest b=F
booltest b=f
booltest b=False
booltest b=FALSE
booltest b=false
`

const withEscapes = `w\,\ eather,host=local temp=99 1465839830100400200
w\,eather,host=local temp=99 1465839830100400200
weather,location=us\,midwest temperature=82 1465839830100400200
weather,location=us-midwest temp\=rature=82 1465839830100400200
weather,location\ place=us-midwest temperature=82 1465839830100400200
weather,location=us-midwest temperature="too\"hot\"" 1465839830100400200
`

const withTimestamps = `cpu usage=99 1480595849000000000
cpu usage=99 1480595850000000000
cpu usage=99 1480595851700030000
cpu usage=99 1480595852000000300
`

const sevenMetrics = `cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
cpu,host=foo,datacenter=us-east idle=99,busy=1i,b=true,s="string"
`

const negMetrics = `weather,host=local temp=-99i,temp_float=-99.4 1465839830100400200
`

// some metrics are invalid
const someInvalid = `cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3, host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4 , usage_idle=99,usage_busy=1
cpu 1480595852000000300
cpu usage=99 1480595852foobar300
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

func TestParse(t *testing.T) {
	start := time.Now()
	metrics, err := Parse([]byte(sevenMetrics))
	assert.NoError(t, err)
	assert.Len(t, metrics, 7)

	// all metrics parsed together w/o a timestamp should have the same time.
	firstTime := metrics[0].Time()
	for _, m := range metrics {
		assert.Equal(t,
			map[string]interface{}{
				"idle": float64(99),
				"busy": int64(1),
				"b":    true,
				"s":    "string",
			},
			m.Fields(),
		)
		assert.Equal(t,
			map[string]string{
				"host":       "foo",
				"datacenter": "us-east",
			},
			m.Tags(),
		)
		assert.True(t, m.Time().After(start))
		assert.True(t, m.Time().Equal(firstTime))
	}
}

func TestParseNegNumbers(t *testing.T) {
	metrics, err := Parse([]byte(negMetrics))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)

	assert.Equal(t,
		map[string]interface{}{
			"temp":       int64(-99),
			"temp_float": float64(-99.4),
		},
		metrics[0].Fields(),
	)
	assert.Equal(t,
		map[string]string{
			"host": "local",
		},
		metrics[0].Tags(),
	)
}

func TestParseErrors(t *testing.T) {
	start := time.Now()
	metrics, err := Parse([]byte(someInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	// all metrics parsed together w/o a timestamp should have the same time.
	firstTime := metrics[0].Time()
	for _, m := range metrics {
		assert.Equal(t,
			map[string]interface{}{
				"usage_idle": float64(99),
				"usage_busy": float64(1),
			},
			m.Fields(),
		)
		assert.Equal(t,
			map[string]string{
				"host":       "foo",
				"datacenter": "us-east",
			},
			m.Tags(),
		)
		assert.True(t, m.Time().After(start))
		assert.True(t, m.Time().Equal(firstTime))
	}
}

func TestParseWithTimestamps(t *testing.T) {
	metrics, err := Parse([]byte(withTimestamps))
	assert.NoError(t, err)
	assert.Len(t, metrics, 4)

	expectedTimestamps := []time.Time{
		time.Unix(0, 1480595849000000000),
		time.Unix(0, 1480595850000000000),
		time.Unix(0, 1480595851700030000),
		time.Unix(0, 1480595852000000300),
	}

	// all metrics parsed together w/o a timestamp should have the same time.
	for i, m := range metrics {
		assert.Equal(t,
			map[string]interface{}{
				"usage": float64(99),
			},
			m.Fields(),
		)
		assert.True(t, m.Time().Equal(expectedTimestamps[i]))
	}
}

func TestParseEscapes(t *testing.T) {
	metrics, err := Parse([]byte(withEscapes))
	assert.NoError(t, err)
	assert.Len(t, metrics, 6)

	tests := []struct {
		name   string
		fields map[string]interface{}
		tags   map[string]string
	}{
		{
			name:   `w, eather`,
			fields: map[string]interface{}{"temp": float64(99)},
			tags:   map[string]string{"host": "local"},
		},
		{
			name:   `w,eather`,
			fields: map[string]interface{}{"temp": float64(99)},
			tags:   map[string]string{"host": "local"},
		},
		{
			name:   `weather`,
			fields: map[string]interface{}{"temperature": float64(82)},
			tags:   map[string]string{"location": `us,midwest`},
		},
		{
			name:   `weather`,
			fields: map[string]interface{}{`temp=rature`: float64(82)},
			tags:   map[string]string{"location": `us-midwest`},
		},
		{
			name:   `weather`,
			fields: map[string]interface{}{"temperature": float64(82)},
			tags:   map[string]string{`location place`: `us-midwest`},
		},
		{
			name:   `weather`,
			fields: map[string]interface{}{`temperature`: `too"hot"`},
			tags:   map[string]string{"location": `us-midwest`},
		},
	}

	for i, test := range tests {
		assert.Equal(t, test.name, metrics[i].Name())
		assert.Equal(t, test.fields, metrics[i].Fields())
		assert.Equal(t, test.tags, metrics[i].Tags())
	}
}

func TestParseTrueBooleans(t *testing.T) {
	metrics, err := Parse([]byte(trues))
	assert.NoError(t, err)
	assert.Len(t, metrics, 5)

	for _, metric := range metrics {
		assert.Equal(t, "booltest", metric.Name())
		assert.Equal(t, true, metric.Fields()["b"])
	}
}

func TestParseFalseBooleans(t *testing.T) {
	metrics, err := Parse([]byte(falses))
	assert.NoError(t, err)
	assert.Len(t, metrics, 5)

	for _, metric := range metrics {
		assert.Equal(t, "booltest", metric.Name())
		assert.Equal(t, false, metric.Fields()["b"])
	}
}

func TestParsePointBadNumber(t *testing.T) {
	for _, tt := range []string{
		"cpu v=- ",
		"cpu v=-i ",
		"cpu v=-. ",
		"cpu v=. ",
		"cpu v=1.0i ",
		"cpu v=1ii ",
		"cpu v=1a ",
		"cpu v=-e-e-e ",
		"cpu v=42+3 ",
		"cpu v= ",
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParseTagsMissingParts(t *testing.T) {
	for _, tt := range []string{
		`cpu,host`,
		`cpu,host,`,
		`cpu,host=`,
		`cpu,f=oo=bar value=1`,
		`cpu,host value=1i`,
		`cpu,host=serverA,region value=1i`,
		`cpu,host=serverA,region= value=1i`,
		`cpu,host=serverA,region=,zone=us-west value=1i`,
		`cpu, value=1`,
		`cpu, ,,`,
		`cpu,,,`,
		`cpu,host=serverA,=us-east value=1i`,
		`cpu,host=serverAa\,,=us-east value=1i`,
		`cpu,host=serverA\,,=us-east value=1i`,
		`cpu, =serverA value=1i`,
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParsePointWhitespace(t *testing.T) {
	for _, tt := range []string{
		`cpu    value=1.0 1257894000000000000`,
		`cpu value=1.0     1257894000000000000`,
		`cpu      value=1.0     1257894000000000000`,
		`cpu value=1.0 1257894000000000000   `,
	} {
		m, err := Parse([]byte(tt + "\n"))
		assert.NoError(t, err, tt)
		assert.Equal(t, "cpu", m[0].Name())
		assert.Equal(t, map[string]interface{}{"value": float64(1)}, m[0].Fields())
	}
}

func TestParsePointInvalidFields(t *testing.T) {
	for _, tt := range []string{
		"test,foo=bar a=101,=value",
		"test,foo=bar =value",
		"test,foo=bar a=101,key=",
		"test,foo=bar key=",
		`test,foo=bar a=101,b="foo`,
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParsePointNoFields(t *testing.T) {
	for _, tt := range []string{
		"cpu_load_short,host=server01,region=us-west",
		"very_long_measurement_name",
		"cpu,host==",
		"============",
		"cpu",
		"cpu\n\n\n\n\n\n\n",
		"                ",
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

// a b=1 << this is the shortest possible metric
// any shorter is just ignored
func TestParseBufTooShort(t *testing.T) {
	for _, tt := range []string{
		"",
		"a",
		"a ",
		"a b=",
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParseInvalidBooleans(t *testing.T) {
	for _, tt := range []string{
		"test b=tru",
		"test b=fals",
		"test b=faLse",
		"test q=foo",
		"test b=lambchops",
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParseInvalidNumbers(t *testing.T) {
	for _, tt := range []string{
		"test b=-",
		"test b=1.1.1",
		"test b=nan",
		"test b=9i10",
		"test b=9999999999999999999i",
	} {
		_, err := Parse([]byte(tt + "\n"))
		assert.Error(t, err, tt)
	}
}

func TestParseNegativeTimestamps(t *testing.T) {
	for _, tt := range []string{
		"test foo=101 -1257894000000000000",
	} {
		metrics, err := Parse([]byte(tt + "\n"))
		assert.NoError(t, err, tt)
		assert.True(t, metrics[0].Time().Equal(time.Unix(0, -1257894000000000000)))
	}
}

func TestParsePrecision(t *testing.T) {
	for _, tt := range []struct {
		line      string
		precision string
		expected  int64
	}{
		{"test v=42 1491847420", "s", 1491847420000000000},
		{"test v=42 1491847420123", "ms", 1491847420123000000},
		{"test v=42 1491847420123456", "u", 1491847420123456000},
		{"test v=42 1491847420123456789", "ns", 1491847420123456789},

		{"test v=42 1491847420123456789", "1s", 1491847420123456789},
		{"test v=42 1491847420123456789", "asdf", 1491847420123456789},
	} {
		metrics, err := ParseWithDefaultTimePrecision(
			[]byte(tt.line+"\n"), time.Now(), tt.precision)
		assert.NoError(t, err, tt)
		assert.Equal(t, tt.expected, metrics[0].UnixNano())
	}
}

func TestParseMaxKeyLength(t *testing.T) {
	key := ""
	for {
		if len(key) > MaxKeyLength {
			break
		}
		key += "test"
	}

	_, err := Parse([]byte(key + " value=1\n"))
	assert.Error(t, err)
}
