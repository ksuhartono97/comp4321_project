package main

// "fmt"
//
// "github.com/ksuhartono97/stringutil"

import (
    "log"

    "github.com/ksuhartono97/webcrawler"
    "github.com/boltdb/bolt"
)


func main() {
    // fmt.Printf("hello, world\n")
    // fmt.Printf(stringutil.Reverse("!oG ,olleH"))
    db, err := bolt.Open("index.db", 0600, nil)
  	if err != nil {
  		log.Fatal(err)
  	}
  	defer db.Close()
    webcrawler.CrawlerInit()
    webcrawler.PrintLinks("http://www.cse.ust.hk/")
    // webcrawler.CrawlerInit()
    // webcrawler.PrintLinks("http://www.facebook.com")
}
