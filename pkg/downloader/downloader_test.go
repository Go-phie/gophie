package downloader

import (
	"testing"
)

var f = &fileDownloader{
	url:      "http://cdn2.mhpbooks.com/2016/02/google.jpg",
	filepath: "./",
}

func TestResponse(t *testing.T) {
	_, err := f.resp()
	if err != nil {
		t.Errorf("Response returning error")
	}
}

// Filesize must be greater than 0
func TestFileSize(t *testing.T) {
	if f.Filesize() == 0.0 {
		t.Errorf("Filesize returning 0")
	}
}
