package webcrawler

import (
    "fmt"
    "net/http"
    "golang.org/x/net/html"
	  "strings"
)

type UrlData struct {
	sourceUrl string
  foundUrl string
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

	defer func() {
		// Notify that we're done after this function
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
				ch <-UrlData{src, url}
			}
		}
	}
}

//Main search function
func PrintLinks(links ...string) {
	foundUrls := make(map[UrlData]bool)
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
			foundUrls[url] = true
		case <-chFinished:
			c++
      exploredPages++
		}
	}

	//Printing the results

	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url.foundUrl + " ||| " + url.sourceUrl)
	}
  fmt.Println("Total explored ", exploredPages)

  //Calculate remaining URLs needed
  diff := 30 - exploredPages
  remaining := diff - len(foundUrls)
  var toBeCalled = 0
  if remaining < 0 {
    toBeCalled = len(foundUrls) + remaining
  } else {
    toBeCalled = len(foundUrls)
  }
  //Filter out the urls from the bool in the map
  urlArray := make([]string, toBeCalled)
  idx := 0
  for url, _ := range foundUrls {
    if idx >= toBeCalled {break}
     urlArray[idx] = url.foundUrl
     idx++
  }

	close(chUrls)

  if toBeCalled > 0 {
    PrintLinks(urlArray...)
  }
}

//To be called before each initial search
func CrawlerInit () {
  exploredPages = 0
}
