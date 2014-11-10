package main

import (
	"encoding/xml"
	"fmt"
	id3 "github.com/mikkyang/id3-go"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"net/url"
	"os"
	"path/filepath"
)

const (
	Header   = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	iTunesNs = "http://www.itunes.com/dtds/podcast-1.0.dtd"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
	Version string   `xml:"version,attr"`
	NS      string   `xml:"xmlns:itunes,attr"`
}

type Channel struct {
	Title       string `xml:"title,omitempty"`
	Link        string `xml:"link,omitempty"`
	Description string `xml:"description,omitempty"`
	Items       []Item `xml:"item"`
	Image       Image  `xml:"image,omitempty"`
}

type Image struct {
	Title  string `xml:"title,omitempty"`
	Url    string `xml:"url,omitempty"`
	Link   string `xml:"link,omitempty"`
	Width  int    `xml:"width,omitempty"`
	Height int    `xml:"height,omitempty"`
}

type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Enclosure   Enclosure `xml:"enclosure"`
	Guid        string    `xml:"guid"`
	Subtitle    string    `xml:"itunes:subtitle"`
}

type Enclosure struct {
	Url    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func fileUrl(relativePath string, baseUrl string) string {
	Url, _ := url.Parse(baseUrl)
	Url.Path += relativePath
	return Url.String()
}

func addMeta(path string, f os.FileInfo, item *Item) {
	fd, err := id3.Open(path)
	if err != nil {
		item.Title = f.Name()
	} else {
		title := fd.Title()
		author := fd.Artist()
		if len(title) > 0 {
			item.Title = title
		} else {
			item.Title = author
			if len(author) > 0 {
				item.Title += " - "
			}
			item.Title += f.Name()
		}
		item.Subtitle = author
	}
}

func visitFiles(workDir string, channel *Channel, publicUrl string) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !!f.IsDir() {
			return nil
		}

		matched, _ := filepath.Match("*.mp3", f.Name())
		if matched {
			url := fileUrl(path[len(workDir)-1:], publicUrl)
			item := Item{Enclosure: Enclosure{Length: f.Size(), Type: "audio/mpeg",
				Url: url}, Guid: url}
			addMeta(path, f, &item)
			channel.Items = append(channel.Items, item)
		}

		return nil

	}
}

var (
	baseUrl     = kingpin.Flag("server", "hostname (and path) to the root e.g. http://myserver.com/rss").Short('s').Default("http://localhost").URL()
	title       = kingpin.Flag("title", "RSS channel title").Short('t').Default("RSS FEED").String()
	description = kingpin.Flag("description", "RSS channel description").Short('d').String()
	path        = kingpin.Arg("directory", "directory to read files relative from").Required().ExistingDir()
)

func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()

	channel := &Channel{Title: *title, Link: (*baseUrl).String(), Description: *description}
	err := filepath.Walk(*path, visitFiles(*path, channel, (*baseUrl).String()))

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		output, err := xml.MarshalIndent(
			&Rss{Channel: *channel, Version: "2.0", NS: iTunesNs}, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			os.Stdout.WriteString(Header)
			os.Stdout.Write(output)
		}
	}

}
