package core

import (
	"net/url"
	"testing"
)

func TestIsImagePath(t *testing.T) {

	if isImagePath("a_image_path", []Image{}) {
		t.Errorf("isImagePath(file,[]]) == %t, want %t", true, false)
	}

	var images = []Image{Image{Url: "http://server/image1.png"}}
	if isImagePath("image0.png", images) {
		t.Errorf("isImagePath(file,[]]) == %t, want %t", true, false)
	}

	if !isImagePath("image1.png", images) {
		t.Errorf("isImagePath(file,[]]) == %t, want %t", false, true)
	}

}

func TestContentPath(t *testing.T) {

	cases := []struct {
		baseUrl, want string
	}{
		{"", "/"},
		{"http://server", "/"},
		{"http://server:80/path/sub", "/path/sub/"},
	}

	for _, c := range cases {
		url, _ := url.Parse(c.baseUrl)
		got := contentPath(url)
		if got != c.want {
			t.Errorf("contentPath(%s) == %s, want %s", c.baseUrl, got, c.want)
		}
	}

}
