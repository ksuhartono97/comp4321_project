package indexer

import (
	"fmt"
	"io"
	"time"

	"../../silver-rush/database"
	"golang.org/x/net/html"
)

//Feed a page to the indexer
func Feed(docID uint64, reader io.ReadCloser, lastModify time.Time, size int, parent uint64, child []uint64) {
	//Map of words and term frequency.
	//TODO: Change string to word ID
	wordMap := make(map[uint64]uint32)

	doc, err := html.Parse(reader)
	if err != nil {
		fmt.Println("Error in parsing")
	}

	//Map is pass by reference, so we're cool
	iterateNode(doc, wordMap)

}

func tokenize(text *string) []string {
	*text = html.UnescapeString(*text)
	var tokens []string
	head := 0

	//If I obtain the i for range, some indexes are skipped. No idea why.
	i := 0
	for _, c := range *text {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			if i == head {
				head++
			} else {
				tokens = append(tokens, (*text)[head:i])
				head = i + 1
			}
		}
		i++
	}

	if head != i {
		tokens = append(tokens, (*text)[head:i])
	}

	return tokens
}

func iterateNode(node *html.Node, wordMap map[uint64]uint32) {
	if node.Type == html.TextNode && node.Parent.Data != "script" {
		list := tokenize(&(node.Data))
		for _, s := range list {
			id, _ := database.GetIDWithWord(s)
			wordMap[id]++
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		iterateNode(child, wordMap)
	}
}
