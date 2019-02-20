package sqlserver_extensible

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

// Query struct
type Querytest struct {
	Measurement string
	Tags        []string
	Fields      []string
	Fieldname   []string
	Result      string
}

// MapQuery type
type MapQuery map[string]Querytest

func TestSqlServer_ParseMetrics(t *testing.T) {

	var acc testutil.Accumulator

	queries := make(MapQuery)
	queries["PerformanceCounters"] = Querytest{
		Measurement: "sql_server_perf_counters",
		Tags:        []string{"server_name", "counter_name", "instance_name"},
		Fields:      []string{"cntr_value"},
		Result:      mockPerformanceCounters}

	queries["MemoryClerk"] = Querytest{
		Measurement: "sql_server_memory_clerk",
		Tags:        []string{"server_name", "counter_name"},
		Fields:      []string{"Buffer pool", "Cache (objects)", "Cache (sql plans)", "Other"},
		Result:      mockPerformanceCounters}

	var headers, mock, cols []string
	var tags = make(map[string]string)
	var fields = make(map[string]interface{})

	for _, query := range queries {
		mock = strings.Split(query.Result, "\n")
		idx := 0

		for _, line := range mock {
			if idx == 0 { // headers
				headers = strings.Split(line, ";")
				continue
			}

			// columnMap
			cols = strings.Split(line, ";")
			columnMap := make(map[string]*interface{})
			jdx := 0
			for _, column := range headers {
				*columnMap[column] = cols[jdx]
				jdx++
			}

			// measurement
			var measurementb bytes.Buffer
			var measurement string
			if query.Measurement == "" {
				measurementb.WriteString((*columnMap["measurement"]).(string))
				measurement = measurementb.String()
				delete(columnMap, "measurement")
			} else {
				measurement = query.Measurement
			}

		COLUMN:
			for col, val := range columnMap {

				// tags
				for _, tag := range query.Tags {
					if col != tag {
						continue
					}
					switch v := (*val).(type) {
					case string:
						tags[col] = v
					case []byte:
						tags[col] = string(v)
					case int64, int32, int:
						tags[col] = fmt.Sprintf("%d", v)
					default:
						log.Println("Failed to add additional tag", col)
					}
					continue COLUMN
				}

				// fields
				for _, field := range query.Fields {
					if col != field {
						continue
					}
					// default fieldname
					var fieldheader string = col

					// custom fieldname
					if len(query.Fieldname) > 0 {
						fieldheader = ""
						for _, fname := range query.Fieldname {
							_, ok := columnMap[fname]
							if ok {
								fheader := (*columnMap[fname]).(string)
								if len(fheader) > 0 {
									fieldheader += fheader + " | "
								}
							}
						}
						sz := len(fieldheader)
						if sz > 0 && fieldheader[sz-2] == '|' {
							fieldheader = fieldheader[:sz-3]
						}
					}
					if v, ok := (*val).([]byte); ok {
						fields[fieldheader] = string(v)
					} else {
						fields[fieldheader] = *val
					}
					continue COLUMN
				}
			}

			acc.AddFields(measurement, fields, tags, time.Now())

			// assert
			acc.AssertContainsTaggedFields(t, measurement, fields, tags)

			idx++
		}
	}
}

const mockPerformanceCounters string = `measurement;server_name;counter_name;instance_name;cntr_value
sql_server_perf_counters;WIN8-DEV;Page_lookups/sec;"";1754431i 
sql_server_perf_counters;WIN8-DEV;Lazy_writes/sec;"";0i 
sql_server_perf_counters;WIN8-DEV;Readahead_pages/sec;"";595i`

const mockMemoryClerk string = `measurement;server_name;type;counter_name;instance_name;Buffer\ pool;Cache\ (objects);Cache\ (sql\ plans);Other
sql_server_memory_clerk;WIN8-DEV;Memory\ breakdown\ (%);31.40;0.30;8.50;59.80
sql_server_memory_clerk;WIN8-DEV;Memory\ breakdown\ (bytes);475136.00;13991936.00;98820096.00;51871744.00`
