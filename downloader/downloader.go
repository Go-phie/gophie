package downloader

import (
	"encoding/json"
	"os"
	"path"

	"github.com/bisoncorps/gophie/engine"
	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Extract is the main function for extracting data before passing to Annie
func Extract(url, source string) ([]downloader.Data, error) {

	filename, ext, err := utils.GetNameAndExt(url)
	if err != nil {
		return nil, err
	}
	size, err := request.Size(url, url)
	if err != nil {
		return nil, err
	}
	urlData := downloader.URL{
		URL:  url,
		Size: size,
		Ext:  ext,
	}
	streams := map[string]downloader.Stream{
		"default": {
			URLs: []downloader.URL{urlData},
			Size: size,
		},
	}
	contentType, err := request.ContentType(url, url)
	if err != nil {
		return nil, err
	}

	return []downloader.Data{
		{
			Site:    source,
			Title:   filename,
			Type:    contentType,
			Streams: streams,
			URL:     url,
		},
	}, nil
}

// Downloader : pausable downloader
type Downloader struct {
	URL       string // URL Source
	Dir       string // Directory to store the file
	Name      string // Name of file
	Source    string // Name of the Source
	Completed bool   // Status of Download
}

//TODO:  Check if Download is completed and ask for redownload confirmation

// DownloadFile : Download Files using Annie Downloader
func (f *Downloader) DownloadFile() error {
	var (
		err  error
		data []downloader.Data
	)

	// Extract data to be downloaded with the streams
	data, err = Extract(f.URL, f.Source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(f.Dir, os.ModePerm)
	if err != nil {
		return err
	}

	config.OutputPath = f.Dir
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			return item.Err
		}
		err = downloader.Download(item, f.URL, config.ChunkSizeMB)
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadMovie : Download the movie
func DownloadMovie(movie *engine.Movie, outputDir string) error {
	downloadHandler := &Downloader{
		URL:    movie.DownloadLink.String(),
		Name:   movie.Title,
		Source: movie.Source,
	}

	downloadHandler.Dir = path.Join(outputDir, downloadHandler.Name)
	downloadListFile := path.Join(viper.GetString("gophie_cache"), "downloadList.json")

	var (
		downloads     []Downloader
		downloadsFile *os.File
		err           error
	)
	// Load Downloads if file exists
	if _, err = os.Stat(downloadListFile); err == nil {
		downloadsFile, err = os.OpenFile(downloadListFile, os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}
		dec := json.NewDecoder(downloadsFile)
		if err = dec.Decode(&downloads); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		// Create Download List File if it does not exists
		downloadsFile, err = os.Create(downloadListFile)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	// Load Config
	// Check for existing downloads
	exist := func() bool {
		for _, downloader := range downloads {
			if downloader.Name == downloadHandler.Name {
				return true
			}
		}
		return false
	}()
	// if file is not in Downloads cache, add and start the download
	// in the download cache, append it and save
	if !exist {
		log.Debug("Movie getting added to Download list")
		downloads = append(downloads, *downloadHandler)
		enc := json.NewEncoder(downloadsFile)
		if err = enc.Encode(downloads); err != nil {
			return err
		}
	}
	downloadsFile.Close()

	return downloadHandler.DownloadFile()
}
