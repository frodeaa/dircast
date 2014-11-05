package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	id3 "github.com/mikkyang/id3-go"
	"net/url"
	"os"
	"path/filepath"
)

const (
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [dir]\n", os.Args[0])
}

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

func fileUrl(f os.FileInfo, baseUrl string) (string, error) {
	Url, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	Url.Path += f.Name()
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

func visitFiles(channel *Channel, publicUrl string) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil
		}

		if !!f.IsDir() {
			return nil // not a file. ignore
		}

		matched, err := filepath.Match("*.mp3", f.Name())
		if err != nil {
			fmt.Println(err) // malformed pattern
			return err
		}

		if matched {
			url, err := fileUrl(f, publicUrl)
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

func main() {
	flag.Parse()

	if flag.NArg() > 1 {
		usage()
		os.Exit(-1)
	}

	var workDir = os.Args[0]
	if flag.NArg() == 1 {
		workDir = flag.Arg(0)
	}

	publicUrl := "http://kirst.local:8080"
	channel := &Channel{Title: "RSS FEED", Link: publicUrl}
	err := filepath.Walk(workDir, visitFiles(channel, publicUrl))

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		output, err := xml.MarshalIndent(
			&Rss{Channel: *channel, Version: "2.0", NS: "http://www.itunes.com/dtds/podcast-1.0.dtd"}, " ", "	")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			os.Stdout.WriteString(Header)
			os.Stdout.Write(output)
		}
	}

}
