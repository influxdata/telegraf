package internal

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	commonFlag := ""
	commandFlag := ""

	cli := New("test desc",
		func(f *flag.FlagSet) {
			f.StringVar(&commonFlag, "c", "", "common flag")
		}, []*Command{
			{
				Name: "test",
				Args: []string{"A", "B", "C"},
				Desc: "just a test",
				Handler: func(args []string) error {
					return OutputLine(strings.Join(args, ""))
				},
				ParseFunc: func(fs *flag.FlagSet) {
					fs.StringVar(&commandFlag, "s", "", "command flag")
				},
			},
		},
	)

	g, err := capture(func() error {
		return cli.Run([]string{"run", "-c", "c", "test", "-s", "s", "a", "b", "c"})
	})
	if err != nil {
		t.Fatal(err)
	}

	w := "abc\n"
	if string(g) != w {
		t.Errorf("output = %q, want %q", string(g), w)
	}
	if commonFlag != "c" {
		t.Errorf("commonFlag = %q, want %q", commonFlag, "c")
	}
	if commandFlag != "s" {
		t.Errorf("commandFlag = %q, want %q", commandFlag, "s")
	}
}

// capture stdout
func capture(fn func() error) ([]byte, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())

	tmp := os.Stdout
	os.Stdout = f
	if err := fn(); err != nil {
		return nil, err
	}
	os.Stdout = tmp

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}
