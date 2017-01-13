package core

import (
	"crypto/sha1"
	"encoding/base64"
	id3 "github.com/mikkyang/id3-go"
	"github.com/mikkyang/id3-go/v2"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Source struct {
	channel   Channel
	Root      string
	PublicUrl string
	recursive bool
	fileType  string
	autoImage bool
	image     []byte
}

type MediaFile struct {
	*id3.File
	defaultName string
}

type MediaItem struct {
	Item
}

func trimmed(value string) string {
	cutset := string(rune(0))
	return strings.Trim(strings.TrimRight(value, cutset), " ")
}

func fileUrl(relativePath string, baseUrl string) string {
	Url, _ := url.Parse(baseUrl)
	Url.Path += relativePath
	Url.Path = strings.Replace(Url.Path, "//", "/", -1)
	return Url.String()
}

func formatYearRFC1123Z(pattern, year string) (string, error) {
	t, err := time.Parse(pattern, year)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC1123Z), nil
}

func formatUrl(url string) string {
	return "http://" + strings.TrimPrefix(url, "http://")
}

func formatYear(year string) string {
	if len(year) > 0 {
		t, err := formatYearRFC1123Z("20060102", year)
		if err == nil {
			return t
		}
		t, err = formatYearRFC1123Z("2006", year)
		if err == nil {
			return t
		}
		t, err = formatYearRFC1123Z("20060201", year)
		if err == nil {
			return t
		}
	}
	return ""
}

func (fd *MediaFile) titleArtist() string {
	title := trimmed(fd.Title())
	author := trimmed(fd.Artist())
	var res string
	if len(title) > 0 {
		res = title
	} else {
		res = author
		if len(author) > 0 {
			res += " - "
		}
		res += fd.defaultName
	}
	return res
}

func (fd *MediaFile) readImage() []byte {
	var data []byte
	apic := fd.Frame("APIC")
	if apic != nil {
		switch t := apic.(type) {
		case *v2.ImageFrame:
			data = v2.ImageFrame(*t).Data()
		}
	}
	return data
}

func (fd *MediaFile) copyCategoryTo(m *MediaItem) {
	tcon := fd.Frame("TCON")
	if tcon != nil {
		m.Item.Categories = append(m.Item.Categories, Text{Value: trimmed(tcon.String())})
	}
}

func (fd *MediaFile) yearFormatted() string {
	return formatYear(fd.Year())
}

func (fd *MediaFile) copyMetaTo(m *MediaItem) {
	m.Title = fd.titleArtist()
	m.Item.PubDate = fd.yearFormatted()
	fd.copyCategoryTo(m)
}

func (m *MediaItem) addMeta(path, defaultName string, source *Source, includeFileName bool) {
	fd, err := id3.Open(path)
	if err != nil {
		m.Item.Title = defaultName
	} else {
		media := &MediaFile{fd, defaultName}
		defer fd.Close()
		media.copyMetaTo(m)
		if source.autoImage && len(source.image) == 0 {
			source.SetImage(media.readImage())
		}
	}

	if includeFileName {
		m.Title = m.Title + " - " + filepath.Base(path)
	}

}

func NewSource(root string, recursive bool, publicUrl,
	title, description, language, filetype string, autoImage bool) *Source {
	channel := &Channel{
		PubDate:     time.Now().Format(time.RFC1123Z),
		Title:       title,
		Link:        formatUrl(publicUrl),
		Description: description,
		Language:    language}
	return &Source{
		Root:      root,
		recursive: recursive,
		PublicUrl: formatUrl(publicUrl),
		channel:   *channel,
		fileType:  filetype,
		autoImage: autoImage}
}

func (s *Source) SetImage(image []byte) {
	s.image = image
	if len(image) > 0 {
		hasher := sha1.New()
		hasher.Write(image)
		s.channel.Images = append(s.channel.Images,
			Image{Title: s.channel.Title, Link: s.channel.Link, Url: s.channel.Link + base64.URLEncoding.EncodeToString(hasher.Sum(nil))})
	}
}

func (s *Source) SetChannelImageUrl(url string) {
	s.channel.Images = append(s.channel.Images, Image{Title: s.channel.Title, Link: s.channel.Link, Url: url})
}

func (s *Source) addFile(path string, info os.FileInfo, includeFileName bool) {
	url := fileUrl(path[len(s.Root)-1:], s.PublicUrl)
	item := MediaItem{Item{Enclosure: Enclosure{Length: info.Size(), Type: "audio/mpeg",
		Url: url}, Guid: url}}
	item.addMeta(path, info.Name(), s, includeFileName)
	s.channel.Items = append(s.channel.Items, item.Item)
}

func (s *Source) HandleWalk(includeFileName bool) filepath.WalkFunc {
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
			s.addFile(path, info, includeFileName)
		}
		return nil
	}
}

func (s *Source) Rss() *Rss {
	return &Rss{Channel: s.channel, Version: "2.0", NS: ITunesNs}
}
