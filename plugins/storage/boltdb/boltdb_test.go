package boltdb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type Foo struct {
	Shoes    int
	Carrots  string
	Aardvark bool
	Neon     float64
}

func TestSaveLoadCycle(t *testing.T) {
	p := &BoltDBStorage{
		Filename: "testdb.db",
	}
	defer func() {
		_ = os.Remove("testdb.db")
	}()
	err := p.Init()
	require.NoError(t, err)

	// check that loading a missing key doesn't fail
	err = p.Load("testing", "foo", &Foo{})
	require.NoError(t, err)

	foo := Foo{
		Shoes:    3,
		Carrots:  "blue",
		Aardvark: true,
		Neon:     3.1415,
	}
	err = p.Save("testing", "foo", foo)
	require.NoError(t, err)

	obj := &Foo{}
	err = p.Load("testing", "foo", obj)
	require.NoError(t, err)

	require.EqualValues(t, foo, *obj)

	err = p.Close()
	require.NoError(t, err)
}
