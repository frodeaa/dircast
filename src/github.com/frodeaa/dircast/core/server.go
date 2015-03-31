package core

import (
    "os"
    "os/signal"
    "log"
    "net/http"
    "net/url"
    "strings"
    "syscall"
    "fmt"
)

type rssHandler struct {
    feed     Rss
    fs         http.Handler
    path       string
    blobImages []Image
}

func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        handler.ServeHTTP(w, r)
    })
}

func findBlob(path string, blobImages []Image) []byte {
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
        rss.feed.Out(w)
    } else if len(rss.blobImages) > 0 && len(findBlob(path, rss.blobImages)) > 0 {
        w.Write(findBlob(path, rss.blobImages))
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

func Server(source Source, logEnabled bool) error {

    url, _ := url.Parse(source.publicUrl)
    path := url.Path
    if !strings.HasSuffix(path, "/") {
        path += "/"
    }

    var blobImages []Image
    if source.autoImage {
        blobImages = source.channel.Images
    }

    rss := &rssHandler{feed: *source.Rss(),
        fs: http.FileServer(http.Dir(source.Root)), blobImages: blobImages}

    http.Handle(path, rss)

    writeStartupMsg(source.Root, source.publicUrl)
    onShutdown("dircast stopped.")

    if logEnabled {
        http.ListenAndServe(url.Host, Log(http.DefaultServeMux))
    }
    return http.ListenAndServe(url.Host, nil)

}