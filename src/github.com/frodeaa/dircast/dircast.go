package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	dircast "github.com/frodeaa/dircast/core"
	id3 "github.com/mikkyang/id3-go"
	"github.com/mikkyang/id3-go/v2"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func readImages(fd *id3.File) []dircast.Image {
	var images []dircast.Image

	apic := fd.Frame("APIC")
	if apic != nil {
		switch t := apic.(type) {
		case *v2.ImageFrame:
			v2if := v2.ImageFrame(*t)
			hasher := sha1.New()
			hasher.Write(v2if.Data())
			images = append(images, dircast.Image{Title: "", Link: "", Url: base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
				Blob: v2if.Data()})
		}
	}

	return images
}

func addMeta(path string, f os.FileInfo, item *dircast.Item, autoImage bool) []dircast.Image {
	var images []dircast.Image
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
			item.Categories = append(item.Categories, dircast.Text{Value: strings.TrimRight(tcon.String(), cutset)})
		}
		item.PubDate = strings.TrimRight(formatYear(fd.Year()), cutset)
		if autoImage {
			images = readImages(fd)
		}
	}
	return images
}

func visitFiles(workDir string, channel *dircast.Channel, publicUrl string, recursive bool, fileType string, autoImage bool) filepath.WalkFunc {
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
			item := dircast.Item{Enclosure: dircast.Enclosure{Length: f.Size(), Type: "audio/mpeg",
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

	channel := &dircast.Channel{
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
		channel.Images = append(channel.Images, dircast.Image{Title: channel.Title, Link: channel.Link, Url: (*imageUrl).String()})
	}
	err := filepath.Walk(*path, visitFiles(*path, channel, (*baseUrl).String(), *recursive, *fileType, *autoImage))

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		rssFeed := &dircast.Rss{Channel: *channel, Version: "2.0", NS: dircast.ITunesNs}
		if *bind {
			var blobImages []dircast.Image
			if *autoImage {
				blobImages = channel.Images
			}
			err = dircast.Server(*rssFeed, *path, *baseUrl, blobImages, *logEnabled)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
		} else {
			rssFeed.Out(os.Stdout)
		}
	}

}
