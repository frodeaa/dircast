package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [dir]\n", os.Args[0])
}

type Rss struct {
	Channel Channel `xml:"channel"`
	Version string  `xml:"version,attr"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func visitFiles(channel *Channel) filepath.WalkFunc {
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
			item := Item{Title: f.Name()}
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

	channel := &Channel{Title: "RSS FEED"}
	err := filepath.Walk(workDir, visitFiles(channel))

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		output, err := xml.MarshalIndent(&Rss{Channel: *channel, Version: "2.0"}, " ", "	")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			os.Stdout.Write(output)
		}
	}

}
