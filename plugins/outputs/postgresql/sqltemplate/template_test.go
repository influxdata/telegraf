package sqltemplate

import (
	"strings"
	"testing"
)

func TestRenderNilTableReturnsError(t *testing.T) {
	var tmpl Template
	if err := tmpl.UnmarshalText([]byte(`CREATE TABLE {{ .table }} ({{ .columns }})`)); err != nil {
		t.Fatal(err)
	}

	got, err := tmpl.Render(nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("rendered %q", got)
	}
	if got != nil {
		t.Fatalf("sql %q", got)
	}

	table := NewTable("public", "cpu", nil)
	got, err = tmpl.Render(table, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "cpu") {
		t.Fatalf("sql %q", got)
	}
}
