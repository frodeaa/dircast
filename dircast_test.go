package main

import (
	"os"
	"strings"
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

func TestAddMetaTitle(t *testing.T) {
	title := "Nice Life (Feat. Basick)"
	in := Item{}
	path := "./vendor/src/github.com/mikkyang/id3-go/test.mp3"
	f, _ := os.Stat(path)

	addMeta(path, f, &in, false)

	if in.Title != title {
		t.Errorf("addMeta(<test.mp3>, <FileInfo>, <Item>) Item.Title == %q want %q", in.Title, title)
	}
}

func TestAddMetaTitleFallbackOnFilieName(t *testing.T) {
	title := "test.mp3"
	in := Item{}
	f, _ := os.Stat("./vendor/src/github.com/mikkyang/id3-go/test.mp3")

	addMeta("NOT_PATH_TO_FILE", f, &in, false)

	if in.Title != title {
		t.Errorf("addMeta(<test.mp3>, <FileInfo>, <Item>) Item.Title == %q want %q", in.Title, title)
	}
}

func TestAddMetaPubDate(t *testing.T) {
	pubDate := "Mon, 25 Nov 2013 00:00:00 +0000"
	in := Item{}
	path := "./vendor/src/github.com/mikkyang/id3-go/test.mp3"
	f, _ := os.Stat(path)

	addMeta(path, f, &in, false)

	if in.PubDate != pubDate {
		t.Errorf("addMeta(<test.mp3>, <FileInfo>, <Item>) Item.PubDate == %q want %q", in.PubDate, pubDate)
	}

}

func TestAddMetaSubTitle(t *testing.T) {
	subtitle := "Paloalto"
	in := Item{}
	path := "./vendor/src/github.com/mikkyang/id3-go/test.mp3"
	f, _ := os.Stat(path)

	addMeta(path, f, &in, false)

	if !strings.HasPrefix(in.Subtitle, subtitle) {
		t.Errorf("addMeta(<test.mp3>, <FileInfo>, <Item>) Item.Subtitle ~= %q want %q", in.Subtitle, subtitle)
	}

}

func TestVisitFiles(t *testing.T) {

	cases := []struct {
		path, fileType string
		want           int
	}{
		{"./vendor/src/github.com/mikkyang/id3-go/", "mp3", 0},
		{"./vendor/src/github.com/mikkyang/id3-go/test.mp3", "mp3", 1},
	}

	for _, c := range cases {
		channel := &Channel{}
		f, e := os.Stat(c.path)
		v := visitFiles(c.path, channel, "test://", false, c.fileType, false)
		v(c.path, f, e)
		if len(channel.Items) != c.want {
			t.Errorf("visitFiles(%q, channel,...), len(channel.Items) == %d want %d",
				c.path, len(channel.Items), c.want)
		}
	}

}
