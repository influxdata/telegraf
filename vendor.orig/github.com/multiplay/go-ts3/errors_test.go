package ts3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *Error
	}{
		{"ok",
			`error id=0 msg=ok`,
			&Error{Msg: "ok"},
		},
		{"invalid-server",
			`error id=1024 msg=invalid\sserverID`,
			&Error{
				ID:  1024,
				Msg: "invalid serverID",
			},
		},
		{"permission",
			`error id=2568 msg=insufficient\sclient\spermissions failed_permid=4 other=test`,
			&Error{
				ID:      2568,
				Msg:     "insufficient client permissions",
				Details: map[string]interface{}{"failed_permid": 4, "other": "test"},
			},
		},
		{"invalid",
			`   error id=0 msg=ok`,
			nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := respTrailerRe.FindStringSubmatch(tc.line)
			if tc.expected == nil {
				assert.Equal(t, 0, len(matches))
				return
			}

			if !assert.Equal(t, 4, len(matches)) {
				return
			}
			err := NewError(matches)
			assert.Error(t, err)
			assert.Equal(t, tc.expected, err)
			assert.NotEmpty(t, err.Error())
		})
	}
}

func TestNewInvalidResponseError(t *testing.T) {
	reason := "my reason"
	lines := []string{"line1"}
	err := NewInvalidResponseError(reason, lines)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), err.Reason)
	assert.Contains(t, err.Error(), err.Data[0])
	assert.Equal(t, reason, err.Reason)
	assert.Equal(t, lines, err.Data)
}
