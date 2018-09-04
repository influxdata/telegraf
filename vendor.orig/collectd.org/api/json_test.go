package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestValueList(t *testing.T) {
	vlWant := ValueList{
		Identifier: Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Unix(1426585562, 999000000),
		Interval: 10 * time.Second,
		Values:   []Value{Gauge(42)},
		DSNames:  []string{"legacy"},
	}

	want := `{"values":[42],"dstypes":["gauge"],"dsnames":["legacy"],"time":1426585562.999,"interval":10.000,"host":"example.com","plugin":"golang","type":"gauge"}`

	got, err := vlWant.MarshalJSON()
	if err != nil || string(got) != want {
		t.Errorf("got (%s, %v), want (%s, nil)", got, err, want)
	}

	var vlGot ValueList
	if err := vlGot.UnmarshalJSON([]byte(want)); err != nil {
		t.Errorf("got %v, want nil)", err)
	}

	// Conversion to float64 and back takes its toll -- the conversion is
	// very accurate, but not bit-perfect.
	vlGot.Time = vlGot.Time.Round(time.Millisecond)
	if !reflect.DeepEqual(vlWant, vlGot) {
		t.Errorf("got %#v, want %#v)", vlGot, vlWant)
	}
}

func ExampleValueList_UnmarshalJSON() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("while reading body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var vls []*ValueList
		if err := json.Unmarshal(data, &vls); err != nil {
			log.Printf("while parsing JSON: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, vl := range vls {
			var w Writer
			w.Write(r.Context(), vl)
			// "w" is a placeholder to avoid cyclic dependencies.
			// In real live, you'd do something like this here:
			// exec.Putval.Write(vl)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
