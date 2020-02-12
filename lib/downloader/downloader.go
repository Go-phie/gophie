package downloader

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
)

// FileDownloader : structure for file downloader
type FileDownloader struct {
	// Url to be downloaded from
	URL string
	// Filepath to be saved to
	Filepath string
	// Mb is the size in megabytes
	Mb float64
	// raw size
	Rawsize float64
}

// PrintDownloadProgress ; Prints progress using the ishell context
func (f *FileDownloader) PrintDownloadProgress(c *ishell.Context, done chan int64) {
	var stop bool = false
	for {
		select {
		case <-done:
			stop = true
			c.ProgressBar().Stop()
		default:
			file, err := os.Open(f.Filepath)
			if err != nil {
				c.Println(err)
			}
			fi, err := file.Stat()
			if err != nil {
				c.Println(err)
			}
			// get file size
			size := fi.Size()

			if size == 0 {
				size = 1
			}
			// compute integer percent of current size against rawsize
			percent := math.Round((float64(size) / f.Rawsize) * 100)
			c.ProgressBar().Suffix(fmt.Sprint(" ", percent, "%"))
			c.ProgressBar().Progress(int(percent))
		}

		if stop {
			break
		}

		time.Sleep(time.Second)
	}

}

func (f *FileDownloader) resp() (*http.Response, error) {
	resp, err := http.Get(f.URL)
	return resp, err
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (f *FileDownloader) DownloadFile(c *ishell.Context) error {

	// Get the data
	resp, err := f.resp()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`filename="(.*)"`)
	content := resp.Header["Content-Disposition"][0]
	filename := re.FindStringSubmatch(content)[1]
	cwd, _ := os.Getwd()
	filepath := path.Join(cwd, filename)
	f.Filepath = filepath
	c.Println("Downloading at", filepath)
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(f.Filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	done := make(chan int64)
	c.ProgressBar().Start()
	go f.PrintDownloadProgress(c, done)

	// Write the body to file
	comp, err := io.Copy(out, resp.Body)

	done <- comp

	green := color.New(color.FgGreen).SprintFunc()
	c.Println(green("Completed Downloading ", f.Filepath))

	return err
}

// Filesize : Check the file size
func (f *FileDownloader) Filesize() float64 {
	resp, err := f.resp()
	if err == nil {
		f.Mb = math.Round(float64(resp.ContentLength) / 1048576)
		f.Rawsize = float64(resp.ContentLength)
		return f.Rawsize
	}
	return 0.00
}
