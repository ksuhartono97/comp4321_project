package main

import (
	//"./github.com/silver-rush/indexer"
	"./github.com/silver-rush/database"
	"./github.com/ksuhartono97/webcrawler"
)

func main() {
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

    webcrawler.CrawlerInit()
    webcrawler.PrintLinks("http://www.cse.ust.hk/")
}
