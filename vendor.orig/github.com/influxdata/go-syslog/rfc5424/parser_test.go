package rfc5424

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserParse(t *testing.T) {
	p := NewParser()
	for _, tc := range testCases {
		tc := tc
		t.Run(rxpad(string(tc.input), 50), func(t *testing.T) {
			t.Parallel()

			bestEffort := true
			message, merr := p.Parse(tc.input, nil)
			partial, perr := p.Parse(tc.input, &bestEffort)

			if !tc.valid {
				assert.Nil(t, message)
				assert.Error(t, merr)
				assert.EqualError(t, merr, tc.errorString)

				assert.Equal(t, tc.partialValue, partial)
				assert.EqualError(t, perr, tc.errorString)
			}
			if tc.valid {
				assert.Nil(t, merr)
				assert.NotEmpty(t, message)
				assert.Equal(t, message, partial)
				assert.Equal(t, merr, perr)
			}

			assert.Equal(t, tc.value, message)
		})
	}
}
