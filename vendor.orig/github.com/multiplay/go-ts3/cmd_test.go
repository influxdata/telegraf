package ts3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmd(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *Cmd
		expect string
	}{
		{"cmd", NewCmd("version"),
			"version",
		},
		{"arg", NewCmd("use").
			WithArgs(NewArg("sid", 1)),
			"use sid=1",
		},
		{"args", NewCmd("use").
			WithArgs(NewArg("sid", 1), NewArg("port", 1234)),
			"use sid=1 port=1234",
		},
		{"args-option", NewCmd("use").
			WithArgs(NewArg("sid", 1), NewArg("port", 1234)).
			WithOptions("-virtual"),
			"use sid=1 port=1234 -virtual",
		},
		{"options", NewCmd("serverlist").
			WithOptions("-uid", "-short"),
			"serverlist -uid -short",
		},
		{"arg-group-single", NewCmd("servergroupdelperm").
			WithArgs(
				NewArg("sgid", 1),
				NewArgGroup(NewArg("permid", 1), NewArg("permid", 2)),
			),
			"servergroupdelperm sgid=1 permid=1|permid=2",
		},
		{"arg-group-multi", NewCmd("servergroupaddperm").
			WithArgs(
				NewArg("sgid", 1),
				NewArgGroup(
					NewArgSet(
						NewArg("permid", 1),
						NewArg("permvalue", 1),
						NewArg("permnegated", 0),
						NewArg("permskip", 0),
					),
					NewArgSet(
						NewArg("permid", 2),
						NewArg("permvalue", 2),
						NewArg("permnegated", 1),
						NewArg("permskip", 1),
					),
				),
			),
			"servergroupaddperm sgid=1 permid=1 permvalue=1 permnegated=0 permskip=0|permid=2 permvalue=2 permnegated=1 permskip=1",
		},
		{"escaped-chars", NewCmd("servergroupadd").
			WithArgs(NewArg("name", "Chars:\\/ |\a\b\f\n\r\t\v")),
			`servergroupadd name=Chars:\\\/\s\p\a\b\f\n\r\t\v`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect+"\n", tc.cmd.String())
		})
	}
}
