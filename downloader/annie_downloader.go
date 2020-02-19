package downloader

import (
	"encoding/json"
	"github.com/iawia002/annie/config"
	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors/universal"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// AnnieDownloader : pausable downloader
type AnnieDownloader struct {
	URL  string
	Dir  string // Directory to store the file
	Name string // Name of file
}

// DownloadFile : wrapper around the core annie downloader
func (f *AnnieDownloader) DownloadFile() error {
	var (
		err  error
		data []downloader.Data
	)
	var resume []AnnieDownloader
	var exist bool
	resumeFile := "gophie_cache/resume"
	data, err = universal.Extract(f.URL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if f.Dir == "" {
		cwd, _ := os.Getwd()
		f.Dir = path.Join(cwd, "Gophie_Downloads")
	}

	err = os.MkdirAll(f.Dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// store struct for later resumption
	if f2, err := os.Open(resumeFile); err != nil {
		log.Debug("No resume found in cache, creating new")
		resume = append(resume, *f)
		f1, err := os.Create(resumeFile)
		if err != nil {
			log.Debug(err)
		}
		enc := json.NewEncoder(f1)
		err = enc.Encode(resume)
		f1.Close()
	} else {
		dec := json.NewDecoder(f2)
		err = dec.Decode(&resume)
		f2.Close()
		if err == nil {
			exist = false
			for _, downloader := range resume {
				if downloader.Name == f.Name {
					exist = true
				}
			}
			// if the anniedownloader for that movie does not exist
			// in the resume cache, append it and save
			if !exist {
				log.Debug("Movie getting added to resume list")
				resume = append(resume, *f)
			}
			f3, _ := os.OpenFile(resumeFile, os.O_CREATE|os.O_WRONLY, 0644)
			enc := json.NewEncoder(f3)
			err = enc.Encode(resume)
			log.Debug(err)
			f3.Close()
		}
	}
	config.OutputPath = f.Dir
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			log.Fatal(item.URL, item.Err)
			continue
		}
		err = downloader.Download(item, f.URL, config.ChunkSizeMB)
		if err != nil {
			log.Fatal(item.URL, err)
			return err
		}
	}
	return nil

}
