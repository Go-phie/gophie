# Gophie

Search and download movies without the ads or having to stare at naked butts

![Demo](assets/reel.jpeg)

[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bisoncorps/gophie)
[![Go Report Card](https://goreportcard.com/badge/github.com/bisoncorps/gophie)](https://goreportcard.com/report/github.com/bisoncorps/gophie)

## Usage

gophie

```bash
# to access the help menu
>>> help

Commands:
  api         host gophie as an API on a PORT env variable, fallback to set argument
  clear       clear the screen
  exit        exit the program
  help        display help
  search      search for movie


>>> search avenge 
Which do you want to download?
 ❯ Yoruba Movie: Avenger
   Revenge (2017)
   Nollywood Movie: Heartbreaker's Revenge
   Avengers: Endgame (2019)
   Avengers: Infinity War (2018)
   Nollywood Movie: Heartbreaker's Revenge (Part 2)
   ...

>>> api 9000
listening on :9000

# use the following to search for "good boys" on the hosted api
curl -s 'http://127.0.0.1:9000/?search=good+boys'
[
  {
    "Index":0,
    "Title":"Good Boys (2019)",
    "PictureLink":"https://img.netnaija.com/-c2HHK.jpg",
    "Description":...
  },
  ...
]

2020/02/11 01:45:42 searching for good boys
2020/02/11 01:45:50 Completed search for good boys
```


## Deployed

Deployed version is hosted [here](#)

## Todo 

- [x] Create cli and api
- [ ] Fix NetNaija link issue
- [ ] Setup CI/CD pipeline to autodeploy
- [ ] Patch download pkg into CLI with progress bar
- [ ] Host API on Heroku
- [ ] Update README
- [ ] Generate binaries for all platforms
- [ ] Write tests
- [ ] Create React app to consume hosted API

## License (MIT)

This project is opened under the [MIT 2.0 License](https://github.com/bisoncorps/gophie/blob/master/LICENSE) which allows very broad use for both academic and commercial purposes.


## Credits
Library | Use
------- | -----
[github.com/fatih/color](https://github.com/fatih/color) | color capabilities
[github.com/abiosoft/shell](https://github.com/abiosoft/shell) | creating an interactive cli
[github.com/gocolly/colly](https://github.com/gocolly/colly) | scraping the net for links
