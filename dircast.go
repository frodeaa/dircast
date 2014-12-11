package main

import (
	"encoding/xml"
	"fmt"
	id3 "github.com/mikkyang/id3-go"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"net/url"
	"os"
	"path/filepath"
	"time"
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
	Title       string  `xml:"title,omitempty"`
	Link        string  `xml:"link,omitempty"`
	Description string  `xml:"description,omitempty"`
	Language    string  `xml:"language,omitempty"`
	Images      []Image `xml:"image,omitempty"`
	Items       []Item  `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Enclosure   Enclosure `xml:"enclosure"`
	Guid        string    `xml:"guid"`
	Subtitle    string    `xml:"itunes:subtitle"`
	PubDate     string    `xml:"itunes:pubDate"`
}

type Image struct {
	Link  string `xml:"link"`
	Title string `xml:"title"`
	Url   string `xml:"url"`
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

func formatYear(year string) string {
	if len(year) > 0 {
		t, err := time.Parse("20060102", year)
		if err != nil {
			t, err = time.Parse("20060102", year[0:len(year)-1])
			if err == nil {
				return t.String()
			}
			t, err = time.Parse("2006", year)
			if err == nil {
				return t.String()
			}
			t, err = time.Parse("20060201", year)
			if err != nil {
				return ""
			}
		}
		return t.String()
	}
	return year
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
		item.PubDate = formatYear(fd.Year())
	}
}

func visitFiles(workDir string, channel *Channel, publicUrl string, recursive bool) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if f.IsDir() && path != workDir && !recursive {
			return filepath.SkipDir
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
	recursive   = kingpin.Flag("recursive", "how to handle the directory scan").Short('r').Bool()
	language    = kingpin.Flag("language", "the language of the RSS document, a ISO 639 value").Short('l').String()
	title       = kingpin.Flag("title", "RSS channel title").Short('t').Default("RSS FEED").String()
	description = kingpin.Flag("description", "RSS channel description").Short('d').String()
	imageUrl    = kingpin.Flag("image", "Image URL for the RSS channel image").Short('i').URL()
	path        = kingpin.Arg("directory", "directory to read files relative from").Required().ExistingDir()
)

func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()

	channel := &Channel{
		Title:       *title,
		Link:        (*baseUrl).String(),
		Description: *description,
		Language:    *language}

	err := filepath.Walk(*path, visitFiles(*path, channel, (*baseUrl).String(), *recursive))
	if *imageUrl != nil {
		channel.Images = append(channel.Images, Image{Title: channel.Title, Link: channel.Link, Url: (*imageUrl).String()})
	}

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
