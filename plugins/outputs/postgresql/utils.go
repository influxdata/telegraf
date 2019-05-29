package postgresql

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/jackc/pgx"
)

func buildJsonbTags(tags map[string]string) ([]byte, error) {
	js := make(map[string]interface{})
	for column, value := range tags {
		js[column] = value
	}

	return buildJsonb(js)
}

func buildJsonb(data map[string]interface{}) ([]byte, error) {
	if len(data) > 0 {
		d, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return d, nil
	}

	return nil, nil
}

func quoteIdent(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

func quoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

func deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case bool:
		datatype = "boolean"
	case uint64:
		datatype = "int8"
	case int64:
		datatype = "int8"
	case float64:
		datatype = "float8"
	case string:
		datatype = "text"
	default:
		datatype = "text"
		log.Printf("E! Unknown datatype %T(%v)", value, value)
	}
	return datatype
}

func contains(haystack []string, needle string) bool {
	for _, key := range haystack {
		if key == needle {
			return true
		}
	}
	return false
}
