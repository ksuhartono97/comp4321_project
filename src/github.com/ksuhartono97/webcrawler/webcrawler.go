package webcrawler

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
  "io/ioutil"
)

type UrlData struct {
	sourceUrl string
	foundUrl  []string
  pageTitle string
  pageSize int
  rawHTML string
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
func crawl(url string, ch chan UrlData, chFinished chan bool) {
  //Retrieve the webpage
	resp, err := http.Get(url)

  defer func() {
    // Notify that we're done after this function
    chFinished <- true
  }()

  if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

  urlResult := UrlData{sourceUrl: url}

	b := resp.Body
	defer b.Close() // close Body when the function returns

  //Open a secondary stream

  res, err := http.Get(url)

  c := res.Body
  defer c.Close()

  htmlData, err := ioutil.ReadAll(c)

  if err != nil {
    fmt.Println(err)
    return
  }

  // Get page size in bytes
  urlResult.pageSize = len(htmlData)
  // Get raw HTML
  urlResult.rawHTML = string(htmlData)

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

			// Check if the token tags
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
          t := z.Next()
          if t == html.TextToken {
            u := z.Token()
            urlResult.pageTitle += u.Data
          } else if t == html.EndTagToken {
            u := z.Token()
            if u.Data == "title" {
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

  fmt.Println(foundUrls)

  fmt.Println("\n\nTotal explored ", exploredPages)

	for _, url := range foundUrls {

    //Printing the results
		fmt.Println("\nFound", len(url.foundUrl), "non unique urls:\n")
		for i := 0; i < len(url.foundUrl); i++ {
			fmt.Println(" > " + url.foundUrl[i])
		}
    fmt.Println("Page Title: " + url.pageTitle)
    fmt.Println("Page Size: ", url.pageSize)

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
