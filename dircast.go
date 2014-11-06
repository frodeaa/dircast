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
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
	Version string   `xml:"version,attr"`
	NS      string   `xml:"xmlns:itunes,attr"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
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

func fileUrl(relativePath string, baseUrl string) (string, error) {
	Url, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	Url.Path += relativePath
	return Url.String(), nil
}

func addMeta(path string, f os.FileInfo, item *Item) string {
	fd, err := id3.Open(path)
	name := ""
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
	return name
}

func visitFiles(workDir string, channel *Channel, publicUrl string) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !!f.IsDir() {
			return nil
		}

		matched, err := filepath.Match("*.mp3", f.Name())
		if err != nil {
			fmt.Println(err) // malformed pattern
			return err
		}

		if matched {
			relativePath := path[len(workDir):]
			url, err := fileUrl(relativePath, publicUrl)
			if err != nil {
				return err
			}

			enclosure := Enclosure{Length: f.Size(), Type: "audio/mpeg",
				Url: url}
			item := Item{Enclosure: enclosure, Guid: enclosure.Url}
			addMeta(path, f, &item)
			channel.Items = append(channel.Items, item)
		}

		return nil

	}
}

var (
	baseUrl = kingpin.Flag("server", "hostname (and path) to the root e.g. http://myserver.com/rss").Short('s').Default("http://localhost").String()
	path    = kingpin.Arg("directory", "directory to read files relative from").Required().String()
)

func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()

	file, err := os.Open(*path)
	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
		return
	}
	stat, err := file.Stat()
	if err != nil || !stat.IsDir() {
		fmt.Printf("%s: %v: Is a file\n", os.Args[0], *path)
		return
	}

	channel := &Channel{Title: "RSS FEED", Link: *baseUrl}
	err = filepath.Walk(*path, visitFiles(*path, channel, *baseUrl))

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		output, err := xml.MarshalIndent(
			&Rss{Channel: *channel, Version: "2.0", NS: "http://www.itunes.com/dtds/podcast-1.0.dtd"}, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			os.Stdout.WriteString(Header)
			os.Stdout.Write(output)
		}
	}

}
