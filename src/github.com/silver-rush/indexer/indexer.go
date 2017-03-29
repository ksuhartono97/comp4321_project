package indexer

import (
	"fmt"
	"sync"

	"strings"

	"../../silver-rush/database"
	"golang.org/x/net/html"
)

//Feed a page to the indexer
func Feed(docID uint64, raw string, lastModify uint32, size uint32, parent uint64, child []uint64, title string) {
	//Map of words and term frequency.
	wordMap := make(map[uint64]uint32)

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		fmt.Println("Error in parsing")
	}

	//Map is pass by reference, so we're cool
	iterateNode(doc, wordMap)

	var wg sync.WaitGroup
	wg.Add(len(wordMap) + 1)
	//Start goroutines to add words to the posting list
	for wordID, tf := range wordMap {
		go func(wordID uint64, tf uint32) {
			defer wg.Done()
			p := database.Posting{tf}
			database.InsertIntoPostingList(wordID, docID, &p)
		}(wordID, tf)
	}

	go func() {
		defer wg.Done()
		var d database.DocInfo
		d.Size = size
		d.Time = lastModify
		d.ParentID = parent
		d.ChildNum = uint32(len(child))
		d.Child = child
		d.Title = title
		database.InsertDocInfo(docID, &d)
	}()

	wg.Wait()
	fmt.Println("Done. ", docID)
}

func tokenize(text *string) []string {
	*text = html.UnescapeString(*text)
	var tokens []string
	head := 0

	//If I obtain the i for range, some indexes are skipped. No idea why.
	i := 0
	for _, c := range *text {
		//Index english alphahets only
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			if i == head {
				head++
			} else {
				//Append a slice
				tokens = append(tokens, (*text)[head:i])
				head = i + 1
			}
		}
		i++
	}

	//Deal with the last word
	if head != i {
		tokens = append(tokens, (*text)[head:i])
	}

	return tokens
}

func iterateNode(node *html.Node, wordMap map[uint64]uint32) {
	if node.Type == html.TextNode && node.Parent.Data != "script" {
		list := tokenize(&(node.Data))
		for _, s := range list {
			//Collect word id usin the word
			id, _ := database.GetIDWithWord(s)
			wordMap[id]++
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		//Recursively iterate through all nodes
		iterateNode(child, wordMap)
	}
}
