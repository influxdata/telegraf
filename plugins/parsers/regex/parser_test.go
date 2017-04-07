package regex_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	oneMatch                   = "Lookingfor  I don't care about other thing in the string"
	twoMatch                   = "firstMatch I don't care about other thing\n secound match is really good for me"
	noMatch                    = "Not matches "
	extractValueTwo            = "a=1,don'tcare=2,somethingelse=4"
	extractValuethreeTwoMetric = "inMetric1=1,don'tcare=2,somethingelse=4,a=3\n inMetric2=2"
)

func TestOneMatch(t *testing.T) {
	parser := REGEXParser{
		RegexEXPRList: map[string][]string{
			"m1": []string{"Lookingfor"},
		},
	}
	metrics, err := parser.Parse([]byte(oneMatch))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "m1", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"Lookingfor": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}
func TestTwoMatch(t *testing.T) {
	parser := REGEXParser{
		RegexEXPRList: map[string][]string{
			"m1": []string{"firstMatch", "secound match"},
		},
	}
	metrics, err := parser.Parse([]byte(twoMatch))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "m1", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"firstMatch":    float64(1),
		"secound match": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}
func TestNoMatch(t *testing.T) {
	parser := REGEXParser{
		RegexEXPRList: map[string][]string{
			"m1": []string{"something"},
		},
	}
	metrics, err := parser.Parse([]byte(noMatch))
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}
func TestExtractValueOne(t *testing.T) {
	parser := REGEXParser{
		RegexEXPRList: map[string][]string{
			"m1": []string{"(a)=([0-9]+)", "(somethingelse)=([0-9]+)"},
		},
	}
	metrics, err := parser.Parse([]byte(extractValueTwo))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "m1", metrics[0].Name())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
	assert.Equal(t, map[string]interface{}{
		"a":             float64(1),
		"somethingelse": float64(4),
	}, metrics[0].Fields())
}
func TestExtractValueInTwoMetric(t *testing.T) {
	parser := REGEXParser{
		RegexEXPRList: map[string][]string{
			"m1": []string{"(inMetric1)=([0-9]+)"},
			"m2": []string{"(inMetric2)=([0-9]+)"},
		},
	}
	metrics, err := parser.Parse([]byte(extractValuethreeTwoMetric))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "m1", metrics[0].Name())
	assert.Equal(t, "m2", metrics[1].Name())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
	assert.Equal(t, map[string]string{}, metrics[1].Tags())
	assert.Equal(t, map[string]interface{}{
		"inMetric1": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]interface{}{
		"inMetric2": float64(2),
	}, metrics[1].Fields())

}
