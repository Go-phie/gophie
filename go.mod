module main.go

go 1.13

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/gocolly/colly/v2 v2.0.1
	gophie/pkg/scraper v0.0.0-00010101000000-000000000000
)

replace gophie/pkg/scraper => ./pkg/scraper
