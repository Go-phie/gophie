package downloader

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
)

// FileDownloader : structure for file downloader
type FileDownloader struct {
	URL      string  // Url to be downloaded from
	Name     string  // Name of File to be Download
	Dir      string  // Directory to store the file
	FileName string  // Filename of file with extension
	Mb       float64 // Mb is the size in megabytes
	RawSize  int64   // raw size
}

func (f *FileDownloader) resp() (*http.Response, error) {
	resp, err := http.Get(f.URL)
	return resp, err
}

func (f *FileDownloader) filePath() string {
	return path.Join(f.Dir, f.FileName)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (f *FileDownloader) DownloadFile() error {

	// Get the data
	resp, err := f.resp()
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`filename="(.*)"`)
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		f.FileName = re.FindStringSubmatch(contentDisposition)[1]
	} else {
		// example: Content-Type: [text/mp4]
		mimeType := strings.Split(resp.Header.Get("Content-Type"), "/")[1]
		f.FileName = fmt.Sprintf("%v.%v", f.Name, mimeType)
	}

	// TODO Choose Default File Path for Download, preferably &HOME/Downloads (unix)
	// %USERPROFILE%\Downloads Windows
	if f.Dir == "" {
		cwd, _ := os.Getwd()
		f.Dir = path.Join(cwd, "Gophie_Downloads", f.Name)
	}
	err = os.MkdirAll(f.Dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Downloading at ", f.filePath())
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(f.filePath())
	if err != nil {
		return err
	}
	defer out.Close()

	bar := pb.Full.Start64(f.RawSize)

	// create proxy reader
	barReader := bar.NewProxyReader(resp.Body)

	// Write the body to file
	_, err = io.Copy(out, barReader)
	if err != nil {
		log.Fatal(err)
	}
	bar.Finish()

	log.Infof("Download Complete. Saved at %v", f.filePath())

	return err
}

// GetFileSize : Check the file size
func (f *FileDownloader) GetFileSize() int64 {
	resp, err := f.resp()
	if err != nil {
		log.Fatal(err)
	}

	f.Mb = math.Round(float64(resp.ContentLength) / 1048576)
	f.RawSize = resp.ContentLength
	return f.RawSize
}
