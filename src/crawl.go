package main

import (
	"log"
	"time"

	"./github.com/ksuhartono97/stopword_rmv"
	"./github.com/ksuhartono97/webcrawler"
	"./github.com/silver-rush/database"
)

func main() {
	start := time.Now()
	stopword_rmv.ConstructRegex()
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()
	webcrawler.CrawlerInit()
	webcrawler.CrawlLinks(-1, "https://course.cse.ust.hk/comp4321/labs/TestPages/testpage.htm")
	elapsed := time.Since(start)
	log.Printf("Took %s\n", elapsed)
}
