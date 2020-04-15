package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

func newStopQuery(token int64) Query {
	return Query{
		Type:  p.Query_STOP,
		Token: token,
		Opts: map[string]interface{}{
			"noreply": true,
		},
	}
}
