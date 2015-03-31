package core

import "encoding/xml"

const (
    Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
    ITunesNs = "http://www.itunes.com/dtds/podcast-1.0.dtd"
)

type Rss struct {
    XMLName xml.Name `xml:"rss"`
    Channel Channel  `xml:"channel"`
    Version string   `xml:"version,attr"`
    NS      string   `xml:"xmlns:itunes,attr"`
}

type Channel struct {
    PubDate     string  `xml:"pubDate,omitempty"`
    Title       string  `xml:"title,omitempty"`
    Link        string  `xml:"link,omitempty"`
    Description string  `xml:"description"`
    Language    string  `xml:"language,omitempty"`
    Images      []Image `xml:"image,omitempty"`
    Items       []Item  `xml:"item"`
}

type Item struct {
    Title       string    `xml:"title"`
    Description string    `xml:"description"`
    Enclosure   Enclosure `xml:"enclosure"`
    Guid        string    `xml:"guid"`
    Subtitle    string    `xml:"itunes:subtitle,omitempty"`
    Categories  []Text    `xml:"itunes:category,omitempty"`
    PubDate     string    `xml:"pubDate,omitempty"`
}

type Text struct {
    Value string `xml:"text,attr"`
}

type Image struct {
    Link  string `xml:"link"`
    Title string `xml:"title"`
    Url   string `xml:"url"`
    Blob  []byte `xml:"-"`
}

type Enclosure struct {
    Url    string `xml:"url,attr"`
    Length int64  `xml:"length,attr"`
    Type   string `xml:"type,attr"`
}

type Writer interface {
    Write([]byte) (int, error)
}

func (rss *Rss) Out(w Writer) (error) {
    body, err := xml.MarshalIndent(rss, "", "  ")
    w.Write([]byte(Header))
    w.Write([]byte(body))
    return err
}



