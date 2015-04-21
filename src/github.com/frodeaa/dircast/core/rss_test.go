package core

import (
	"testing"
)

type Output struct {
	Writer
	data string
}

func (o *Output) Write(data []byte) (int, error) {
	o.data += string(data)
	return 0, nil
}

func TestOut(t *testing.T) {
	channel := Channel{
		PubDate:     "PUBDATE",
		Title:       "TITLE",
		Link:        "LINK",
		Description: "DESCRIPTION",
		Language:    "LANGUAGE",
	}

	channel.Items =
		append(channel.Items, Item{})

	out := &Output{}
	(&Rss{
		Channel: channel,
		Version: "2.0",
		NS:      "itunes",
	}).Out(out)

	want := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" +
		"\n<rss version=\"2.0\" xmlns:itunes=\"itunes\">" +
		"\n  <channel>\n    <pubDate>PUBDATE</pubDate>" +
		"\n    <title>TITLE</title>\n    <link>LINK</link>" +
		"\n    <description>DESCRIPTION</description>" +
		"\n    <language>LANGUAGE</language>\n    <item>" +
		"\n      <title></title>\n      <description></description>" +
		"\n      <enclosure url=\"\" length=\"0\" type=\"\"></enclosure>" +
		"\n      <guid></guid>\n    </item>\n  </channel>\n</rss>"

	if out.data != want {
		t.Errorf("rss.Out(writer) = %q, want %q", out.data, want)
	}

}
