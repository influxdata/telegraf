package netflow

import (
	"encoding/csv"
	"fmt"
	"os"
)

var funcMapping = map[string]decoderFunc{
	"bool":    decodeBool,
	"int":     decodeInt,
	"uint":    decodeUint,
	"float32": decodeFloat32,
	"float64": decodeFloat64,
	"hex":     decodeHex,
	"string":  decodeString,
	"mac":     decodeMAC,
	"ip":      decodeIP,
	"proto":   decodeL4Proto,
}

func loadMapping(filename string) (map[string]fieldMapping, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening %q failed: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.Comment = '#'
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading csv failed: %w", err)
	}

	mappings := make(map[string]fieldMapping, len(records))
	for _, r := range records {
		id, name, dtype := r[0], r[1], r[2]
		fun, found := funcMapping[dtype]
		if !found {
			return nil, fmt.Errorf("unknown data-type %q for id %q", dtype, id)
		}
		mappings[id] = fieldMapping{name, fun}
	}

	return mappings, nil
}
