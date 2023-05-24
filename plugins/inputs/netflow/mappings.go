package netflow

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

var funcMapping = map[string]decoderFunc{
	"uint":   decodeUint,
	"hex":    decodeHex,
	"string": decodeString,
	"ip":     decodeIP,
	"proto":  decodeL4Proto,
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

func loadNumericMapping(filename string) (map[uint16]fieldMapping, error) {
	raw, err := loadMapping(filename)
	if err != nil {
		return nil, err
	}

	mappings := make(map[uint16]fieldMapping, len(raw))
	for key, fm := range raw {
		id, err := strconv.ParseUint(key, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("parsing ID %q failed: %w", key, err)
		}
		mappings[uint16(id)] = fm
	}

	return mappings, nil
}
