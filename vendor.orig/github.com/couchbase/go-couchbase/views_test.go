package couchbase

import (
	"testing"
	"unsafe"
)

func TestViewError(t *testing.T) {
	e := ViewError{"f", "r"}
	exp := `Node: f, reason: r`
	if e.Error() != exp {
		t.Errorf("Expected %q, got %q", exp, e.Error())
	}
}

func mkNL(in []Node) unsafe.Pointer {
	return unsafe.Pointer(&in)
}

func TestViewURL(t *testing.T) {
	// Missing URL
	b := Bucket{nodeList: mkNL([]Node{{}})}
	v, err := b.ViewURL("a", "b", nil)
	if err == nil {
		t.Errorf("Expected error on missing URL, got %v", v)
	}

	// Invalidish URL
	b = Bucket{nodeList: mkNL([]Node{{CouchAPIBase: "::gopher:://localhost:80x92/"}})}
	v, err = b.ViewURL("a", "b", nil)
	if err == nil {
		t.Errorf("Expected error on broken URL, got %v", v)
	}

	// Unmarshallable parameter
	b = Bucket{nodeList: mkNL([]Node{{CouchAPIBase: "http:://localhost:8092/"}})}
	v, err = b.ViewURL("a", "b",
		map[string]interface{}{"ch": make(chan bool)})
	if err == nil {
		t.Errorf("Expected error on unmarshalable param, got %v", v)
	}

	tests := []struct {
		ddoc, name string
		params     map[string]interface{}
		exppath    string
		exp        map[string]string
	}{
		{"a", "b",
			map[string]interface{}{"i": 1, "b": true, "s": "ess"},
			"/x/_design/a/_view/b",
			map[string]string{"i": "1", "b": "true", "s": `"ess"`}},
		{"a", "b",
			map[string]interface{}{"unk": DocID("le"), "startkey_docid": "ess"},
			"/x/_design/a/_view/b",
			map[string]string{"unk": "le", "startkey_docid": "ess"}},
		{"a", "b",
			map[string]interface{}{"stale": "update_after"},
			"/x/_design/a/_view/b",
			map[string]string{"stale": "update_after"}},
		{"a", "b",
			map[string]interface{}{"startkey": []string{"a"}},
			"/x/_design/a/_view/b",
			map[string]string{"startkey": `["a"]`}},
		{"", "_all_docs", nil, "/x/_all_docs", map[string]string{}},
	}

	b = Bucket{Name: "x",
		nodeList: mkNL([]Node{{CouchAPIBase: "http://localhost:8092/", Status: "healthy"}})}
	for _, test := range tests {
		us, err := b.ViewURL(test.ddoc, test.name, test.params)
		if err != nil {
			t.Errorf("Failed on %v: %v", test, err)
			continue
		}

		u, err := ParseURL(us)
		if err != nil {
			t.Errorf("Failed on %v", test)
			continue
		}

		if u.Path != test.exppath {
			t.Errorf("Expected path of %v to be %v, got %v",
				test, test.exppath, u.Path)
		}

		got := u.Query()

		if len(got) != len(test.exp) {
			t.Errorf("Expected %v, got %v", test.exp, got)
			continue
		}

		for k, v := range test.exp {
			if len(got[k]) != 1 || got.Get(k) != v {
				t.Errorf("Expected param %v to be %q on %v, was %#q",
					k, v, test, got[k])
			}
		}
	}
}

func TestBadViewParam(t *testing.T) {
	b := Bucket{Name: "x",
		nodeList: mkNL([]Node{{CouchAPIBase: "http://localhost:8092/",
			Status: "healthy"}})}
	thing, err := b.ViewURL("adoc", "aview", map[string]interface{}{
		"aparam": make(chan bool),
	})
	if err == nil {
		t.Errorf("Failed to build a view with a bad param, got %v",
			thing)
	}

}
