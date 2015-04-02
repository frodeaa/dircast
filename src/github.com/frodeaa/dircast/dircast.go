package main

import (
	"fmt"
	dircast "github.com/frodeaa/dircast/core"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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
			onShutdown("dircast stopped.")
			writeStartupMsg(source.Root, source.PublicUrl)
			err = dircast.Server(*source, *logEnabled)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
		} else {
			source.Rss().Out(os.Stdout)
		}
	}

}
