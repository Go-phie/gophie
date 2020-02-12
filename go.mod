module main.go

go 1.13

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/briandowns/spinner v1.8.0
	github.com/fatih/color v1.9.0
	gophie/lib/downloader v0.0.0-00010101000000-000000000000
	gophie/pkg/scraper v0.0.0-00010101000000-000000000000
)

replace gophie/pkg/scraper => ./pkg/scraper

replace gophie/lib/downloader => ./lib/downloader
