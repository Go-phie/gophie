<p align="center"><img src="assets/reel.png" alt="Gophie" height="100px"></p>

<div align="center">
  <a href="https://godoc.org/github.com/bisoncorps/gophie">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square" alt="Documentation">
  </a>
  <a href="https://goreportcard.com/report/github.com/bisoncorps/gophie">
    <img src="https://goreportcard.com/badge/github.com/bisoncorps/gophie" alt="Go Report Card">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT">
  </a>
  <a href="https://travis-ci.org/bisoncorps/gophie">
    <img src="https://travis-ci.org/bisoncorps/gophie.svg?branch=master" alt="Build Status">
  </a>
</div>

# Gophie

Search and download movies without having to bump into ads. Feel free to add any new movie sites


## Installation
With Golang installed

```bash
go get github.com/bisoncorps/gophie
```
Alternatively download the Binaries and add to Path

- Windows - [64 bit](bin/windows/64-bit/gophie)
- Windows - [32-bit](bin/windows/32-bit/gophie)
- Linux - [x86_64](bin/linux/x86-64/gophie)

## Usage

gophie

```bash
Gophie

Usage:
  gophie [command]

Available Commands:
  api         host gophie as an API on a PORT env variable, fallback to set argument
  engines     Show summary and list of available engines
  help        Help about any command
  list        lists the recent movies by page number
  resume      resume downloads for previously stopped movies
  search      search for a movie
  version     Get Gophie Version

Flags:
  -e, --engine string       The Engine to use for querying and downloading (default "netnaija")
  -h, --help                help for gophie
  -o, --output-dir string   Path to download files to
  -v, --verbose             Display Verbose logs

Use "gophie [command] --help" for more information about a command.


Gophie - Bisoncorp (2020) (https://github.com/bisoncorps/gophie)
```

For Development use `go run main.go [command]`

### Supported Engines

- NetNaija
- FzMovies

## Deployed

Deployed version is hosted [here](https://gophie.herokuapp.com)

## Todo 

- [x] Create cli and api
- [x] Fix NetNaija link issue
- [x] Setup CI/CD pipeline to autodeploy
- [x] Patch download pkg into CLI with progress bar
- [x] Host API on Heroku
- [x] Update README
- [x] Generate binaries for all platforms
- [x] Create list movies by page feature
- [x] Add list movies by page feature into api
- [x] Write first ever tech article using Project experience
- [x] Write initial tests
- [x] Create React app to consume hosted API
- [x] Implement resume downloads
- [ ] Increment tests

## License (MIT)

This project is opened under the [MIT 2.0 License](https://github.com/bisoncorps/gophie/blob/master/LICENSE) which allows very broad use for both academic and commercial purposes.


## Credits
Library | Use
------- | -----
[github.com/gocolly/colly](https://github.com/gocolly/colly) | scraping the net for links
[github.com/manifoldco/promptui](https://github.com/manifoldco/promptui/) | interactive CLI
[github.com/spf13/cobra](https://github.com/spf13/cobra) | CLI interface
[github.com/iawia002/annie](https://github.com/iawia002/annie) | Downloader (resume capabilities)
