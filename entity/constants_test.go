package entity

import "testing"

func TestStatusText(t *testing.T) {
	cases := []struct {
		status byte
		want   string
	}{
		{SUCCESS, "success"},
		{WARM, "warning"},
		{ERROR, "error"},
		{9, "unknown"},
		{255, "unknown"},
	}
	for _, c := range cases {
		if got := StatusText(c.status); got != c.want {
			t.Errorf("StatusText(%d) = %q, want %q", c.status, got, c.want)
		}
	}
}
