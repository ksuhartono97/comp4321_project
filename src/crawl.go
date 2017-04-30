package main

import (
	"log"
	"time"

	"./github.com/ksuhartono97/webcrawler"
	"./github.com/silver-rush/database"
)

func main() {
	start := time.Now()
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()
	webcrawler.CrawlerInit()
	webcrawler.CrawlLinks("https://course.cse.ust.hk/comp4321/labs/TestPages/testpage.htm")
	elapsed := time.Since(start)
	log.Printf("Took %s\n", elapsed)
}
