package gnmi

import (
	"testing"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	path := "/foo/bar/bla[shoo=woo][shoop=/woop/]/z"
	parsed, err := ParsePath("theorigin", path, "thetarget")

	require.NoError(t, err)
	require.Equal(t, "theorigin", parsed.Origin)
	require.Equal(t, "thetarget", parsed.Target)
	require.Equal(t, []*gnmi.PathElem{{Name: "foo"}, {Name: "bar"},
		{Name: "bla", Key: map[string]string{"shoo": "woo", "shoop": "/woop/"}}, {Name: "z"}}, parsed.Elem)

	parsed, err = ParsePath("", "", "")
	require.NoError(t, err)
	require.Equal(t, &gnmi.Path{}, parsed)

	parsed, err = ParsePath("", "/foo[[", "")
	require.Nil(t, parsed)
	require.Error(t, err)
}
