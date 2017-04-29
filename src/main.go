package main

import (
	"log"
	"time"

	"./github.com/ksuhartono97/webcrawler"
	"./github.com/ksuhartono97/webserver"
	"./github.com/silver-rush/database"
)

func main() {
	start := time.Now()
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

	webcrawler.CrawlerInit()
	webcrawler.CrawlLinks("http://www.cse.ust.hk/")

	elapsed := time.Since(start)
	log.Printf("Took %s\n", elapsed)

	webserver.StartWebServer()
}
