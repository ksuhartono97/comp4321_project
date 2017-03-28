package webcrawler

import (
    "fmt"
    "net/http"
    "golang.org/x/net/html"
	  "strings"
)

type UrlData struct {
	sourceUrl string
  foundUrl []string
}


var exploredPages = 0;

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

// Extract all http** links from a given webpage
func crawl(src string, ch chan UrlData, chFinished chan bool) {
	resp, err := http.Get(src)

  urlResult := UrlData{sourceUrl: src}

	defer func() {
		// Notify that we're done after this function
    ch<-urlResult
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + src + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// Make sure the url begins in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				// ch <-UrlData{src, url}
        urlResult.foundUrl = append(urlResult.foundUrl, url)
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
      exploredPages++
		}
	}

	//Printing the results

	for _, url := range foundUrls {
    fmt.Println("\nFound", len(url.foundUrl), "non unique urls:\n")
    for i:= 0; i < len(url.foundUrl); i++ {
      fmt.Println(" - " + url.foundUrl[i])
    }
      fmt.Println("Total explored ", exploredPages)

      //Calculate remaining URLs needed
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
func CrawlerInit () {
  exploredPages = 0
}
