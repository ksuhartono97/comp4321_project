package main

import (
	//"./github.com/ksuhartono97/webcrawler"
	"fmt"

	"./github.com/silver-rush/database"
)

func main() {
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

	docIDList := database.GetAllDoc()
	fmt.Println(len(docIDList))

	for _, docID := range docIDList {
		fmt.Println(database.GetURLWithID(docID))
		termList := database.GetTermsInDoc(docID)
		for _, termID := range termList {
			p := database.GetPosting(termID, docID)
			fmt.Println(database.GetWordWithID(termID)+" ", p.TermFreq)
		}
	}

	// webcrawler.CrawlerInit()
	// webcrawler.PrintLinks("http://www.cse.ust.hk/")
}
