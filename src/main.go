package main

import (
	"./github.com/ksuhartono97/webcrawler"

	"./github.com/silver-rush/database"
)

func main() {
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

	webcrawler.CrawlerInit()
	webcrawler.PrintLinks("http://www.cse.ust.hk/")
}
