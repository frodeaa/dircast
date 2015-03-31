package main

import (
	"fmt"
	dircast "github.com/frodeaa/dircast/core"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"os"
	"path/filepath"
	"strings"
)

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

	if !strings.HasSuffix((*baseUrl).Path, "/") {
		(*baseUrl).Path = (*baseUrl).Path + "/"
	}

	if !*bind || *imageUrl != nil {
		*autoImage = false
	}

	source := dircast.NewSource(*path, *recursive, (*baseUrl).String())
	source.SetChannel(*title, (*baseUrl).String(), *description, *language)
	source.SetFileType(*fileType)
	source.SetAutoImage(*autoImage)
	if *imageUrl != nil {
		source.SetChannelImageUrl((*imageUrl).String())
	}
	err := filepath.Walk(*path, source.HandleWalk())

	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
	} else {
		if *bind {
			err = dircast.Server(*source, *logEnabled)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
		} else {
			source.Rss().Out(os.Stdout)
		}
	}

}
