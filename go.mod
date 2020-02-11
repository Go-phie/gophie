module main.go

go 1.13

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/briandowns/spinner v1.8.0
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/fatih/color v1.9.0
	github.com/gocolly/colly/v2 v2.0.1
	gophie/pkg/scraper v0.0.0-00010101000000-000000000000
)

replace gophie/pkg/scraper => ./pkg/scraper
