package statsd

func (s *Statsd) timingsC(percentileLimit int) (chan struct{}, chan *metric, chan map[string]cachedtimings) {
	tC := make(chan map[string]cachedtimings)
	mC := make(chan *metric)
	tCReset := make(chan struct{})
	timings := make(map[string]cachedtimings)

	go func() {
		for {
			timingsCopy := make(map[string]cachedtimings, len(timings))
			for key := range timings {

				// Deep copy the maps
				fieldsCopy := make(map[string]RunningStats, len(timings[key].fields))
				for k := range timings[key].fields {
					fieldsCopy[k] = timings[key].fields[k]
				}

				tagsCopy := make(map[string]string, len(timings[key].tags))
				for k := range timings[key].tags {
					tagsCopy[k] = timings[key].tags[k]
				}

				// Finally assign new copy of cachedtimings to the bucket
				timingsCopy[key] = cachedtimings{
					name:   timings[key].name,
					fields: fieldsCopy,
					tags:   tagsCopy,
				}
			}
			select {
			case <-tCReset:
				for key := range timings {
					timings[key] = cachedtimings{name: "", fields: nil, tags: nil}
				}
				timings = make(map[string]cachedtimings)

				for key := range timingsCopy {
					timingsCopy[key] = cachedtimings{name: "", fields: nil, tags: nil}
				}
				timingsCopy = make(map[string]cachedtimings)
			case m, ok := <-mC:
				if !ok {
					return
				}
				// Check if the measurement exists
				cached, ok := timings[m.hash]
				if !ok {
					cached = cachedtimings{
						name:   m.name,
						fields: make(map[string]RunningStats),
						tags:   m.tags,
					}
				}
				// Check if the field exist If we've not enabled multiple fields per timer
				// this will be the default field name, eg. "value"
				field, ok := cached.fields[m.field]
				if !ok {
					field = RunningStats{
						PercLimit: percentileLimit,
					}
				}
				if m.samplerate > 0 {
					for i := 0; i < int(1.0/m.samplerate); i++ {
						field.AddValue(m.floatvalue)
					}
				} else {
					field.AddValue(m.floatvalue)
				}
				cached.fields[m.field] = field
				timings[m.hash] = cached
			case tC <- timingsCopy:
			}
		}
	}()
	return tCReset, mC, tC
}

func (s *Statsd) countersC() (chan struct{}, chan *metric, chan map[string]cachedcounter) {
	cC := make(chan map[string]cachedcounter)
	mC := make(chan *metric)
	cCReset := make(chan struct{})
	counters := make(map[string]cachedcounter)

	go func() {
		for {
			countersCopy := make(map[string]cachedcounter, len(counters))
			for key := range counters {

				// Deep copy the maps
				fieldsCopy := make(map[string]interface{}, len(counters[key].fields))
				for k := range counters[key].fields {
					fieldsCopy[k] = counters[key].fields[k]
				}

				tagsCopy := make(map[string]string, len(counters[key].tags))
				for k := range counters[key].tags {
					tagsCopy[k] = counters[key].tags[k]
				}

				// Finally assign new copy of cachedcounter to the bucket
				countersCopy[key] = cachedcounter{
					name:   counters[key].name,
					fields: fieldsCopy,
					tags:   tagsCopy,
				}
			}
			select {
			case <-cCReset:
				for key := range counters {
					counters[key] = cachedcounter{name: "", fields: nil, tags: nil}
				}
				counters = make(map[string]cachedcounter)

				for key := range countersCopy {
					countersCopy[key] = cachedcounter{name: "", fields: nil, tags: nil}
				}
				countersCopy = make(map[string]cachedcounter)
			case m, ok := <-mC:
				if !ok {
					return
				}

				// check if the measurement exists
				_, ok = counters[m.hash]
				if !ok {
					counters[m.hash] = cachedcounter{
						name:   m.name,
						fields: make(map[string]interface{}),
						tags:   m.tags,
					}
				}
				// check if the field exists
				_, ok = counters[m.hash].fields[m.field]
				if !ok {
					counters[m.hash].fields[m.field] = int64(0)
				}
				counters[m.hash].fields[m.field] = counters[m.hash].fields[m.field].(int64) + m.intvalue
			case cC <- countersCopy:
			}
		}
	}()
	return cCReset, mC, cC
}

func (s *Statsd) gaugesC() (chan struct{}, chan *metric, chan map[string]cachedgauge) {
	gC := make(chan map[string]cachedgauge)
	mC := make(chan *metric)
	gCReset := make(chan struct{})
	gauges := make(map[string]cachedgauge)

	go func() {
		for {
			gaugesCopy := make(map[string]cachedgauge, len(gauges))
			for key := range gauges {

				// Deep copy the maps
				fieldsCopy := make(map[string]interface{}, len(gauges[key].fields))
				for k := range gauges[key].fields {
					fieldsCopy[k] = gauges[key].fields[k]
				}

				tagsCopy := make(map[string]string, len(gauges[key].tags))
				for k := range gauges[key].tags {
					tagsCopy[k] = gauges[key].tags[k]
				}

				// Finally assign new copy of cachedgauge to the bucket
				gaugesCopy[key] = cachedgauge{
					name:   gauges[key].name,
					fields: fieldsCopy,
					tags:   tagsCopy,
				}
			}
			select {
			case <-gCReset:
				for key := range gauges {
					gauges[key] = cachedgauge{name: "", fields: nil, tags: nil}
				}
				gauges = make(map[string]cachedgauge)

				for key := range gaugesCopy {
					gaugesCopy[key] = cachedgauge{name: "", fields: nil, tags: nil}
				}
				gaugesCopy = make(map[string]cachedgauge)
			case m, ok := <-mC:
				if !ok {
					return
				}
				// check if the measurement exists
				_, ok = gauges[m.hash]
				if !ok {
					gauges[m.hash] = cachedgauge{
						name:   m.name,
						fields: make(map[string]interface{}),
						tags:   m.tags,
					}
				}
				// check if the field exists
				_, ok = gauges[m.hash].fields[m.field]
				if !ok {
					gauges[m.hash].fields[m.field] = float64(0)
				}
				if m.additive {
					gauges[m.hash].fields[m.field] =
						gauges[m.hash].fields[m.field].(float64) + m.floatvalue
				} else {
					gauges[m.hash].fields[m.field] = m.floatvalue
				}
			case gC <- gaugesCopy:
			}
		}
	}()
	return gCReset, mC, gC
}

func (s *Statsd) setsC() (chan struct{}, chan *metric, chan map[string]cachedset) {
	sC := make(chan map[string]cachedset)
	mC := make(chan *metric)
	sCReset := make(chan struct{})
	sets := make(map[string]cachedset)

	go func() {
		for {
			setsCopy := make(map[string]cachedset, len(sets))
			for key := range sets {

				// Deep copy the maps
				fieldsCopy := make(map[string]map[string]bool, len(sets[key].fields))
				for k := range sets[key].fields {
					fieldsBoolCopy := make(map[string]bool, len(sets[key].fields[k]))
					for iK := range sets[key].fields[k] {
						fieldsBoolCopy[iK] = sets[key].fields[k][iK]
					}
					fieldsCopy[k] = fieldsBoolCopy
				}

				tagsCopy := make(map[string]string, len(sets[key].tags))
				for k := range sets[key].tags {
					tagsCopy[k] = sets[key].tags[k]
				}

				// Finally assign new copy of cachedset to the bucket
				setsCopy[key] = cachedset{
					name:   sets[key].name,
					fields: fieldsCopy,
					tags:   tagsCopy,
				}
			}
			select {
			case <-sCReset:
				for key := range sets {
					sets[key] = cachedset{name: "", fields: nil, tags: nil}
				}
				sets = make(map[string]cachedset)

				for key := range setsCopy {
					setsCopy[key] = cachedset{name: "", fields: nil, tags: nil}
				}
				setsCopy = make(map[string]cachedset)
			case m, ok := <-mC:
				if !ok {
					return
				}
				// check if the measurement exists
				_, ok = sets[m.hash]
				if !ok {
					sets[m.hash] = cachedset{
						name:   m.name,
						fields: make(map[string]map[string]bool),
						tags:   m.tags,
					}
				}
				// check if the field exists
				_, ok = sets[m.hash].fields[m.field]
				if !ok {
					sets[m.hash].fields[m.field] = make(map[string]bool)
				}
				sets[m.hash].fields[m.field][m.strvalue] = true
			case sC <- setsCopy:
			}
		}
	}()
	return sCReset, mC, sC
}
