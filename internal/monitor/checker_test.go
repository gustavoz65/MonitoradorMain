package monitor

import "testing"

func TestMatchStatus(t *testing.T) {
	cases := []struct {
		code     int
		expected string
		want     bool
	}{
		{200, "", true},
		{404, "", false},
		{201, "201", true},
		{200, "201", false},
		{201, "200-299", true},
		{301, "200-299", false},
		{201, "2xx", true},
		{301, "2xx", false},
		{301, "3xx", true},
	}
	for _, c := range cases {
		if got := matchStatus(c.code, c.expected); got != c.want {
			t.Errorf("matchStatus(%d, %q) = %v, want %v", c.code, c.expected, got, c.want)
		}
	}
}
