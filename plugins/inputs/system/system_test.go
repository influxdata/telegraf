package system

import (
	"testing"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/stretchr/testify/require"
)

func TestUniqueUsers(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		data     []host.UserStat
	}{
		{
			name:     "single entry",
			expected: 1,
			data: []host.UserStat{
				{User: "root"},
			},
		},
		{
			name:     "emptry entry",
			expected: 0,
			data:     []host.UserStat{},
		},
		{
			name:     "all duplicates",
			expected: 1,
			data: []host.UserStat{
				{User: "root"},
				{User: "root"},
				{User: "root"},
			},
		},
		{
			name:     "all unique",
			expected: 3,
			data: []host.UserStat{
				{User: "root"},
				{User: "ubuntu"},
				{User: "ec2-user"},
			},
		},
		{
			name:     "mix of dups",
			expected: 3,
			data: []host.UserStat{
				{User: "root"},
				{User: "ubuntu"},
				{User: "ubuntu"},
				{User: "ubuntu"},
				{User: "ec2-user"},
				{User: "ec2-user"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := findUniqueUsers(tt.data)
			require.Equal(t, tt.expected, actual, tt.name)
		})
	}
}
