package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"./github.com/silver-rush/database"
)

func TestReadDatabase(t *testing.T) {
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

	f, err := os.Create("./spider_result.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	docIDList := database.GetAllDoc()
	fmt.Println(len(docIDList))

	for _, docID := range docIDList {
		docInfo := database.GetDocInfo(docID)

		fmt.Println(docInfo.Title)
		fmt.Fprintln(f, docInfo.Title)

		url := database.GetURLWithID(docID)
		fmt.Println(url)
		fmt.Fprintln(f, url)

		timeString := time.Unix(docInfo.Time, 0).Format("Mon, 02 Jan 2006 15:04:05 GMT")
		fmt.Println(timeString, " ", docInfo.Size)
		fmt.Fprintln(f, timeString, " ", docInfo.Size)

		termList := database.GetTermsInDoc(docID)
		for _, termID := range termList {
			p := database.GetPosting(termID, docID)
			word := database.GetWordWithID(termID)
			fmt.Print(word, " ", p.TermFreq, ";")
			fmt.Fprint(f, word, " ", p.TermFreq, ";")
		}
		fmt.Println()
		fmt.Fprintln(f)

		for _, id := range docInfo.Child {
			url := database.GetURLWithID(id)
			fmt.Println("Child ", url)
			fmt.Fprintln(f, "Child ", url)
		}

		fmt.Println("-------------------------------------------------------------------------------------------")
		fmt.Fprintln(f, "-------------------------------------------------------------------------------------------")
	}
}
