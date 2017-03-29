package indexer

import (
	"fmt"
	"io"
	"time"

	"strings"

	"unicode"

	"golang.org/x/net/html"
)

//Feed a page to the indexer
func Feed(docID int, reader io.ReadCloser, lastModify time.Time, size int, parent int, child []int) {
	//Map of words and term frequency.
	//TODO: Change string to word ID
	wordMap := make(map[string]int)

	doc, err := html.Parse(reader)
	if err != nil {
		fmt.Println("Error in parsing")
	}

	//Map is pass by reference, so we're cool
	iterateNode(doc, wordMap)
}

var replacer = strings.NewReplacer("\n", " ", "\t", " ")

func tokenize(text *string) []string {
	*text = html.UnescapeString(*text)
	var tokens []string
	head := 0

	//If I obtain the i for range, some indexes are skipped. No idea why.
	i := 0
	for _, c := range *text {
		if !unicode.IsLetter(c) {
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

func iterateNode(node *html.Node, wordMap map[string]int) {
	if node.Type == html.TextNode && node.Parent.Data != "script" {
		list := tokenize(&(node.Data))
		for _, s := range list {
			fmt.Printf("%s\n", s)
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		iterateNode(child, wordMap)
	}
}
