package main

import (
	"encoding/xml"
	"fmt"
	id3 "github.com/mikkyang/id3-go"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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
			if err != nil {
				t, err = time.Parse("2006", year)
				if err != nil {
					t, err = time.Parse("20060201", year)
					if err != nil {
						return ""
					}
				}
			}
		}
		return t.Format(time.RFC1123Z)
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
		tcon := fd.Frame("TCON")
		if tcon != nil {
			item.Categories = append(item.Categories, Text{Value: tcon.String()})
		}
		item.PubDate = formatYear(fd.Year())
	}
}

func visitFiles(workDir string, channel *Channel, publicUrl string, recursive bool, fileType string) filepath.WalkFunc {
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

		matched, _ := filepath.Match("*."+fileType, f.Name())
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

type rssHandler struct {
	header string
	body   []byte
	fs     http.Handler
	path   string
}

func (rss *rssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "" || path == "/" {
		w.Write([]byte(rss.header))
		w.Write(rss.body)
	} else {
		http.StripPrefix(rss.path, rss.fs).ServeHTTP(w, r)
	}
}

func writeStartupMsg(workdir string, url string) {
	fmt.Printf(
		"\x1b[33;1m%v\x1b[0m \x1b[36;1m%v\x1b[0m \x1b[33;1mon:\x1b[0m \x1b[36;1m%v\x1b[0m\n",
		"Starting up dircast, serving", workdir, url)
	fmt.Println("Hit CTRL-C to stop the server")
}

func onShutdown(message string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("\x1b[31;1m%v\x1b[0m\n", message)
		os.Exit(1)
	}()
}

func server(output []byte, workdir string, baseUrl *url.URL) error {

	path := baseUrl.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	rss := &rssHandler{header: Header, body: output,
		fs: http.FileServer(http.Dir(workdir))}

	http.Handle(path, rss)

	writeStartupMsg(workdir, baseUrl.String())
	onShutdown("dircast stopped.")

	return http.ListenAndServe(baseUrl.Host, nil)

}

var (
	baseUrl     = kingpin.Flag("server", "hostname (and path) to the root e.g. http://myserver.com/rss").Short('s').Default("http://localhost:8000/").URL()
	bind        = kingpin.Flag("bind", "Start HTTP server, bind to the server").Short('b').Bool()
	recursive   = kingpin.Flag("recursive", "how to handle the directory scan").Short('r').Bool()
	autoImage   = kingpin.Flag("auto-image", "Resolve RSS image automatically, will use cover art if available, image overrides this option").Short('a').Bool()
	language    = kingpin.Flag("language", "the language of the RSS document, a ISO 639 value").Short('l').String()
	title       = kingpin.Flag("title", "RSS channel title").Short('t').Default("RSS FEED").String()
	description = kingpin.Flag("description", "RSS channel description").Short('d').String()
	imageUrl    = kingpin.Flag("image", "Image URL for the RSS channel image").Short('i').URL()
	fileType    = kingpin.Flag("file", "File type to include in the RSS document").Short('f').Default("mp3").String()
	path        = kingpin.Arg("directory", "directory to read files relative from").Required().ExistingDir()
)

func main() {

	kingpin.Version("0.2.0")
	kingpin.Parse()

	channel := &Channel{
		PubDate:     time.Now().Format(time.RFC1123Z),
		Title:       *title,
		Link:        (*baseUrl).String(),
		Description: *description,
		Language:    *language}

	if !strings.HasSuffix((*baseUrl).Path, "/") {
		(*baseUrl).Path = (*baseUrl).Path + "/"
	}

	if *imageUrl != nil {
		channel.Images = append(channel.Images, Image{Title: channel.Title, Link: channel.Link, Url: (*imageUrl).String()})
		*autoImage = false
	}
	err := filepath.Walk(*path, visitFiles(*path, channel, (*baseUrl).String(), *recursive, *fileType))

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		output, err := xml.MarshalIndent(
			&Rss{Channel: *channel, Version: "2.0", NS: iTunesNs}, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			if *bind {
				err = server(output, *path, *baseUrl)
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
			} else {
				os.Stdout.WriteString(Header)
				os.Stdout.Write(output)
			}
		}
	}

}
