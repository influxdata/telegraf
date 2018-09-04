package cdtime // import "collectd.org/cdtime"

import (
	"testing"
	"time"
)

// TestConversion converts a time.Time to a cdtime.Time and back, expecting the
// original time.Time back.
func TestConversion(t *testing.T) {
	cases := []string{
		"2009-02-04T21:00:57-08:00",
		"2009-02-04T21:00:57.1-08:00",
		"2009-02-04T21:00:57.01-08:00",
		"2009-02-04T21:00:57.001-08:00",
		"2009-02-04T21:00:57.0001-08:00",
		"2009-02-04T21:00:57.00001-08:00",
		"2009-02-04T21:00:57.000001-08:00",
		"2009-02-04T21:00:57.0000001-08:00",
		"2009-02-04T21:00:57.00000001-08:00",
		"2009-02-04T21:00:57.000000001-08:00",
	}

	for _, s := range cases {
		want, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t.Errorf("time.Parse(%q): got (%v, %v), want (<time.Time>, nil)", s, want, err)
			continue
		}

		cdtime := New(want)
		got := cdtime.Time()
		if !got.Equal(want) {
			t.Errorf("cdtime.Time(): got %v, want %v", got, want)
		}
	}
}

func TestDecompose(t *testing.T) {
	cases := []struct {
		in    Time
		s, ns int64
	}{
		// 1546167635576736987 / 2^30 = 1439980823.1524536265...
		{Time(1546167635576736987), 1439980823, 152453627},
		// 1546167831554815222 / 2^30 = 1439981005.6712620165...
		{Time(1546167831554815222), 1439981005, 671262017},
		// 1546167986577716567 / 2^30 = 1439981150.0475896215...
		{Time(1546167986577716567), 1439981150, 47589622},
	}

	for _, c := range cases {
		s, ns := c.in.decompose()

		if s != c.s || ns != c.ns {
			t.Errorf("decompose(%d) = (%d, %d) want (%d, %d)", c.in, s, ns, c.s, c.ns)
		}
	}
}

func TestNewNano(t *testing.T) {
	cases := []struct {
		ns   uint64
		want Time
	}{
		// 1439981652801860766 * 2^30 / 10^9 = 1546168526406004689.4
		{1439981652801860766, Time(1546168526406004689)},
		// 1439981836985281914 * 2^30 / 10^9 = 1546168724171447263.4
		{1439981836985281914, Time(1546168724171447263)},
		// 1439981880053705608 * 2^30 / 10^9 = 1546168770415815077.4
		{1439981880053705608, Time(1546168770415815077)},
	}

	for _, c := range cases {
		got := newNano(c.ns)

		if got != c.want {
			t.Errorf("newNano(%d) = %d, want %d", c.ns, got, c.want)
		}
	}
}
