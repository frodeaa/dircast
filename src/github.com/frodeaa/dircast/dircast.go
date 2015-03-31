package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	rss "github.com/frodeaa/dircast/rss"
	id3 "github.com/mikkyang/id3-go"
	"github.com/mikkyang/id3-go/v2"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"log"
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

func fileUrl(relativePath string, baseUrl string) string {
	Url, _ := url.Parse(baseUrl)
	Url.Path += relativePath
	Url.Path = strings.Replace(Url.Path, "//", "/", -1)
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

func readImages(fd *id3.File) []rss.Image {
	var images []rss.Image

	apic := fd.Frame("APIC")
	if apic != nil {
		switch t := apic.(type) {
		case *v2.ImageFrame:
			v2if := v2.ImageFrame(*t)
			hasher := sha1.New()
			hasher.Write(v2if.Data())
			images = append(images, rss.Image{Title: "", Link: "", Url: base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
				Blob: v2if.Data()})
		}
	}

	return images
}

func addMeta(path string, f os.FileInfo, item *rss.Item, autoImage bool) []rss.Image {
	var images []rss.Image
	fd, err := id3.Open(path)
	if err != nil {
		item.Title = f.Name()
	} else {
		defer fd.Close()
		cutset := string(rune(0))
		title := strings.TrimRight(fd.Title(), cutset)
		author := strings.TrimRight(fd.Artist(), cutset)
		if len(title) > 0 {
			item.Title = title
		} else {
			item.Title = author
			if len(author) > 0 {
				item.Title += " - "
			}
			item.Title += strings.TrimRight(f.Name(), cutset)
		}
		item.Subtitle = author
		tcon := fd.Frame("TCON")
		if tcon != nil {
			item.Categories = append(item.Categories, rss.Text{Value: strings.TrimRight(tcon.String(), cutset)})
		}
		item.PubDate = strings.TrimRight(formatYear(fd.Year()), cutset)
		if autoImage {
			images = readImages(fd)
		}
	}
	return images
}

func visitFiles(workDir string, channel *rss.Channel, publicUrl string, recursive bool, fileType string, autoImage bool) filepath.WalkFunc {
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
			item := rss.Item{Enclosure: rss.Enclosure{Length: f.Size(), Type: "audio/mpeg",
				Url: url}, Guid: url}
			images := addMeta(path, f, &item, autoImage && len(channel.Images) == 0)
			if len(images) > 0 {
				channel.Images = images
				images[0].Title = channel.Title
				images[0].Link = channel.Link
				images[0].Url = channel.Link + images[0].Url
			}
			channel.Items = append(channel.Items, item)
		}

		return nil

	}
}

type rssHandler struct {
	header     string
	body       []byte
	fs         http.Handler
	path       string
	blobImages []rss.Image
}

func findBlob(path string, blobImages []rss.Image) []byte {
	blob := []byte{}
	for i := 0; i < len(blobImages); i++ {
		if strings.HasSuffix(blobImages[i].Url, path) {
			blob = blobImages[i].Blob
		}
	}
	return blob

}

func (rss *rssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "" || path == "/" {
		w.Write([]byte(rss.header))
		w.Write(rss.body)
	} else if len(rss.blobImages) > 0 && len(findBlob(path, rss.blobImages)) > 0 {
		w.Write(findBlob(path, rss.blobImages))
	} else {
		http.StripPrefix(rss.path, rss.fs).ServeHTTP(w, r)
	}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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

func server(output []byte, workdir string, baseUrl *url.URL, blobImages []rss.Image, logEnabled bool) error {

	path := baseUrl.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	rss := &rssHandler{header: Header, body: output,
		fs: http.FileServer(http.Dir(workdir)), blobImages: blobImages}

	http.Handle(path, rss)

	writeStartupMsg(workdir, baseUrl.String())
	onShutdown("dircast stopped.")

	if logEnabled {
		http.ListenAndServe(baseUrl.Host, Log(http.DefaultServeMux))
	}
	return http.ListenAndServe(baseUrl.Host, nil)

}

var (
	baseUrl     = kingpin.Flag("server", "hostname (and path) to the root e.g. http://myserver.com/rss").Short('s').Default("http://localhost:8000/").URL()
	bind        = kingpin.Flag("bind", "Start HTTP server, bind to the server").Short('b').Bool()
	logEnabled  = kingpin.Flag("log", "Enable log of HTTP requests").Short('l').Bool()
	recursive   = kingpin.Flag("recursive", "how to handle the directory scan").Short('r').Bool()
	autoImage   = kingpin.Flag("auto-image", "Resolve RSS image automatically, will use ID3 attached image, image overrides this option, only available in combination with bind").Short('a').Bool()
	language    = kingpin.Flag("language", "the language of the RSS document, a ISO 639 value").Short('l').String()
	title       = kingpin.Flag("title", "RSS channel title").Short('t').Default("RSS FEED").String()
	description = kingpin.Flag("description", "RSS channel description").Short('d').String()
	imageUrl    = kingpin.Flag("image", "Image URL for the RSS channel image").Short('i').URL()
	fileType    = kingpin.Flag("file", "File type to include in the RSS document").Short('f').Default("mp3").String()
	path        = kingpin.Arg("directory", "directory to read files relative from").Required().ExistingDir()
)

func main() {

	kingpin.Version("0.3.0")
	kingpin.Parse()

	channel := &rss.Channel{
		PubDate:     time.Now().Format(time.RFC1123Z),
		Title:       *title,
		Link:        (*baseUrl).String(),
		Description: *description,
		Language:    *language}

	if !strings.HasSuffix((*baseUrl).Path, "/") {
		(*baseUrl).Path = (*baseUrl).Path + "/"
	}

	if !*bind || *imageUrl != nil {
		*autoImage = false
	}

	if *imageUrl != nil {
		channel.Images = append(channel.Images, rss.Image{Title: channel.Title, Link: channel.Link, Url: (*imageUrl).String()})
	}
	err := filepath.Walk(*path, visitFiles(*path, channel, (*baseUrl).String(), *recursive, *fileType, *autoImage))

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		output, err := xml.MarshalIndent(
			&rss.Rss{Channel: *channel, Version: "2.0", NS: iTunesNs}, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			if *bind {
				var blobImages []rss.Image
				if *autoImage {
					blobImages = channel.Images
				}
				err = server(output, *path, *baseUrl, blobImages, *logEnabled)
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
