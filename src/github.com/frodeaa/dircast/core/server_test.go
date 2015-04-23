package core

import (
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
