package webcrawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"../../../golang.org/x/net/html"

	"../../silver-rush/database"
	"../../silver-rush/indexer"
)

type UrlData struct {
	sourceUrl    string
	sourceID     int64
	foundUrl     []string
	pageTitle    string
	pageSize     int
	rawHTML      string
	lastModified int64
}

type CrawlObject struct {
	url string
	id  int64
}

var exploredPages = 0
var expectedNumberOfPages = 300

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

//Fixes the URL to an absolute URL
func fixURL(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseURL.ResolveReference(uri)
	return uri.String()
}

//Gets the last modified time of a string
func getLastModifiedTime(href string) (int64, error) {
	currTime := time.Now().UTC().Unix()
	response, err := http.Head(href)
	if err != nil {
		fmt.Println("Error while downloading head of", href)
		return currTime, err
	} else {
		timeString := ""
		for k, v := range response.Header {
			if k == "Last-Modified" {
				timeString = v[0]
				if err != nil {
					fmt.Println(err)
					return currTime, err
				} else {
					layout := "Mon, 02 Jan 2006 15:04:05 GMT"
					t, err := time.Parse(layout, timeString)

					if err != nil {
						fmt.Println("Time Parsing error")
						panic(err)
						return currTime, err
					}
					return t.UTC().Unix(), nil
				}
			}
		}
		return currTime, nil
	}
}

// Extract all required info from a given webpage
func crawl(src string, srcID int64, parentID int64, ch chan UrlData, chFinished chan bool) {
	//Retrieve the webpage
	resp, err := http.Get(src)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println(err)
		// fmt.Println("ERROR: Failed to crawl \"" + src + "\"")
		return
	}

	urlResult := UrlData{sourceUrl: src, sourceID: srcID}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	//Open a secondary stream

	res, err := http.Get(src)

	c := res.Body
	defer c.Close()

	htmlData, err := ioutil.ReadAll(c)

	if err != nil {
		fmt.Println("Error cannot open")
		return
	}

	// Get page size in bytes
	urlResult.pageSize = len(htmlData)
	// Get raw HTML
	urlResult.rawHTML = string(htmlData)

	timeString, err := getLastModifiedTime(src)
	if err != nil {
		fmt.Println(err)
	}

	urlResult.lastModified = timeString

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done, increment explored pages and return result
			ch <- urlResult
			feedToIndexer(src, srcID, &urlResult, parentID)
			exploredPages++
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check the token tags
			if t.Data == "a" {
				// Extract the href value, if there is one
				ok, url := getHref(t)
				if !ok {
					continue
				}

				//Fix the URL into a absolute and valid form
				url = fixURL(url, src)
				// Make sure the url begins in http**
				hasProto := strings.Index(url, "http") == 0
				if hasProto && len(url) > 0 {
					urlResult.foundUrl = append(urlResult.foundUrl, url)
				}
			} else if t.Data == "title" {
				for {
					//Extract the title tag content
					t := z.Next()
					if t == html.TextToken {
						u := z.Token()
						urlResult.pageTitle += u.Data
					} else if t == html.EndTagToken {
						u := z.Token()
						if u.Data == "title" {
							//Finished, end extraction
							break
						}
					}
				}
			} else {
				continue
			}
		}
	}
}

//Feed result of crawl to the indexer to be indexed and saved to DB
func feedToIndexer(thisURL string, thisID int64, urlData *UrlData, parentID int64) {
	//Feeding to the indexer
	var wg sync.WaitGroup
	wg.Add(len(urlData.foundUrl))
	var childID []int64

	for _, u := range urlData.foundUrl {
		go func(url string) {
			defer wg.Done()
			id, _ := database.GetURLID(url)
			childID = append(childID, id)
		}(u)
	}

	wg.Wait()
	indexer.Feed(thisID, urlData.rawHTML, urlData.lastModified, int32(urlData.pageSize), parentID, childID, urlData.pageTitle)
	fmt.Printf("\nTime: %v\nSize: %v\nParent: %v\nChild: %v\nTitle: %v\n", urlData.lastModified, urlData.pageSize, parentID, childID, urlData.pageTitle)
}

//Main search function call this to crawl all the links, this is the base link, pass in -1 as parentID
func CrawlLinks(parentID int64, links ...string) {
	foundUrls := make(map[string]UrlData)
	seedUrls := []CrawlObject{}

	//Check to see if URL is okay to crawl
	for _, url := range links {
		lastMod, _ := getLastModifiedTime(url)
		_id, crawlCheck := indexer.CheckURL(url, lastMod)
		if crawlCheck {
			seedUrls = append(seedUrls, CrawlObject{url: url, id: _id})
		} else {
			if parentID != -1 {
				indexer.JustAddParentIDToURL(parentID, _id)
				fmt.Println("Added parentID ", parentID, "to ", _id)
			}
		}
	}

	// Channels
	chUrls := make(chan UrlData)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently)
	skipped := 0
	for _, obj := range seedUrls {
		// urlID, _ := database.GetURLID(url)
		size := database.GetTermsInDoc(obj.id)
		if len(size) == 0 {
			go crawl(obj.url, obj.id, parentID, chUrls, chFinished)
		} else {
			skipped++
		}
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls)-skipped; {
		select {
		case url := <-chUrls:
			foundUrls[url.sourceUrl] = url
		case <-chFinished:
			c++
		}
	}

	fmt.Println("\nTotal explored ", exploredPages)

	for _, url := range foundUrls {
		// Calculate remaining URLs needed
		diff := expectedNumberOfPages - exploredPages
		remaining := diff - len(url.foundUrl)
		toBeCalled := 0
		if remaining < 0 {
			toBeCalled = len(url.foundUrl) + remaining
		} else {
			toBeCalled = len(url.foundUrl)
		}
		urlArray := url.foundUrl[:toBeCalled]

		if toBeCalled > 0 {
			CrawlLinks(url.sourceID, urlArray...)
		}
	}
	close(chUrls)
	close(chFinished)
}

//To be called before each initial search, initializes the crawler
func CrawlerInit() {
	exploredPages = 0
}
