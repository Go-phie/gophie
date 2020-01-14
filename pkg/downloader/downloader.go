package downloader

import (
	"io"
	"net/http"
	"os"
)

type fileDownloader struct {
	url      string
	filepath string
}

func (f *fileDownloader) resp() (*http.Response, error) {
	resp, err := http.Get(f.url)
	return resp, err
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (f *fileDownloader) Download() error {

	// Get the data
	resp, err := f.resp()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(f.filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (f *fileDownloader) Filesize() float64 {
	resp, err := f.resp()
	if err == nil {
		return float64(resp.ContentLength)
	}
	return 0.00
}
