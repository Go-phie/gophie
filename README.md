# Gophie

Search and download movies without having to bump into ads. Feel free to add any new movie sites

![Demo](assets/reel.jpeg)

[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bisoncorps/gophie)
[![Go Report Card](https://goreportcard.com/badge/github.com/bisoncorps/gophie)](https://goreportcard.com/report/github.com/bisoncorps/gophie)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/bisoncorps/gophie.svg?branch=master)](https://travis-ci.org/bisoncorps/gophie)

## Binaries

- Windows - [64 bit](bin/windows/64-bit/gophie)
- Windows - [32-bit](bin/windows/32-bit/gophie)
- Linux - [x86_64](bin/linux/x86-64/gophie)

## Usage

gophie

```bash
Usage:
  gophie [command]

Available Commands:
  api         host gophie as an API on a PORT env variable, fallback to set argument                                                                
  help        Help about any command
  list        lists the recent movies by page number
  search      search for a movie
  version     Get Gophie Version

Flags:
      --config string   config file (default is $HOME/.gophie.yaml)
      --engine string   The Engine to use for querying and downloading (default "NetNaija")                                                         
  -h, --help            help for gophie
  -t, --toggle          Help message for toggle

Use "gophie [command] --help" for more information about a command.
```

For Development use `go run main.go [command]`

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
- [ ] Implement resume downloads
- [ ] Include downloads over multiple goroutines
- [ ] Increment tests

## License (MIT)

This project is opened under the [MIT 2.0 License](https://github.com/bisoncorps/gophie/blob/master/LICENSE) which allows very broad use for both academic and commercial purposes.


## Credits
Library | Use
------- | -----
[github.com/gocolly/colly](https://github.com/gocolly/colly) | scraping the net for links
[github.com/manifoldco/promptui](https://github.com/manifoldco/promptui/) | interactive CLI
[github.com/spf13/cobra](https://github.com/spf13/cobra) | CLI interface
