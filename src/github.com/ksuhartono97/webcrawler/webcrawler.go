package webcrawler

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type UrlData struct {
	sourceUrl string
	foundUrl  []string
	pageTitle string
	pageSize  int
	rawHTML   string
	lastModified string
}

var exploredPages = 0

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

// Extract all required info from a given webpage
func crawl(src string, ch chan UrlData, chFinished chan bool) {
	//Retrieve the webpage
	resp, err := http.Get(src)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + src + "\"")
		return
	}

	urlResult := UrlData{sourceUrl: src}

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

	response, err := http.Head(src)
	if err != nil {
		fmt.Println("Error while downloading head of", src)
	} else {
		for k, v := range response.Header {
			if k == "Last-Modified" {
				urlResult.lastModified = v[0]
			}
		}
		if urlResult.lastModified == "" {
			ti := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
			urlResult.lastModified = ti
		}
	}

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done, increment explored pages and return result
			ch <- urlResult
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

				// Make sure the url begins in http**
				hasProto := strings.Index(url, "http") == 0
				if hasProto {
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

//Main search function
func PrintLinks(links ...string) {
	foundUrls := make(map[string]UrlData)
	seedUrls := links

	// Channels
	chUrls := make(chan UrlData)
	chFinished := make(chan bool)

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url.sourceUrl] = url
		case <-chFinished:
			c++
		}
	}

	fmt.Println("\n\nTotal explored ", exploredPages)

	for _, url := range foundUrls {

		//Printing the results
		fmt.Println("\nFound", len(url.foundUrl), "non unique urls:\n")
		// for i := 0; i < len(url.foundUrl); i++ {
		// 	fmt.Println(" > " + url.foundUrl[i])
		// }
		fmt.Println("Page Title: " + url.pageTitle)
		fmt.Println("Page Size: ", url.pageSize)
		fmt.Println("Last Modified: " + url.lastModified)

		// Calculate remaining URLs needed
		diff := 30 - exploredPages
		remaining := diff - len(url.foundUrl)
		var toBeCalled = 0
		if remaining < 0 {
			toBeCalled = len(url.foundUrl) + remaining
		} else {
			toBeCalled = len(url.foundUrl)
		}

		urlArray := url.foundUrl[:toBeCalled]

		if toBeCalled > 0 {
			PrintLinks(urlArray...)
		}
	}

	close(chUrls)
}

//To be called before each initial search
func CrawlerInit() {
	exploredPages = 0
}
