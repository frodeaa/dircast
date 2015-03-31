package core

import (
    "time"
    "path/filepath"
    "os"
    "net/url"
    "strings"
    "crypto/sha1"
    "encoding/base64"
    id3 "github.com/mikkyang/id3-go"
    "github.com/mikkyang/id3-go/v2"
)

const (
    DEFAULT_FILE_TYPE = "mp3"
    DEFAULT_CHANNEL_TITLE = "RSS FEED"
)

type Source struct {
    channel Channel
    Root string
    publicUrl string
    recursive bool
    fileType string
    autoImage bool
}

type MediaItem struct {
    Item
}

func trimmed(value string) string {
    cutset := string(rune(0))
    return strings.TrimRight(value, cutset)
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
    return trimmed(year)
}

func fileUrl(relativePath string, baseUrl string) string {
    Url, _ := url.Parse(baseUrl)
    Url.Path += relativePath
    Url.Path = strings.Replace(Url.Path, "//", "/", -1)
    return Url.String()
}

func readImages(fd *id3.File, channel *Channel) []Image {
    var images []Image
    apic := fd.Frame("APIC")
    if apic != nil {
        switch t := apic.(type) {
            case *v2.ImageFrame:
            v2if := v2.ImageFrame(*t)
            hasher := sha1.New()
            hasher.Write(v2if.Data())
            images = append(images, Image{Title: channel.Title, Link: channel.Link, Url: channel.Link + base64.URLEncoding.EncodeToString(hasher.Sum(nil)),
                Blob: v2if.Data()})
        }
    }
    return images
}

func (m *MediaItem) addMeta(path, defaultName string, source *Source) {
    fd, err := id3.Open(path)
    if err != nil {
        m.Item.Title = defaultName
    }else {
        defer fd.Close()
        title := trimmed(fd.Title())
        author := trimmed(fd.Artist())
        if len(title) > 0 {
            m.Item.Title = title
        } else {
            m.Item.Title = author
            if len(author) > 0 {
                m.Item.Title += " - "
            }
            m.Item.Title += defaultName
        }
        m.Item.Subtitle = author
        tcon := fd.Frame("TCON")
        if tcon != nil {
            m.Item.Categories = append(m.Item.Categories, Text{Value: trimmed(tcon.String())})
        }
        m.Item.PubDate = formatYear(fd.Year())
        if source.autoImage && len(source.channel.Images) == 0 {
            source.channel.Images = readImages(fd, &source.channel)
        }
    }
}

func NewSource(root string, recursive bool, publicUrl string) *Source {
    channel := &Channel{
        PubDate:     time.Now().Format(time.RFC1123Z),
        Title:       DEFAULT_CHANNEL_TITLE}
    return &Source{
        Root: root,
        recursive: recursive,
        publicUrl: publicUrl,
        channel: *channel,
        fileType: DEFAULT_FILE_TYPE}
}

func (s *Source) SetFileType(fileType string) {
    s.fileType = fileType
}

func (s *Source) SetChannel(title, link, description, language string) {
    s.channel.Title = title
    s.channel.Link = link
    s.channel.Description = description
    s.channel.Language = language
}

func (s *Source) SetChannelImageUrl(url string) {
    s.channel.Images = append(s.channel.Images, Image{Title: s.channel.Title, Link: s.channel.Link, Url: url})
}

func (s *Source) SetAutoImage(autoImage bool) {
    s.autoImage = autoImage
}

func (s *Source) addFile(path string, info os.FileInfo) {
    url := fileUrl(path[len(s.Root)-1:], s.publicUrl)
    item := MediaItem{Item{Enclosure: Enclosure{Length: info.Size(), Type: "audio/mpeg",
        Url: url}, Guid: url}}
    item.addMeta(path, info.Name(), s)
    s.channel.Items = append(s.channel.Items, item.Item)
}

func (s *Source) HandleWalk() filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !s.recursive && path != s.Root && info.IsDir() {
            return filepath.SkipDir
        }

        if !!info.IsDir() {
            return nil
        }

        matched, _ := filepath.Match("*."+s.fileType, info.Name())
        if matched {
            s.addFile(path, info)
        }
        return nil
    }
}

func (s *Source) Rss() *Rss {
    return &Rss{Channel: s.channel, Version: "2.0", NS: ITunesNs}
}


