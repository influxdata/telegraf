package basicstats

import (
//	"log"
	"math"

	"github.com/influxdata/telegraf"
)

type CoStats struct {

	cache    *map[string]map[string]map[string]map[string]map[uint64]map[uint64]float64

	m *BasicStats
}

type costat struct {
	Metrics []metrics `toml:"metrics"`
}

type metrics struct {
	Name  string	`toml:"name"`
	Field  string	`toml:"field"`
}

func NewCoStats(m *BasicStats) *CoStats {
	ee := &CoStats{m: m}
	ee.Reset()
	return ee
}

func (e *CoStats) Add(in telegraf.Metric, id1 uint64, field1 string) {

//	log.Printf("Costats Add %d %s", id1, field1);

	//https://en.m.wikipedia.org/wiki/Algorithms_for_calculating_variance (for covariance online algorithm)

	// coStats products update
	if(e.cache == nil) {
		e.initCache()
	}
	name1 :=in.Name()
	m := (*(e.m))
	coCache:= (*e.cache)
	if _, ok := coCache[name1]; !ok {
		return
	}
	if _, ok := coCache[name1][field1]; !ok {
		return
	}
	delta1 := m.cache[id1].fields[field1].delta
	for name2 := range coCache[name1][field1] {
		for field2 := range coCache[name1][field1][name2] {
			if _, ok := coCache[name1][field1][name2][field2][id1]; !ok {
				coCache[name1][field1][name2][field2][id1] = make(map[uint64]float64)
				for id2 := range coCache[name2][field2][name1][field1] {
					if _, ok := m.cache[id2].fields[field2]; ok {
						delta2 := m.cache[id2].fields[field2].delta
						tmpProduct := (delta1 * delta2)
						coCache[name1][field1][name2][field2][id1][id2] = tmpProduct
					}
					coCache[name2][field2][name1][field1][id2][id1] = 0.0
				}
			}
			for id2, product := range coCache[name1][field1][name2][field2][id1] {
				if _, ok := m.cache[id2].fields[field2]; ok {
					delta2 := m.cache[id2].fields[field2].delta
					tmpProduct := product + (delta1 * delta2)
					coCache[name1][field1][name2][field2][id1][id2] = tmpProduct
				}
			}
		}
	}
}


func (e *CoStats) Push(
	id1 uint64, field1 string, fieldsOutput map[string]interface{}) {

//	log.Printf("Push stats %d %s %+v", id1, field1, fieldsOutput);

	m := (*(e.m))
	v := m.cache[id1].fields[field1]

	variance := v.M2 / (v.count - 1)

	name1 := m.cache[id1].name
	coCache:= (*e.cache)
	if _, ok := coCache[name1]; !ok {
		return
	}
	if _, ok := coCache[name1][field1]; !ok {
		return
	}
//	log.Printf("Push costats coCache %+v", coCache[name1][field1]);
	for name2 := range coCache[name1][field1] {
		for field2 := range coCache[name1][field1][name2] {
			if _, ok := coCache[name1][field1][name2][field2][id1]; ok {
				for id2, product := range coCache[name1][field1][name2][field2][id1] {
					tagString2 :=""
					for tagK, tagV := range m.cache[id2].tags {
						tagString2 +="/"+tagK+":"+tagV
					}
					tagString1 :=""
					for tagK, tagV := range m.cache[id1].tags {
						tagString1 +="/"+tagK+":"+tagV
					}
					variance2 := m.cache[id2].fields[field2].M2 / (m.cache[id2].fields[field2].count - 1)
					covariance := product/(v.count-1)
					fieldsOutput["covariance["+name1+"/"+field1+tagString1+"]["+name2+"/"+field2+tagString2+"]"] = covariance
					fieldsOutput["correlation["+name1+"/"+field1+tagString1+"]["+name2+"/"+field2+tagString2+"]"] = covariance/(math.Sqrt(variance)*math.Sqrt(variance2))
				}
			}
		}
	}

}


func (e *CoStats) initCache() {

	parsed := make(map[string]map[string]map[string]map[string]map[uint64]map[uint64]float64)

//	log.Printf("Stats %+v", (*(e.m)).Stats);

	if (*(e.m)).CoStatsConfig == nil {
//		log.Printf("coCache '%v'", parsed);
		e.cache = &parsed
		return
	}

//	log.Printf("CoStats %+v", (*(e.m)).CoStatsConfig);

	for _, costat := range (*(e.m)).CoStatsConfig {

		if len(costat.Metrics) !=2 {
				continue
		}
		metric1 := costat.Metrics[0];
		metric2 := costat.Metrics[1];
		name1 := metric1.Name
		field1 := metric1.Field
		name2 := metric2.Name
		field2 := metric2.Field
		if name1 == "" || field1 == "" || name2 == "" || field2 == "" {
			continue
		}
		if _, ok := parsed[name1]; !ok {
			parsed[name1] = make(map[string]map[string]map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name1][field1]; !ok {
			parsed[name1][field1] = make(map[string]map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name1][field1][name2]; !ok {
			parsed[name1][field1][name2] = make(map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name1][field1][name2][field2]; !ok {
			parsed[name1][field1][name2][field2] = make(map[uint64]map[uint64]float64)
		}

		if _, ok := parsed[name2]; !ok {
			parsed[name2] = make(map[string]map[string]map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name2][field2]; !ok {
			parsed[name2][field2] = make(map[string]map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name2][field2][name1]; !ok {
			parsed[name2][field2][name1] = make(map[string]map[uint64]map[uint64]float64)
		}
		if _, ok := parsed[name2][field2][name1][field1]; !ok {
			parsed[name2][field2][name1][field1] = make(map[uint64]map[uint64]float64)
		}
	}

	e.cache = &parsed
}

func (e *CoStats) Reset() {
	e.cache = nil
}
