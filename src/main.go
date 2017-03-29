package main

import (
	"fmt"
	"net/http"
	"time"

	"./github.com/silver-rush/indexer"
	"./github.com/silver-rush/database"
)

func main() {
	database.OpenAllDatabase()
	defer database.CloseAllDatabase()

	url := "http://www.cse.ust.hk/"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	slice := []uint64{0}
	indexer.Feed(0, resp.Body, time.Now(), 0, 0, slice)
}
