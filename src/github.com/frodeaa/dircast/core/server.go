package core

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type rssHandler struct {
	feed      Rss
	fs        http.Handler
	path      string
	imageBlob []byte
}

func NewRssHandler(source Source) *rssHandler {
	return &rssHandler{feed: *source.Rss(),
		fs: http.FileServer(http.Dir(source.Root)), imageBlob: source.image}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func isImagePath(path string, images []Image) bool {
    return len(images) > 0 && strings.HasSuffix(images[0].Url, path)
}

func (rss *rssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "" || path == "/" {
		rss.feed.Out(w)
	} else if len(rss.imageBlob) > 0 && isImagePath(path, rss.feed.Channel.Images) {
		w.Write(rss.imageBlob)
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

func contentPath(url *url.URL) string {
	path := url.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func Server(source Source, logEnabled bool) error {

	url, _ := url.Parse(source.publicUrl)
	http.Handle(contentPath(url), NewRssHandler(source))

	writeStartupMsg(source.Root, source.publicUrl)
	onShutdown("dircast stopped.")

	if logEnabled {
		http.ListenAndServe(url.Host, Log(http.DefaultServeMux))
	}
	return http.ListenAndServe(url.Host, nil)

}
