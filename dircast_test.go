package main

import "testing"

func TestFormatYear(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2015", "Thu, 01 Jan 2015 00:00:00 +0000"},
		{"20151102", "Mon, 02 Nov 2015 00:00:00 +0000"},
		{"20152402", "Tue, 24 Feb 2015 00:00:00 +0000"},
	}
	for _, c := range cases {
		got := formatYear(c.in)
		if got != c.want {
			t.Errorf("formatYear(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
