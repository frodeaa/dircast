package core

import (
	"testing"
)

func TestFormatYear(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2015", "Thu, 01 Jan 2015 00:00:00 +0000"},
		{"20151102", "Mon, 02 Nov 2015 00:00:00 +0000"},
		{"20152402", "Tue, 24 Feb 2015 00:00:00 +0000"},
		{"", ""},
		{"NOT_A_YEAR", ""},
	}
	for _, c := range cases {
		got := formatYear(c.in)
		if got != c.want {
			t.Errorf("formatYear(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFileUrl(t *testing.T) {
	cases := []struct {
		relativePath, baseUrl, want string
	}{
		{"", "", ""},
		{"c/d/e", "a/b/", "a/b/c/d/e"},
		{"c d", "a/b/", "a/b/c%20d"},
		{"/c", "a/b/", "a/b/c"},
	}
	for _, c := range cases {
		got := fileUrl(c.relativePath, c.baseUrl)
		if got != c.want {
			t.Errorf("fileUrl(%q, %q) == %q, want %q", c.relativePath, c.baseUrl, got, c.want)
		}
	}
}
