package main

import (
	"github.com/bisoncorps/gophie/cmd"
)

//func searchOrList(c *ishell.Context, result *scraper.SearchResult) {
//  choices := []string{}
//  for _, i := range result.Movies {
//    if i.Title != "" {
//      choices = append(choices, strconv.Itoa(1+i.Index)+" "+i.Title)
//    }
//  }
//  c.ProgressBar().Stop()
//  if len(choices) > 0 {
//    choice := c.MultiChoice(choices, yellow(result.Query))
//    if result.Movies[choice].IsSeries && strings.HasSuffix(result.Movies[choice].DownloadLink.String(), "download") {
//      c.Println("This series could not be parsed")
//      c.Println(result.Movies[choice].SDownloadLink)
//    } else {
//      url := result.Movies[choice].DownloadLink.String()
//      downloadhandler := &downloader.FileDownloader{
//        URL: url,
//        Mb:  0.0,
//      }
//      if file := downloadhandler.Filesize(); file != 0.0 {
//        c.Println("Starting Download ==> Size: ", downloadhandler.Mb, "MB")
//        err := downloadhandler.DownloadFile(c)
//        if err != nil {
//          c.Println(red(err))
//        }
//      }
//    }
//  } else {
//    c.Println(red("Could not find any match"))
//  }
//}

func main() {
	cmd.Execute()
}
