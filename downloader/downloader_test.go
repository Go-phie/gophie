package downloader

import (
	"testing"
)

var f = &Downloader{
	URL:    "http://cdn2.mhpbooks.com/2016/02/google.jpg",
	Name:   "Test-image",
	Source: "NetNaija",
}

// Filesize must be greater than 0
func TestFileSize(t *testing.T) {
	f.DownloadFile()
}
